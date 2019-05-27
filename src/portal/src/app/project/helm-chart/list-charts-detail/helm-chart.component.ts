import {
  Component,
  Input,
  OnInit,
  ChangeDetectionStrategy,
  Output,
  EventEmitter,
  ViewChild,
  ChangeDetectorRef
} from "@angular/core";
import { NgForm } from '@angular/forms';
import { TranslateService } from "@ngx-translate/core";
import { State, ErrorHandler, SystemInfo, SystemInfoService, DEFAULT_PAGE_SIZE, downloadFile
  , OperationService, UserPermissionService, USERSTATICPERMISSION, OperateInfo, OperationState, operateChanges
  , errorHandler as errorHandFn } from "@harbor/ui";
import { forkJoin, throwError as observableThrowError, Observable } from "rxjs";
import { finalize, map, catchError } from "rxjs/operators";
import { HelmChartItem } from "../helm-chart.interface.service";
import { HelmChartService } from "../helm-chart.service";
import { DefaultHelmIcon, Roles } from "../../../shared/shared.const";

import { ConfirmationAcknowledgement } from "../../../shared/confirmation-dialog/confirmation-state-message";
import {  ConfirmationDialogComponent } from "../../../shared/confirmation-dialog/confirmation-dialog.component";
import { ConfirmationMessage } from "../../../shared/confirmation-dialog/confirmation-message";
import {
  ConfirmationButtons,
  ConfirmationTargets,
  ConfirmationState,
} from "../../../shared/shared.const";

@Component({
  selector: "hbr-helm-chart",
  templateUrl: "./helm-chart.component.html",
  styleUrls: ["./helm-chart.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class HelmChartComponent implements OnInit {
  signedCon: { [key: string]: any | string[] } = {};
  @Input() projectId: number;
  @Input() projectName = "unknown";
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

  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage = 1;
  totalCount = 0;
  currentState: State;

  @ViewChild('chartUploadForm') uploadForm: NgForm;

  @ViewChild("confirmationDialog") confirmationDialog: ConfirmationDialogComponent;
  hasUploadHelmChartsPermission: boolean;
  hasDownloadHelmChartsPermission: boolean;
  hasDeleteHelmChartsPermission: boolean;
  constructor(
    private errorHandler: ErrorHandler,
    private translateService: TranslateService,
    private systemInfoService: SystemInfoService,
    private helmChartService: HelmChartService,
    private userPermissionService: UserPermissionService,
    private operationService: OperationService,
    private cdr: ChangeDetectorRef,
  ) { }

  public get registryUrl(): string {
    return this.systemInfo ? this.systemInfo.registry_url : "";
  }

  ngOnInit(): void {
    // Get system info for tag views
    this.systemInfoService.getSystemInfo()
      .subscribe(systemInfo => (this.systemInfo = systemInfo)
        , error => this.errorHandler.error(error));
    this.lastFilteredChartName = "";
    this.refresh();
    this.getHelmPermissionRule(this.projectId);
  }
  getHelmPermissionRule(projectId: number): void {
    let hasUploadHelmChartsPermission = this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.HELM_CHART.KEY, USERSTATICPERMISSION.HELM_CHART.VALUE.UPLOAD);
    let hasDownloadHelmChartsPermission = this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.HELM_CHART.KEY, USERSTATICPERMISSION.HELM_CHART.VALUE.DOWNLOAD);
    let hasDeleteHelmChartsPermission = this.userPermissionService.getPermission(projectId,
      USERSTATICPERMISSION.HELM_CHART.KEY, USERSTATICPERMISSION.HELM_CHART.VALUE.DELETE);
    forkJoin(hasUploadHelmChartsPermission, hasDownloadHelmChartsPermission, hasDeleteHelmChartsPermission).subscribe(permissions => {
      this.hasUploadHelmChartsPermission = permissions[0] as boolean;
      this.hasDownloadHelmChartsPermission = permissions[1] as boolean;
      this.hasDeleteHelmChartsPermission = permissions[2] as boolean;
    }, error => this.errorHandler.error(error));
  }
  updateFilterValue(value: string) {
    this.lastFilteredChartName = value;
    this.refresh();
  }

  refresh() {
    this.loading = true;
    this.helmChartService
      .getHelmCharts(this.projectName)
      .pipe(finalize(() => {
        let hnd = setInterval(() => this.cdr.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 3000);
        this.loading = false;
      }))
      .subscribe(
        charts => {
          this.charts = charts.filter(x => x.name.includes(this.lastFilteredChartName));
          this.chartsCopy = charts.map(x => Object.assign({}, x));
          this.totalCount = charts.length;
        },
        err => {
          this.errorHandler.error(err);
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
    if (this.isUploading) { return; }
    this.isUploading = true;
    this.helmChartService
      .uploadChart(this.projectName, this.chartFile, this.provFile)
      .pipe(finalize(() => {
        this.isUploading = false;
        this.isUploadModalOpen = false;
        this.refresh();
      }))
      .subscribe(() => {
        this.translateService
          .get("HELM_CHART.FILE_UPLOADED")
          .subscribe(res => this.errorHandler.info(res));
      },
        err => this.errorHandler.error(err)
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
    operateMsg.name = "OPERATION.DELETE_CHART";
    operateMsg.data.id = chartName;
    operateMsg.state = OperationState.progressing;
    operateMsg.data.name = chartName;
    this.operationService.publishInfo(operateMsg);

    return this.helmChartService.deleteHelmChart(this.projectName, chartName)
      .pipe(map(
        () => operateChanges(operateMsg, OperationState.success),
        catchError( error => {
          const message = errorHandFn(error);
          this.translateService.get(message).subscribe(res =>
            operateChanges(operateMsg, OperationState.failure, res)
          );
          return observableThrowError(message);
        }
      )));
  }

  deleteCharts(charts: HelmChartItem[]) {
    if (charts && charts.length < 1) { return; }
    let chartsDelete$ = charts.map(chart => this.deleteChart(chart.name));
    forkJoin(chartsDelete$)
      .pipe(
        catchError(err => observableThrowError(err)),
        finalize(() => {
          this.refresh();
          this.selectedRows = [];
        }))
      .subscribe(() => { });
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
    this.helmChartService.downloadChart(this.projectName, filename).subscribe(
      res => {
        downloadFile(res);
      },
      error => {
        this.errorHandler.error(error);
      }
    );
  }

  openChartDeleteModal(charts: HelmChartItem[]) {
    let chartNames = charts.map(chart => chart.name).join(",");
    let message = new ConfirmationMessage(
      "HELM_CHART.DELETE_CHART_VERSION_TITLE",
      "HELM_CHART.DELETE_CHART_VERSION",
      chartNames,
      charts,
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
    if (itemName === "card") {
      this.cardHover = true;
    } else {
      this.listHover = true;
    }
  }

  mouseLeave(itemName: string) {
    if (itemName === "card") {
      this.cardHover = false;
    } else {
      this.listHover = false;
    }
  }

  isHovering(itemName: string) {
    if (itemName === "card") {
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
      return "HELM_CHART.DEPRECATED";
    } else {
      return "HELM_CHART.ACTIVE";
    }
  }
}
