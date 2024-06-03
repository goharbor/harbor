import {
    Component,
    EventEmitter,
    Input,
    OnDestroy,
    OnInit,
    Output,
} from '@angular/core';
import { Subscription, timer } from 'rxjs';
import { finalize } from 'rxjs/operators';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import {
    clone,
    CURRENT_BASE_HREF,
    dbEncodeURIComponent,
    DEFAULT_SUPPORTED_MIME_TYPES,
    SBOM_SCAN_STATUS,
} from '../../../../../shared/units/utils';
import { ArtifactService } from '../../../../../../../ng-swagger-gen/services/artifact.service';
import { Artifact } from '../../../../../../../ng-swagger-gen/models/artifact';
import {
    EventService,
    HarborEvent,
} from '../../../../../services/event-service/event.service';
import { ScanService } from '../../../../../../../ng-swagger-gen/services/scan.service';
import { Accessory, ScanType } from 'ng-swagger-gen/models';
import { ScanTypes } from '../../../../../shared/entities/shared.const';
import { SBOMOverview } from './sbom-overview';
import { Scanner } from '../../../../left-side-nav/interrogation-services/scanner/scanner';
const STATE_CHECK_INTERVAL: number = 3000; // 3s
const RETRY_TIMES: number = 3;

@Component({
    selector: 'hbr-sbom-bar',
    templateUrl: './sbom-scan-component.html',
    styleUrls: ['./scanning.scss'],
})
export class ResultSbomComponent implements OnInit, OnDestroy {
    @Input() inputScanner: Scanner;
    @Input() repoName: string = '';
    @Input() projectName: string = '';
    @Input() projectId: string = '';
    @Input() artifactDigest: string = '';
    @Input() sbomDigest: string = '';
    @Input() sbomOverview: SBOMOverview;
    @Input() accessories: Accessory[] = [];
    onSubmitting: boolean = false;
    onStopping: boolean = false;
    retryCounter: number = 0;
    stateCheckTimer: Subscription;
    generateSbomSubscription: Subscription;
    stopSubscription: Subscription;
    timerHandler: any;
    @Output()
    submitFinish: EventEmitter<boolean> = new EventEmitter<boolean>();
    // if sending stop scan request is finished, emit to farther component
    @Output()
    submitStopFinish: EventEmitter<boolean> = new EventEmitter<boolean>();
    @Output()
    scanFinished: EventEmitter<Artifact> = new EventEmitter<Artifact>();

    constructor(
        private artifactService: ArtifactService,
        private scanService: ScanService,
        private errorHandler: ErrorHandler,
        private eventService: EventService
    ) {}

    ngOnInit(): void {
        if (
            (this.status === SBOM_SCAN_STATUS.RUNNING ||
                this.status === SBOM_SCAN_STATUS.PENDING) &&
            !this.stateCheckTimer
        ) {
            // Avoid duplicated subscribing
            this.stateCheckTimer = timer(0, STATE_CHECK_INTERVAL).subscribe(
                () => {
                    this.getSbomOverview();
                }
            );
        }
        if (!this.generateSbomSubscription) {
            this.generateSbomSubscription = this.eventService.subscribe(
                HarborEvent.START_GENERATE_SBOM,
                (artifactDigest: string) => {
                    let myFullTag: string =
                        this.repoName + '/' + this.artifactDigest;
                    if (myFullTag === artifactDigest) {
                        this.generateSbom();
                    }
                }
            );
        }
        if (!this.stopSubscription) {
            this.stopSubscription = this.eventService.subscribe(
                HarborEvent.STOP_SBOM_ARTIFACT,
                (artifactDigest: string) => {
                    let myFullTag: string =
                        this.repoName + '/' + this.artifactDigest;
                    if (myFullTag === artifactDigest) {
                        this.stopSbom();
                    }
                }
            );
        }
    }

    ngOnDestroy(): void {
        if (this.stateCheckTimer) {
            this.stateCheckTimer.unsubscribe();
            this.stateCheckTimer = null;
        }
        if (this.generateSbomSubscription) {
            this.generateSbomSubscription.unsubscribe();
            this.generateSbomSubscription = null;
        }
        if (this.stopSubscription) {
            this.stopSubscription.unsubscribe();
            this.stopSubscription = null;
        }
    }

    // Get vulnerability scanning status
    public get status(): string {
        if (this.sbomOverview && this.sbomOverview.scan_status) {
            return this.sbomOverview.scan_status;
        }
        return SBOM_SCAN_STATUS.NOT_GENERATED_SBOM;
    }

    public get completed(): boolean {
        return !!this.sbomOverview && this.status !== SBOM_SCAN_STATUS.SUCCESS
            ? false
            : this.status === SBOM_SCAN_STATUS.SUCCESS || !!this.sbomDigest;
    }

    public get error(): boolean {
        return this.status === SBOM_SCAN_STATUS.ERROR;
    }

    public get queued(): boolean {
        return this.status === SBOM_SCAN_STATUS.PENDING;
    }

    public get generating(): boolean {
        return this.status === SBOM_SCAN_STATUS.RUNNING;
    }

    public get stopped(): boolean {
        return this.status === SBOM_SCAN_STATUS.STOPPED;
    }

    public get otherStatus(): boolean {
        return !(
            this.completed ||
            this.error ||
            this.queued ||
            this.generating ||
            this.stopped
        );
    }

    generateSbom(): void {
        if (this.onSubmitting) {
            // Avoid duplicated submitting
            console.error('duplicated submit');
            return;
        }

        if (!this.repoName || !this.artifactDigest) {
            console.error('bad repository or tag');
            return;
        }

        this.onSubmitting = true;

        this.scanService
            .scanArtifact({
                projectName: this.projectName,
                reference: this.artifactDigest,
                repositoryName: dbEncodeURIComponent(this.repoName),
                scanType: <ScanType>{
                    scan_type: ScanTypes.SBOM,
                },
            })
            .pipe(finalize(() => this.submitFinish.emit(false)))
            .subscribe(
                () => {
                    this.onSubmitting = false;
                    // Forcely change status to queued after successful submitting
                    this.sbomOverview = {
                        scan_status: SBOM_SCAN_STATUS.PENDING,
                    };
                    // Start check status util the job is done
                    if (!this.stateCheckTimer) {
                        // Avoid duplicated subscribing
                        this.stateCheckTimer = timer(
                            STATE_CHECK_INTERVAL,
                            STATE_CHECK_INTERVAL
                        ).subscribe(() => {
                            this.getSbomOverview();
                        });
                    }
                },
                error => {
                    this.onSubmitting = false;
                    if (error && error.error && error.error.code === 409) {
                        console.error(error.error.message);
                    } else {
                        this.errorHandler.error(error);
                    }
                }
            );
    }

    getSbomOverview(): void {
        if (!this.repoName || !this.artifactDigest) {
            return;
        }
        this.artifactService
            .getArtifact({
                projectName: this.projectName,
                repositoryName: dbEncodeURIComponent(this.repoName),
                reference: this.artifactDigest,
                withSbomOverview: true,
                withAccessory: true,
                XAcceptVulnerabilities: DEFAULT_SUPPORTED_MIME_TYPES,
            })
            .subscribe(
                (artifact: Artifact) => {
                    // To keep the same summary reference, use value copy.
                    if (artifact.sbom_overview) {
                        this.copyValue(artifact.sbom_overview);
                    }
                    if (!this.queued && !this.generating) {
                        // Scanning should be done
                        if (this.stateCheckTimer) {
                            this.stateCheckTimer.unsubscribe();
                            this.stateCheckTimer = null;
                        }
                        this.scanFinished.emit(artifact);
                    }
                    this.eventService.publish(
                        HarborEvent.UPDATE_SBOM_INFO,
                        artifact
                    );
                },
                error => {
                    this.errorHandler.error(error);
                    this.retryCounter++;
                    if (this.retryCounter >= RETRY_TIMES) {
                        // Stop timer
                        if (this.stateCheckTimer) {
                            this.stateCheckTimer.unsubscribe();
                            this.stateCheckTimer = null;
                        }
                        this.retryCounter = 0;
                    }
                }
            );
    }

    copyValue(newVal: SBOMOverview): void {
        if (!this.sbomOverview || !newVal || !newVal.scan_status) {
            return;
        }
        this.sbomOverview = clone(newVal);
    }

    viewLog(): string {
        return `${CURRENT_BASE_HREF}/projects/${
            this.projectName
        }/repositories/${dbEncodeURIComponent(this.repoName)}/artifacts/${
            this.artifactDigest
        }/scan/${this.sbomOverview.report_id}/log`;
    }

    getScanner(): Scanner {
        return this.inputScanner;
    }

    stopSbom() {
        if (this.onStopping) {
            // Avoid duplicated stopping command
            console.error('duplicated stopping command for SBOM generation');
            return;
        }
        if (!this.repoName || !this.artifactDigest) {
            console.error('bad repository or artifact');
            return;
        }
        this.onStopping = true;

        this.scanService
            .stopScanArtifact({
                projectName: this.projectName,
                reference: this.artifactDigest,
                repositoryName: dbEncodeURIComponent(this.repoName),
                scanType: <ScanType>{
                    scan_type: ScanTypes.SBOM,
                },
            })
            .pipe(
                finalize(() => {
                    this.submitStopFinish.emit(false);
                    this.onStopping = false;
                })
            )
            .subscribe(
                () => {
                    // Start check status util the job is done
                    if (!this.stateCheckTimer) {
                        // Avoid duplicated subscribing
                        this.stateCheckTimer = timer(
                            STATE_CHECK_INTERVAL,
                            STATE_CHECK_INTERVAL
                        ).subscribe(() => {
                            this.getSbomOverview();
                        });
                    }
                    this.errorHandler.info('SBOM.TRIGGER_STOP_SUCCESS');
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }
}
