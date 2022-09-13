import {
    Component,
    EventEmitter,
    Input,
    OnInit,
    Output,
    ViewChild,
} from '@angular/core';
import { NgForm } from '@angular/forms';
import { TranslateService } from '@ngx-translate/core';
import { forkJoin, Observable, throwError as observableThrowError } from 'rxjs';
import { catchError, finalize, map } from 'rxjs/operators';
import { HelmChartItem } from '../../helm-chart-detail/helm-chart.interface.service';
import { HelmChartService } from '../../helm-chart-detail/helm-chart.service';
import {
    State,
    SystemInfo,
    SystemInfoService,
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../../../shared/services';
import {
    downloadFile,
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../../shared/units/utils';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { OperationService } from '../../../../../shared/components/operation/operation.service';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../../shared/components/operation/operate';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
    DefaultHelmIcon,
    Roles,
} from '../../../../../shared/entities/shared.const';
import { errorHandler } from '../../../../../shared/units/shared.utils';
import { ConfirmationDialogComponent } from '../../../../../shared/components/confirmation-dialog';
import { ConfirmationMessage } from '../../../../global-confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../../../../global-confirmation-dialog/confirmation-state-message';
import { ClrDatagridStateInterface } from '@clr/angular';

@Component({
    selector: 'hbr-helm-chart',
    templateUrl: './helm-chart.component.html',
    styleUrls: ['./helm-chart.component.scss'],
})
export class HelmChartComponent implements OnInit {
    signedCon: { [key: string]: any | string[] } = {};
    @Input() projectId: number;
    @Input() projectName = 'unknown';
    @Input() urlPrefix: string;
    @Input() hasSignedIn: boolean;
    @Input() projectRoleID = Roles.OTHER;
    @Output() chartClickEvt = new EventEmitter<any>();
    @Output() chartDownloadEve = new EventEmitter<string>();
    @Input() chartDefaultIcon: string = DefaultHelmIcon;

    lastFilteredChartName: string;
    charts: HelmChartItem[] = [];
    chartsCopy: HelmChartItem[] = [];
    systemInfo: SystemInfo;
    selectedRows: HelmChartItem[] = [];
    loading = true;

    // For Upload
    isUploading = false;
    isUploadModalOpen = false;
    provFile: File;
    chartFile: File;

    // For View swtich
    isCardView: boolean;
    cardHover = false;
    listHover = false;

    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.HELM_CHART_COMPONENT
    );
    currentPage = 1;
    totalCount = 0;
    currentState: State;

    @ViewChild('chartUploadForm') uploadForm: NgForm;

    @ViewChild('confirmationDialog')
    confirmationDialog: ConfirmationDialogComponent;
    hasUploadHelmChartsPermission: boolean;
    hasDownloadHelmChartsPermission: boolean;
    hasDeleteHelmChartsPermission: boolean;
    constructor(
        private errorHandlerEntity: ErrorHandler,
        private translateService: TranslateService,
        private systemInfoService: SystemInfoService,
        private helmChartService: HelmChartService,
        private userPermissionService: UserPermissionService,
        private operationService: OperationService
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
        this.lastFilteredChartName = '';
        this.refresh();
        this.getHelmPermissionRule(this.projectId);
    }
    getHelmPermissionRule(projectId: number): void {
        let hasUploadHelmChartsPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.HELM_CHART.KEY,
                USERSTATICPERMISSION.HELM_CHART.VALUE.UPLOAD
            );
        let hasDownloadHelmChartsPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.HELM_CHART.KEY,
                USERSTATICPERMISSION.HELM_CHART.VALUE.DOWNLOAD
            );
        let hasDeleteHelmChartsPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.HELM_CHART.KEY,
                USERSTATICPERMISSION.HELM_CHART.VALUE.DELETE
            );
        forkJoin(
            hasUploadHelmChartsPermission,
            hasDownloadHelmChartsPermission,
            hasDeleteHelmChartsPermission
        ).subscribe(
            permissions => {
                this.hasUploadHelmChartsPermission = permissions[0] as boolean;
                this.hasDownloadHelmChartsPermission =
                    permissions[1] as boolean;
                this.hasDeleteHelmChartsPermission = permissions[2] as boolean;
            },
            error => this.errorHandlerEntity.error(error)
        );
    }
    updateFilterValue(value: string) {
        this.lastFilteredChartName = value;
        this.refresh();
    }

    refresh() {
        this.loading = true;
        this.selectedRows = [];
        this.helmChartService
            .getHelmCharts(this.projectName)
            .pipe(
                finalize(() => {
                    this.loading = false;
                })
            )
            .subscribe(
                charts => {
                    this.charts = charts.filter(x =>
                        x.name.includes(this.lastFilteredChartName)
                    );
                    this.chartsCopy = charts.map(x => Object.assign({}, x));
                    this.totalCount = charts.length;
                },
                err => {
                    this.errorHandlerEntity.error(err);
                }
            );
    }

    onChartClick(item: HelmChartItem) {
        this.chartClickEvt.emit(item.name);
    }

    resetUploadForm() {
        this.chartFile = null;
        this.provFile = null;
        this.uploadForm.reset();
    }

    onChartUpload() {
        this.resetUploadForm();
        this.isUploadModalOpen = true;
    }

    cancelUpload() {
        this.resetUploadForm();
        this.isUploadModalOpen = false;
    }

    upload() {
        if (!this.chartFile && !this.provFile) {
            return;
        }
        if (this.isUploading) {
            return;
        }
        this.isUploading = true;
        this.helmChartService
            .uploadChart(this.projectName, this.chartFile, this.provFile)
            .pipe(
                finalize(() => {
                    this.isUploading = false;
                    this.isUploadModalOpen = false;
                    this.refresh();
                })
            )
            .subscribe(
                () => {
                    this.translateService
                        .get('HELM_CHART.FILE_UPLOADED')
                        .subscribe(res => this.errorHandlerEntity.info(res));
                },
                err => this.errorHandlerEntity.error(err)
            );
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

    deleteChart(chartName: string): Observable<any> {
        let operateMsg = new OperateInfo();
        operateMsg.name = 'OPERATION.DELETE_CHART';
        operateMsg.data.id = chartName;
        operateMsg.state = OperationState.progressing;
        operateMsg.data.name = chartName;
        this.operationService.publishInfo(operateMsg);

        return this.helmChartService
            .deleteHelmChart(this.projectName, chartName)
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

    deleteCharts(charts: HelmChartItem[]) {
        if (charts && charts.length < 1) {
            return;
        }
        let chartsDelete$ = charts.map(chart => this.deleteChart(chart.name));
        forkJoin(chartsDelete$)
            .pipe(
                finalize(() => {
                    this.refresh();
                    this.selectedRows = [];
                })
            )
            .subscribe(
                () => {},
                error => {
                    this.errorHandlerEntity.error(error);
                }
            );
    }

    downloadLatestVersion(evt?: Event, item?: HelmChartItem) {
        if (evt) {
            evt.stopPropagation();
        }
        let selectedChart: HelmChartItem;

        if (item) {
            selectedChart = item;
        } else {
            // return if selected version less then 1
            if (this.selectedRows.length < 1) {
                return;
            }
            selectedChart = this.selectedRows[0];
        }
        if (!selectedChart) {
            return;
        }
        let filename = `charts/${selectedChart.name}-${selectedChart.latest_version}.tgz`;
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

    openChartDeleteModal() {
        let chartNames = this.selectedRows.map(chart => chart.name).join(',');
        let message = new ConfirmationMessage(
            'HELM_CHART.DELETE_CHART_VERSION_TITLE',
            'HELM_CHART.DELETE_CHART_VERSION',
            chartNames,
            this.selectedRows,
            ConfirmationTargets.HELM_CHART,
            ConfirmationButtons.DELETE_CANCEL
        );
        this.confirmationDialog.open(message);
    }

    confirmDeletion(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.source === ConfirmationTargets.HELM_CHART &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            let charts = message.data;
            this.deleteCharts(charts);
        }
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

    getDefaultIcon(chart: HelmChartItem) {
        chart.icon = this.chartDefaultIcon;
    }

    getStatusString(chart: HelmChartItem) {
        if (chart.deprecated) {
            return 'HELM_CHART.DEPRECATED';
        } else {
            return 'HELM_CHART.ACTIVE';
        }
    }
    clrLoad(state: ClrDatagridStateInterface) {
        if (state?.page?.size) {
            setPageSizeToLocalStorage(
                PageSizeMapKeys.HELM_CHART_COMPONENT,
                state.page.size
            );
        }
    }
}
