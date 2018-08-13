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
import { NgForm } from "@angular/forms";
import { Observable } from "rxjs/Observable";
import "rxjs/add/observable/forkJoin";

import { TranslateService } from "@ngx-translate/core";
import { State } from "clarity-angular";

import {
  SystemInfo,
  SystemInfoService,
  HelmChartVersion,
  HelmChartMaintainer
} from "./../../service/index";
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
  DefaultHelmIcon
} from "../../shared/shared.const";

@Component({
  selector: "hbr-helm-chart-version",
  templateUrl: "./helm-chart-version.component.html",
  styleUrls: ["./helm-chart-version.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ChartVersionComponent implements OnInit {
  signedCon: { [key: string]: any | string[] } = {};
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
  versionsCopy: HelmChartVersion[] = [];
  systemInfo: SystemInfo;
  selectedRows: HelmChartVersion[] = [];
  loading = true;

  isCardView: boolean;
  cardHover = false;
  listHover = false;

  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage = 1;
  totalCount = 0;
  currentState: State;

  isUploading = false;
  isUploadModalOpen = false;
  chartFile: File;
  provFile: File;

  @ViewChild("confirmationDialog")
  confirmationDialog: ConfirmationDialogComponent;

  @ViewChild("chartUploadForm") form: NgForm;

  constructor(
    private errorHandler: ErrorHandler,
    private translateService: TranslateService,
    private systemInfoService: SystemInfoService,
    private helmChartService: HelmChartService,
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
    this.lastFilteredVersionName = "";
  }

  updateFilterValue(value: string) {
    this.lastFilteredVersionName = value;
    this.refresh();
  }

  refresh() {
    this.loading = true;
    this.helmChartService
      .getChartVersions(this.projectName, this.chartName)
      .finally(() => {
        this.loading = false;
        let hnd = setInterval(() => this.cdr.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 2000);
      })
      .subscribe(
        versions => {
          this.chartVersions = versions.filter(x => x.version.includes(this.lastFilteredVersionName));
          this.versionsCopy = versions.map(x => Object.assign({}, x));
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
      .map(
        () => operateChanges(operateMsg, OperationState.success),
        err => operateChanges(operateMsg, OperationState.failure, err)
      );
  }

  deleteVersions(versions: HelmChartVersion[]) {
    if (versions && versions.length < 1) { return; }
    let versionObs = versions.map(v => this.deleteVersion(v));
    Observable.forkJoin(versionObs).finally(() => this.refresh()).subscribe(res => {
      if (this.chartVersions.length === versionObs.length) {
        this.backEvt.emit();
      }
    });
  }

  versionDownload(evt: Event, item?: HelmChartVersion) {
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
  versionUpload() {
    this.isUploadModalOpen = true;
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

  upload() {
    if (!this.chartFile && !this.provFile) {
      return;
    }
    if (this.isUploading) { return; };
    this.isUploading = true;
    this.helmChartService
      .uploadChart(this.projectName, this.chartFile, this.provFile)
      .finally(() => {
        this.isUploading = false;
        this.isUploadModalOpen = false;
        this.refresh();
        let hnd = setInterval(() => this.cdr.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 3000);
      })
      .subscribe(
        () => {
          this.translateService.get("HELM_CHART.FILE_UPLOADED")
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

  deleteVersionCard(env: Event, version: HelmChartVersion) {
    env.stopPropagation();
    this.openVersionDeleteModal([version]);
  }
  openVersionDeleteModal(versions: HelmChartVersion[]) {
    let versionNames = versions.map(v => v.name).join(",");
    this.translateService.get("HELM_CHART.DELETE_CHART_VERSION").subscribe(key => {
        let message = new ConfirmationMessage(
          "HELM_CHART.DELETE_CHART_VERSION_TITLE",
          key,
          versionNames,
          versions,
          ConfirmationTargets.HELM_CHART,
          ConfirmationButtons.DELETE_CANCEL
        );
        this.confirmationDialog.open(message);
        let hnd = setInterval(() => this.cdr.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 2000);
      });
  }

  confirmDeletion(message: ConfirmationAcknowledgement) {
    if (
      message &&
      message.source === ConfirmationTargets.HELM_CHART &&
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
}
