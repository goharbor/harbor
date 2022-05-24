import { Component, Input, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { AdditionsService } from '../additions.service';
import {
    ClrDatagridComparatorInterface,
    ClrDatagridStateInterface,
    ClrLoadingState,
} from '@clr/angular';
import { finalize } from 'rxjs/operators';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import {
    ProjectService,
    ScannerVo,
    ScanningResultService,
    SystemInfoService,
    UserPermissionService,
    USERSTATICPERMISSION,
    VulnerabilityItem,
} from '../../../../../../shared/services';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import {
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
    SEVERITY_LEVEL_MAP,
    VULNERABILITY_SEVERITY,
} from '../../../../../../shared/units/utils';
import { ResultBarChartComponent } from '../../vulnerability-scanning/result-bar-chart.component';
import { Subscription } from 'rxjs';
import { Artifact } from '../../../../../../../../ng-swagger-gen/models/artifact';
import { SessionService } from '../../../../../../shared/services/session.service';
import {
    EventService,
    HarborEvent,
} from '../../../../../../services/event-service/event.service';

@Component({
    selector: 'hbr-artifact-vulnerabilities',
    templateUrl: './artifact-vulnerabilities.component.html',
    styleUrls: ['./artifact-vulnerabilities.component.scss'],
})
export class ArtifactVulnerabilitiesComponent implements OnInit, OnDestroy {
    @Input()
    vulnerabilitiesLink: AdditionLink;
    @Input()
    projectName: string;
    @Input()
    projectId: number;
    @Input()
    repoName: string;
    @Input()
    digest: string;
    @Input() artifact: Artifact;
    scan_overview: any;
    scanner: ScannerVo;
    projectScanner: ScannerVo;

    scanningResults: VulnerabilityItem[] = [];
    loading: boolean = false;
    hasEnabledScanner: boolean = false;
    scanBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    severitySort: ClrDatagridComparatorInterface<VulnerabilityItem>;
    cvssSort: ClrDatagridComparatorInterface<VulnerabilityItem>;
    hasScanningPermission: boolean = false;
    onSendingScanCommand: boolean = false;
    onSendingStopCommand: boolean = false;
    hasShowLoading: boolean = false;
    @ViewChild(ResultBarChartComponent)
    resultBarChartComponent: ResultBarChartComponent;
    sub: Subscription;
    hasViewInitWithDelay: boolean = false;
    currentCVEList: Array<{ cve_id: string }> = [];
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.ARTIFACT_VUL_COMPONENT,
        25
    );
    constructor(
        private errorHandler: ErrorHandler,
        private additionsService: AdditionsService,
        private userPermissionService: UserPermissionService,
        private scanningService: ScanningResultService,
        private eventService: EventService,
        private session: SessionService,
        private projectService: ProjectService,
        private systemInfoService: SystemInfoService
    ) {
        const that = this;
        this.severitySort = {
            compare(a: VulnerabilityItem, b: VulnerabilityItem): number {
                return that.getLevel(a) - that.getLevel(b);
            },
        };
        this.cvssSort = {
            compare(a: VulnerabilityItem, b: VulnerabilityItem): number {
                if (
                    a &&
                    a.preferred_cvss &&
                    a.preferred_cvss.score_v3 &&
                    b &&
                    b.preferred_cvss &&
                    b.preferred_cvss.score_v3
                ) {
                    return (
                        +a.preferred_cvss.score_v3 - +b.preferred_cvss.score_v3
                    );
                }
                return 0;
            },
        };
    }

    ngOnInit() {
        this.getVulnerabilities();
        this.getScanningPermission();
        this.getProjectScanner();
        if (!this.sub) {
            this.sub = this.eventService.subscribe(
                HarborEvent.UPDATE_VULNERABILITY_INFO,
                (artifact: Artifact) => {
                    if (artifact?.digest === this.artifact?.digest) {
                        this.getVulnerabilities();
                    }
                }
            );
        }
        setTimeout(() => {
            this.hasViewInitWithDelay = true;
        }, 0);
        // get system and project CVE allow list
        this.getCurrentCVEAllowList();
    }

    ngOnDestroy() {
        if (this.sub) {
            this.sub.unsubscribe();
            this.sub = null;
        }
    }

    getVulnerabilities() {
        if (
            this.vulnerabilitiesLink &&
            !this.vulnerabilitiesLink.absolute &&
            this.vulnerabilitiesLink.href
        ) {
            if (!this.hasShowLoading) {
                this.loading = true;
                this.hasShowLoading = true;
            }
            this.additionsService
                .getDetailByLink(this.vulnerabilitiesLink.href, true, false)
                .pipe(
                    finalize(() => {
                        this.loading = false;
                        this.hasShowLoading = false;
                    })
                )
                .subscribe(
                    res => {
                        this.scan_overview = res;
                        if (
                            this.scan_overview &&
                            Object.values(this.scan_overview)[0]
                        ) {
                            this.scanningResults =
                                (Object.values(this.scan_overview)[0] as any)
                                    .vulnerabilities || [];
                            // sort
                            if (this.scanningResults) {
                                this.scanningResults.sort(
                                    (a, b) =>
                                        this.getLevel(b) - this.getLevel(a)
                                );
                            }
                            this.scanner = (
                                Object.values(this.scan_overview)[0] as any
                            ).scanner;
                        }
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }

    getScanningPermission(): void {
        const permissions = [
            {
                resource: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.KEY,
                action: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.VALUE
                    .CREATE,
            },
        ];
        this.userPermissionService
            .hasProjectPermissions(this.projectId, permissions)
            .subscribe(
                (results: Array<boolean>) => {
                    this.hasScanningPermission = results[0];
                    // only has label permission
                },
                error => this.errorHandler.error(error)
            );
    }

    getProjectScanner(): void {
        this.hasEnabledScanner = false;
        this.scanBtnState = ClrLoadingState.LOADING;
        this.scanningService.getProjectScanner(this.projectId).subscribe(
            response => {
                if (
                    response &&
                    '{}' !== JSON.stringify(response) &&
                    !response.disabled &&
                    response.health === 'healthy'
                ) {
                    this.scanBtnState = ClrLoadingState.SUCCESS;
                    this.hasEnabledScanner = true;
                } else {
                    this.scanBtnState = ClrLoadingState.ERROR;
                }
                this.projectScanner = response;
            },
            error => {
                this.scanBtnState = ClrLoadingState.ERROR;
            }
        );
    }

    getLevel(v: VulnerabilityItem): number {
        if (v && v.severity && SEVERITY_LEVEL_MAP[v.severity]) {
            return SEVERITY_LEVEL_MAP[v.severity];
        }
        return 0;
    }

    refresh(): void {
        this.getVulnerabilities();
    }

    severityText(severity: string): string {
        switch (severity) {
            case VULNERABILITY_SEVERITY.CRITICAL:
                return 'VULNERABILITY.SEVERITY.CRITICAL';
            case VULNERABILITY_SEVERITY.HIGH:
                return 'VULNERABILITY.SEVERITY.HIGH';
            case VULNERABILITY_SEVERITY.MEDIUM:
                return 'VULNERABILITY.SEVERITY.MEDIUM';
            case VULNERABILITY_SEVERITY.LOW:
                return 'VULNERABILITY.SEVERITY.LOW';
            case VULNERABILITY_SEVERITY.NONE:
                return 'VULNERABILITY.SEVERITY.NONE';
            default:
                return 'UNKNOWN';
        }
    }

    scanNow() {
        this.onSendingScanCommand = true;
        this.eventService.publish(
            HarborEvent.START_SCAN_ARTIFACT,
            this.repoName + '/' + this.digest
        );
    }

    submitFinish(e: boolean) {
        this.onSendingScanCommand = e;
    }

    submitStopFinish(e: boolean) {
        this.onSendingStopCommand = e;
    }

    shouldShowBar(): boolean {
        return (
            this.hasViewInitWithDelay &&
            this.resultBarChartComponent &&
            (this.resultBarChartComponent.queued ||
                this.resultBarChartComponent.scanning ||
                this.resultBarChartComponent.error ||
                this.resultBarChartComponent.stopped)
        );
    }

    hasScanned(): boolean {
        return (
            this.hasViewInitWithDelay &&
            this.resultBarChartComponent &&
            !(
                this.resultBarChartComponent.completed ||
                this.resultBarChartComponent.error ||
                this.resultBarChartComponent.queued ||
                this.resultBarChartComponent.stopped ||
                this.resultBarChartComponent.scanning
            )
        );
    }

    handleScanOverview(scanOverview: any): any {
        if (scanOverview) {
            return Object.values(scanOverview)[0];
        }
        return null;
    }

    isSystemAdmin(): boolean {
        const account = this.session.getCurrentUser();
        return account && account.has_admin_role;
    }

    getCurrentCVEAllowList() {
        this.projectService.getProject(this.projectId).subscribe(projectRes => {
            if (
                projectRes &&
                projectRes.cve_allowlist &&
                projectRes.metadata &&
                projectRes.metadata.reuse_sys_cve_allowlist === 'false'
            ) {
                // use project CVE allow list
                this.currentCVEList = projectRes.cve_allowlist['items'];
            } else {
                // use system CVE allow list
                this.systemInfoService
                    .getSystemAllowlist()
                    .subscribe(systemRes => {
                        if (
                            systemRes &&
                            systemRes.items &&
                            systemRes.items.length
                        ) {
                            this.currentCVEList = systemRes.items;
                        }
                    });
            }
        });
    }

    isInAllowList(CVEId: string): boolean {
        if (this.currentCVEList && this.currentCVEList.length) {
            for (let i = 0; i < this.currentCVEList.length; i++) {
                if (CVEId === this.currentCVEList[i].cve_id) {
                    return true;
                }
            }
        }
        return false;
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
            this.resultBarChartComponent &&
            (this.resultBarChartComponent.queued ||
                this.resultBarChartComponent.scanning)
        );
    }

    scanOrStop() {
        if (this.isRunningState()) {
            this.stopNow();
        } else {
            this.scanNow();
        }
    }

    stopNow() {
        this.onSendingStopCommand = true;
        this.eventService.publish(
            HarborEvent.STOP_SCAN_ARTIFACT,
            this.repoName + '/' + this.digest
        );
    }
    canScan(): boolean {
        return (
            this.hasEnabledScanner &&
            this.hasScanningPermission &&
            !this.onSendingScanCommand
        );
    }
    load(state: ClrDatagridStateInterface) {
        if (state?.page?.size) {
            setPageSizeToLocalStorage(
                PageSizeMapKeys.ARTIFACT_VUL_COMPONENT,
                state.page.size
            );
        }
    }
}
