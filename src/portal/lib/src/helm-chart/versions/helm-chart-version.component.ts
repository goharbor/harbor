import {
  Component,
  Input,
  OnInit,
  ViewChild,
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Output,
  EventEmitter
} from "@angular/core";
import { Observable, forkJoin } from "rxjs";
import {finalize, map} from "rxjs/operators";

import { TranslateService } from "@ngx-translate/core";
import { State } from "@clr/angular";

import {
  SystemInfo,
  SystemInfoService,
  HelmChartVersion,
  HelmChartMaintainer,
  LabelService
} from "./../../service/index";
import { Label } from './../../service/interface';
import { ErrorHandler } from "./../../error-handler/error-handler";
import { toPromise, DEFAULT_PAGE_SIZE, downloadFile } from "../../utils";
import { OperationService } from "./../../operation/operation.service";
import { HelmChartService } from "./../../service/helm-chart.service";
import { ConfirmationAcknowledgement, ConfirmationDialogComponent, ConfirmationMessage } from "./../../confirmation-dialog";
import {
  OperateInfo,
  OperationState,
  operateChanges
} from "./../../operation/operate";
import {
  ConfirmationButtons,
  ConfirmationTargets,
  ConfirmationState,
  DefaultHelmIcon,
  ResourceType
} from "../../shared/shared.const";

@Component({
  selector: "hbr-helm-chart-version",
  templateUrl: "./helm-chart-version.component.html",
  styleUrls: ["./helm-chart-version.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ChartVersionComponent implements OnInit {
  signedCon: { [key: string]: any | string[] } = {};
  @Input() projectId: number;
  @Input() projectName: string;
  @Input() chartName: string;
  @Input() roleName: string;
  @Input() hasSignedIn: boolean;
  @Input() hasProjectAdminRole: boolean;
  @Input() chartDefaultIcon: string = DefaultHelmIcon;
  @Output() versionClickEvt = new EventEmitter<string>();
  @Output() backEvt = new EventEmitter<any>();


  lastFilteredVersionName: string;
  chartVersions: HelmChartVersion[] = [];
  systemInfo: SystemInfo;
  selectedRows: HelmChartVersion[] = [];
  projectLabels: Label[] = [];
  loading = true;
  resourceType = ResourceType.CHART_VERSION;

  isCardView: boolean;
  cardHover = false;
  listHover = false;

  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage = 1;
  totalCount = 0;
  currentState: State;

  chartFile: File;
  provFile: File;

  @ViewChild("confirmationDialog")
  confirmationDialog: ConfirmationDialogComponent;

  constructor(
    private errorHandler: ErrorHandler,
    private systemInfoService: SystemInfoService,
    private helmChartService: HelmChartService,
    private resrouceLabelService: LabelService,
    private cdr: ChangeDetectorRef,
    private operationService: OperationService,
  ) {}

  public get registryUrl(): string {
    return this.systemInfo ? this.systemInfo.registry_url : "";
  }

  ngOnInit(): void {
    // Get system info for tag views
    toPromise<SystemInfo>(this.systemInfoService.getSystemInfo())
      .then(systemInfo => (this.systemInfo = systemInfo))
      .catch(error => this.errorHandler.error(error));
    this.refresh();
    this.getProjectLabels();
    this.lastFilteredVersionName = "";
  }

  updateFilterValue(value: string) {
    this.lastFilteredVersionName = value;
    this.refresh();
  }

  getProjectLabels() {
    this.resrouceLabelService.getProjectLabels(this.projectId).subscribe(
      (labels: Label[]) => {
        this.projectLabels = labels;
      }
      );
  }

  refresh() {
    this.loading = true;
    this.helmChartService
      .getChartVersions(this.projectName, this.chartName)
      .pipe(finalize(() => {
        this.selectedRows = [];
        this.loading = false;
        let hnd = setInterval(() => this.cdr.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 2000);
      }))
      .subscribe(
        versions => {
          this.chartVersions = versions.filter(x => x.version.includes(this.lastFilteredVersionName));
          this.totalCount = versions.length;
        },
        err => {
          this.errorHandler.error(err);
        }
      );
  }

  getMaintainerString(maintainers: HelmChartMaintainer[]) {
    if (!maintainers || maintainers.length < 1) {
      return "";
    }

    let maintainer_string = maintainers[0].name;
    if (maintainers.length > 1) {
      maintainer_string = `${maintainer_string} (${maintainers.length - 1} others)`;
    }
    return maintainer_string;
  }

  onVersionClick(version: HelmChartVersion) {
    this.versionClickEvt.emit(version.version);
  }

  deleteVersion(version: HelmChartVersion): Observable<any> {
    // init operation info
    let operateMsg = new OperateInfo();
    operateMsg.name = "OPERATION.DELETE_CHART_VERSION";
    operateMsg.data.id = version.digest;
    operateMsg.state = OperationState.progressing;
    operateMsg.data.name = `${version.name}:${version.version}`;
    this.operationService.publishInfo(operateMsg);

    return this.helmChartService
      .deleteChartVersion(this.projectName, this.chartName, version.version)
      .pipe(map(
        () => operateChanges(operateMsg, OperationState.success),
        err => operateChanges(operateMsg, OperationState.failure, err)
      ));
  }

  deleteVersions(versions: HelmChartVersion[]) {
    if (versions && versions.length < 1) { return; }
    let successCount: number;
    let totalCount = this.chartVersions.length;
    let versionObs = versions.map(v => this.deleteVersion(v));
    forkJoin(versionObs).pipe(finalize(() => {
      if (totalCount !== successCount) {
        this.refresh();
      }
    })).subscribe(res => {
      successCount = res.filter(r => r.state === OperationState.success).length;
      if (totalCount === successCount) {
        this.backEvt.emit();
      }
    });
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
    this.helmChartService.downloadChart(this.projectName, filename).subscribe(
      res => {
        downloadFile(res);
      },
      error => {
        this.errorHandler.error(error);
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

  openVersionDeleteModal(versions: HelmChartVersion[]) {
    let versionNames = versions.map(v => v.version).join(",");
    let message = new ConfirmationMessage(
      "HELM_CHART.DELETE_CHART_VERSION_TITLE",
      "HELM_CHART.DELETE_CHART_VERSION",
      versionNames,
      versions,
      ConfirmationTargets.HELM_CHART_VERSION,
      ConfirmationButtons.DELETE_CANCEL
    );
    this.confirmationDialog.open(message);
    let hnd = setInterval(() => this.cdr.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 2000);
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
      return "HELM_CHART.DEPRECATED";
    } else {
      return "HELM_CHART.ACTIVE";
    }
  }

  onLabelChange(version: HelmChartVersion) {
    this.resrouceLabelService.getChartVersionLabels(this.projectName, this.chartName, version.version)
    .subscribe(labels => {
        let versionIdx = this.chartVersions.findIndex(v => v.name === version.name);
        this.chartVersions[versionIdx].labels = labels;
        let hnd = setInterval(() => this.cdr.markForCheck(), 200);
        setTimeout(() => clearInterval(hnd), 5000);
    });
  }
}
