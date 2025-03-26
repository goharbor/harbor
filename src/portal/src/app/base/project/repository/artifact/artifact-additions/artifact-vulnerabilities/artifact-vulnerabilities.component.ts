// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
} from '../../../../../../shared/units/utils';
import { ResultBarChartComponent } from '../../vulnerability-scanning/result-bar-chart.component';
import { Subscription } from 'rxjs';
import { Artifact } from '../../../../../../../../ng-swagger-gen/models/artifact';
import { SessionService } from '../../../../../../shared/services/session.service';
import {
    EventService,
    HarborEvent,
} from '../../../../../../services/event-service/event.service';
import { severityText } from '../../../../../left-side-nav/interrogation-services/vulnerability-database/security-hub.interface';

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
    @Input() scanBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    @Input() hasEnabledScanner: boolean = false;
    scan_overview: any;
    scanner: ScannerVo;

    scanningResults: VulnerabilityItem[] = [];
    loading: boolean = false;
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
    readonly severityText = severityText;
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

    getLevel(v: VulnerabilityItem): number {
        if (v && v.severity && SEVERITY_LEVEL_MAP[v.severity]) {
            return SEVERITY_LEVEL_MAP[v.severity];
        }
        return 0;
    }

    refresh(): void {
        this.getVulnerabilities();
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
