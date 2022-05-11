import {
    Component,
    EventEmitter,
    Input,
    OnInit,
    Output,
    ViewChild,
} from '@angular/core';
import { forkJoin, Observable, throwError as observableThrowError } from 'rxjs';
import { catchError, finalize, map } from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
import {
    HelmChartMaintainer,
    HelmChartVersion,
} from '../../helm-chart.interface.service';
import { HelmChartService } from '../../helm-chart.service';
import {
    LabelService as OldLabelService,
    State,
    SystemInfo,
    SystemInfoService,
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../../../../shared/services';
import {
    downloadFile,
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../../../shared/units/utils';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { OperationService } from '../../../../../../shared/components/operation/operation.service';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../../../shared/components/operation/operate';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
    DefaultHelmIcon,
    ResourceType,
} from '../../../../../../shared/entities/shared.const';
import { errorHandler } from '../../../../../../shared/units/shared.utils';
import { ConfirmationDialogComponent } from '../../../../../../shared/components/confirmation-dialog';
import { ConfirmationMessage } from '../../../../../global-confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../../../../../global-confirmation-dialog/confirmation-state-message';
import { Label } from '../../../../../../../../ng-swagger-gen/models/label';
import { LabelService } from '../../../../../../../../ng-swagger-gen/services/label.service';
import { ClrDatagridStateInterface } from '@clr/angular';

const PAGE_SIZE: number = 100;
@Component({
    selector: 'hbr-helm-chart-version',
    templateUrl: './helm-chart-version.component.html',
    styleUrls: ['./helm-chart-version.component.scss'],
})
export class ChartVersionComponent implements OnInit {
    signedCon: { [key: string]: any | string[] } = {};
    @Input() projectId: number;
    @Input() projectName: string;
    @Input() chartName: string;
    @Input() roleName: string;
    @Input() hasSignedIn: boolean;
    @Input() chartDefaultIcon: string = DefaultHelmIcon;
    @Output() versionClickEvt = new EventEmitter<string>();
    @Output() backEvt = new EventEmitter<any>();

    lastFilteredVersionName: string;
    chartVersions: HelmChartVersion[] = [];
    systemInfo: SystemInfo;
    selectedRows: HelmChartVersion[] = [];
    labels: Label[] = [];
    loading = true;
    resourceType = ResourceType.CHART_VERSION;

    isCardView: boolean;
    cardHover = false;
    listHover = false;

    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.CHART_VERSION_COMPONENT
    );
    currentPage = 1;
    totalCount = 0;
    currentState: State;

    chartFile: File;
    provFile: File;

    addLabelHeaders = 'HELM_CHART.ADD_LABEL_TO_CHART_VERSION';

    @ViewChild('confirmationDialog')
    confirmationDialog: ConfirmationDialogComponent;
    hasAddRemoveHelmChartVersionPermission: boolean;
    hasDownloadHelmChartVersionPermission: boolean;
    hasDeleteHelmChartVersionPermission: boolean;
    constructor(
        private errorHandlerEntity: ErrorHandler,
        private systemInfoService: SystemInfoService,
        private helmChartService: HelmChartService,
        private labelService: LabelService,
        private resrouceLabelService: OldLabelService,
        public userPermissionService: UserPermissionService,
        private operationService: OperationService,
        private translateService: TranslateService
    ) {}

    public get registryUrl(): string {
        return this.systemInfo ? this.systemInfo.registry_url : '';
    }

    ngOnInit(): void {
        // Get system info for tag views
        this.systemInfoService.getSystemInfo().subscribe(
            systemInfo => (this.systemInfo = systemInfo),
            error => this.errorHandlerEntity.error(error)
        );
        this.refresh();
        this.getLabels();
        this.lastFilteredVersionName = '';
        this.getHelmChartVersionPermission(this.projectId);
    }

    updateFilterValue(value: string) {
        this.lastFilteredVersionName = value;
        this.refresh();
    }

    getLabels() {
        // get all project labels
        this.labelService
            .ListLabelsResponse({
                pageSize: PAGE_SIZE,
                page: 1,
                scope: 'p',
                projectId: this.projectId,
            })
            .subscribe(res => {
                if (res.headers) {
                    const xHeader: string = res.headers.get('X-Total-Count');
                    const totalCount = parseInt(xHeader, 0);
                    let arr = res.body || [];
                    if (totalCount <= PAGE_SIZE) {
                        // already gotten all project labels
                        if (arr && arr.length) {
                            this.labels = this.labels.concat(arr);
                        }
                    } else {
                        // get all the project labels in specified times
                        const times: number = Math.ceil(totalCount / PAGE_SIZE);
                        const observableList: Observable<Label[]>[] = [];
                        for (let i = 2; i <= times; i++) {
                            observableList.push(
                                this.labelService.ListLabels({
                                    page: i,
                                    pageSize: PAGE_SIZE,
                                    scope: 'p',
                                    projectId: this.projectId,
                                })
                            );
                        }
                        forkJoin(observableList).subscribe(response => {
                            if (response && response.length) {
                                response.forEach(item => {
                                    arr = arr.concat(item);
                                });
                                this.labels = this.labels.concat(arr);
                            }
                        });
                    }
                }
            });
        // get all global labels
        this.labelService
            .ListLabelsResponse({
                pageSize: PAGE_SIZE,
                page: 1,
                scope: 'g',
            })
            .subscribe(res => {
                if (res.headers) {
                    const xHeader: string = res.headers.get('X-Total-Count');
                    const totalCount = parseInt(xHeader, 0);
                    let arr = res.body || [];
                    if (totalCount <= PAGE_SIZE) {
                        // already gotten all global labels
                        if (arr && arr.length) {
                            this.labels = this.labels.concat(arr);
                        }
                    } else {
                        // get all the global labels in specified times
                        const times: number = Math.ceil(totalCount / PAGE_SIZE);
                        const observableList: Observable<Label[]>[] = [];
                        for (let i = 2; i <= times; i++) {
                            observableList.push(
                                this.labelService.ListLabels({
                                    page: i,
                                    pageSize: PAGE_SIZE,
                                    scope: 'g',
                                })
                            );
                        }
                        forkJoin(observableList).subscribe(response => {
                            if (response && response.length) {
                                response.forEach(item => {
                                    arr = arr.concat(item);
                                });
                                this.labels = this.labels.concat(arr);
                            }
                        });
                    }
                }
            });
    }

    refresh(state?: ClrDatagridStateInterface) {
        if (state?.page?.size) {
            setPageSizeToLocalStorage(
                PageSizeMapKeys.CHART_VERSION_COMPONENT,
                state.page.size
            );
        }
        this.loading = true;
        this.helmChartService
            .getChartVersions(this.projectName, this.chartName)
            .pipe(
                finalize(() => {
                    this.selectedRows = [];
                    this.loading = false;
                })
            )
            .subscribe(
                versions => {
                    this.chartVersions = versions.filter(x =>
                        x?.version?.includes(this.lastFilteredVersionName)
                    );
                    this.totalCount = versions.length;
                },
                err => {
                    this.errorHandlerEntity.error(err);
                }
            );
    }

    getMaintainerString(maintainers: HelmChartMaintainer[]) {
        if (!maintainers || maintainers.length < 1) {
            return '';
        }

        let maintainer_string = maintainers[0].name;
        if (maintainers.length > 1) {
            maintainer_string = `${maintainer_string} (${
                maintainers.length - 1
            } others)`;
        }
        return maintainer_string;
    }

    onVersionClick(version: HelmChartVersion) {
        this.versionClickEvt.emit(version.version);
    }

    deleteVersion(version: HelmChartVersion): Observable<any> {
        // init operation info
        let operateMsg = new OperateInfo();
        operateMsg.name = 'OPERATION.DELETE_CHART_VERSION';
        operateMsg.data.id = version.digest;
        operateMsg.state = OperationState.progressing;
        operateMsg.data.name = `${version.name}:${version.version}`;
        this.operationService.publishInfo(operateMsg);

        return this.helmChartService
            .deleteChartVersion(
                this.projectName,
                this.chartName,
                version.version
            )
            .pipe(
                map(() => operateChanges(operateMsg, OperationState.success)),
                catchError(error => {
                    const message = errorHandler(error);
                    this.translateService
                        .get(message)
                        .subscribe(res =>
                            operateChanges(
                                operateMsg,
                                OperationState.failure,
                                res
                            )
                        );
                    return observableThrowError(error);
                })
            );
    }

    deleteVersions(versions: HelmChartVersion[]) {
        if (versions && versions.length < 1) {
            return;
        }
        let successCount: number;
        let totalCount = this.chartVersions.length;
        let versionObs = versions.map(v => this.deleteVersion(v));
        forkJoin(versionObs)
            .pipe(
                finalize(() => {
                    if (totalCount !== successCount) {
                        this.refresh();
                    }
                })
            )
            .subscribe(
                res => {
                    successCount = res.filter(
                        r => r.state === OperationState.success
                    ).length;
                    if (totalCount === successCount) {
                        this.backEvt.emit();
                    }
                },
                error => {
                    this.errorHandlerEntity.error(error);
                }
            );
    }

    versionDownload(evt?: Event, item?: HelmChartVersion) {
        if (evt) {
            evt.stopPropagation();
        }
        let selectedVersion: HelmChartVersion;

        if (item) {
            selectedVersion = item;
        } else {
            // return if selected version less then 1
            if (this.selectedRows.length < 1) {
                return;
            }
            selectedVersion = this.selectedRows[0];
        }
        if (!selectedVersion) {
            return;
        }

        let filename = selectedVersion.urls[0];
        this.helmChartService
            .downloadChart(this.projectName, filename)
            .subscribe(
                res => {
                    downloadFile(res);
                },
                error => {
                    this.errorHandlerEntity.error(error);
                }
            );
    }

    showCard(cardView: boolean) {
        if (this.isCardView === cardView) {
            return;
        }
        this.isCardView = cardView;
    }

    mouseEnter(itemName: string) {
        if (itemName === 'card') {
            this.cardHover = true;
        } else {
            this.listHover = true;
        }
    }

    mouseLeave(itemName: string) {
        if (itemName === 'card') {
            this.cardHover = false;
        } else {
            this.listHover = false;
        }
    }

    isHovering(itemName: string) {
        if (itemName === 'card') {
            return this.cardHover;
        } else {
            return this.listHover;
        }
    }

    onChartFileChangeEvent(event) {
        if (event.target.files && event.target.files.length > 0) {
            this.chartFile = event.target.files[0];
        }
    }
    onProvFileChangeEvent(event) {
        if (event.target.files && event.target.files.length > 0) {
            this.provFile = event.target.files[0];
        }
    }

    deleteVersionCard(env: Event, version: HelmChartVersion) {
        env.stopPropagation();
        this.openVersionDeleteModal([version]);
    }

    openVersionDeleteModal(versions?: HelmChartVersion[]) {
        if (!versions) {
            versions = this.selectedRows;
        }
        let versionNames = versions.map(v => v.version).join(',');
        let message = new ConfirmationMessage(
            'HELM_CHART.DELETE_CHART_VERSION_TITLE',
            'HELM_CHART.DELETE_CHART_VERSION',
            versionNames,
            versions,
            ConfirmationTargets.HELM_CHART_VERSION,
            ConfirmationButtons.DELETE_CANCEL
        );
        this.confirmationDialog.open(message);
    }

    confirmDeletion(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.source === ConfirmationTargets.HELM_CHART_VERSION &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            let versions = message.data;
            this.deleteVersions(versions);
        }
    }

    getImgLink(v: HelmChartVersion) {
        if (v.icon) {
            return v.icon;
        } else {
            return DefaultHelmIcon;
        }
    }

    getDefaultIcon(v: HelmChartVersion) {
        v.icon = this.chartDefaultIcon;
    }

    getStatusString(chartVersion: HelmChartVersion) {
        if (chartVersion.deprecated) {
            return 'HELM_CHART.DEPRECATED';
        } else {
            return 'HELM_CHART.ACTIVE';
        }
    }

    onLabelChange(version: HelmChartVersion) {
        this.resrouceLabelService
            .getChartVersionLabels(
                this.projectName,
                this.chartName,
                version.version
            )
            .subscribe(labels => {
                let versionIdx = this.chartVersions.findIndex(
                    v => v.name === version.name
                );
                this.chartVersions[versionIdx].labels = labels;
            });
    }

    getHelmChartVersionPermission(projectId: number): void {
        let hasAddRemoveHelmChartVersionPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.HELM_CHART_VERSION_LABEL.KEY,
                USERSTATICPERMISSION.HELM_CHART_VERSION_LABEL.VALUE.CREATE
            );
        let hasDownloadHelmChartVersionPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.HELM_CHART_VERSION.KEY,
                USERSTATICPERMISSION.HELM_CHART_VERSION.VALUE.READ
            );
        let hasDeleteHelmChartVersionPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.HELM_CHART_VERSION.KEY,
                USERSTATICPERMISSION.HELM_CHART_VERSION.VALUE.DELETE
            );
        forkJoin(
            hasAddRemoveHelmChartVersionPermission,
            hasDownloadHelmChartVersionPermission,
            hasDeleteHelmChartVersionPermission
        ).subscribe(
            permissions => {
                this.hasAddRemoveHelmChartVersionPermission =
                    permissions[0] as boolean;
                this.hasDownloadHelmChartVersionPermission =
                    permissions[1] as boolean;
                this.hasDeleteHelmChartVersionPermission =
                    permissions[2] as boolean;
            },
            error => this.errorHandlerEntity.error(error)
        );
    }
}
