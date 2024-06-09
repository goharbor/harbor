import { Component, Input, OnDestroy, OnInit } from '@angular/core';
import { ClrDatagridStateInterface, ClrLoadingState } from '@clr/angular';
import { finalize } from 'rxjs/operators';
import {
    ScannerVo,
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../../../../shared/services';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import {
    dbEncodeURIComponent,
    downloadJson,
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    SBOM_SCAN_STATUS,
    setPageSizeToLocalStorage,
} from '../../../../../../shared/units/utils';
import { Subscription } from 'rxjs';
import { Artifact } from '../../../../../../../../ng-swagger-gen/models/artifact';
import { SessionService } from '../../../../../../shared/services/session.service';
import {
    EventService,
    HarborEvent,
} from '../../../../../../services/event-service/event.service';
import { severityText } from '../../../../../left-side-nav/interrogation-services/vulnerability-database/security-hub.interface';

import {
    ArtifactSbom,
    ArtifactSbomPackageItem,
    getArtifactSbom,
} from '../../artifact';
import { ArtifactService } from 'ng-swagger-gen/services';
import { ScanTypes } from '../../../../../../shared/entities/shared.const';

@Component({
    selector: 'hbr-artifact-sbom',
    templateUrl: './artifact-sbom.component.html',
    styleUrls: ['./artifact-sbom.component.scss'],
})
export class ArtifactSbomComponent implements OnInit, OnDestroy {
    @Input()
    projectName: string;
    @Input()
    projectId: number;
    @Input()
    repoName: string;
    @Input()
    sbomDigest: string;
    @Input() artifact: Artifact;
    @Input() hasScannerSupportSBOM: boolean = false;

    artifactSbom: ArtifactSbom;
    loading: boolean = false;
    downloadSbomBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    hasSbomPermission: boolean = false;
    hasShowLoading: boolean = false;
    sub: Subscription;
    hasViewInitWithDelay: boolean = false;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.ARTIFACT_SBOM_COMPONENT,
        25
    );
    readonly severityText = severityText;
    constructor(
        private errorHandler: ErrorHandler,
        private artifactService: ArtifactService,
        private userPermissionService: UserPermissionService,
        private eventService: EventService,
        private session: SessionService
    ) {}

    ngOnInit() {
        this.getSbom();
        this.getSbomPermission();
        if (!this.sub) {
            this.sub = this.eventService.subscribe(
                HarborEvent.UPDATE_SBOM_INFO,
                (artifact: Artifact) => {
                    if (artifact?.digest === this.artifact?.digest) {
                        if (artifact.sbom_overview) {
                            const sbomDigest = Object.values(
                                artifact.sbom_overview
                            )?.[0]?.sbom_digest;
                            if (sbomDigest) {
                                this.sbomDigest = sbomDigest;
                            }
                        }
                        this.getSbom();
                    }
                }
            );
        }
        setTimeout(() => {
            this.hasViewInitWithDelay = true;
        }, 0);
    }

    ngOnDestroy() {
        if (this.sub) {
            this.sub.unsubscribe();
            this.sub = null;
        }
    }

    getSbom() {
        if (this.sbomDigest) {
            if (!this.hasShowLoading) {
                this.loading = true;
                this.hasShowLoading = true;
            }
            const sbomAdditionParams = <ArtifactService.GetAdditionParams>{
                repositoryName: dbEncodeURIComponent(this.repoName),
                reference: this.sbomDigest,
                projectName: this.projectName,
                addition: ScanTypes.SBOM,
            };
            this.artifactService
                .getAddition(sbomAdditionParams)
                .pipe(
                    finalize(() => {
                        this.loading = false;
                        this.hasShowLoading = false;
                    })
                )
                .subscribe(
                    res => {
                        if (res) {
                            this.artifactSbom = getArtifactSbom(
                                JSON.parse(res)
                            );
                        } else {
                            this.loading = false;
                            this.hasShowLoading = false;
                        }
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }

    getSbomPermission(): void {
        const permissions = [
            {
                resource: USERSTATICPERMISSION.REPOSITORY_TAG_SBOM_JOB.KEY,
                action: USERSTATICPERMISSION.REPOSITORY_TAG_SBOM_JOB.VALUE.READ,
            },
        ];
        this.userPermissionService
            .hasProjectPermissions(this.projectId, permissions)
            .subscribe(
                (results: Array<boolean>) => {
                    this.hasSbomPermission = results[0];
                    // only has label permission
                },
                error => this.errorHandler.error(error)
            );
    }

    refresh(): void {
        this.getSbom();
    }

    hasGeneratedSbom(): boolean {
        return this.hasViewInitWithDelay;
    }

    isSystemAdmin(): boolean {
        const account = this.session.getCurrentUser();
        return account && account.has_admin_role;
    }

    getScannerInfo(scanner: ScannerVo): string {
        if (scanner) {
            if (scanner.name && scanner.version) {
                return `${scanner.name}@${scanner.version}`;
            }
            if (scanner.name && !scanner.version) {
                return `${scanner.name}`;
            }
        }
        return '';
    }

    isRunningState(): boolean {
        return (
            this.hasViewInitWithDelay &&
            this.artifact.sbom_overview &&
            (this.artifact.sbom_overview.scan_status ===
                SBOM_SCAN_STATUS.PENDING ||
                this.artifact.sbom_overview.scan_status ===
                    SBOM_SCAN_STATUS.RUNNING)
        );
    }

    downloadSbom() {
        this.downloadSbomBtnState = ClrLoadingState.LOADING;
        if (
            this.artifact?.sbom_overview?.scan_status ===
                SBOM_SCAN_STATUS.SUCCESS ||
            !!this.sbomDigest ||
            (this.artifactSbom.sbomJsonRaw && this.artifactSbom.sbomName)
        ) {
            downloadJson(
                this.artifactSbom.sbomJsonRaw,
                `${this.artifactSbom.sbomName}.json`
            );
        }
        this.downloadSbomBtnState = ClrLoadingState.DEFAULT;
    }

    canDownloadSbom(): boolean {
        return (
            this.hasScannerSupportSBOM &&
            //this.hasSbomPermission &&
            this.sbomDigest &&
            this.downloadSbomBtnState !== ClrLoadingState.LOADING &&
            this.artifactSbom !== undefined
        );
    }

    artifactSbomPackages(): ArtifactSbomPackageItem[] {
        return (
            this.artifactSbom?.sbomPackage?.packages?.filter(
                item =>
                    item?.name || item?.versionInfo || item?.licenseConcluded
            ) ?? []
        );
    }

    load(state: ClrDatagridStateInterface) {
        if (state?.page?.size) {
            setPageSizeToLocalStorage(
                PageSizeMapKeys.ARTIFACT_SBOM_COMPONENT,
                state.page.size
            );
        }
    }
}
