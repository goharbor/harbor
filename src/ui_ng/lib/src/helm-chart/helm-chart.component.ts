import {
  Component,
  Input,
  OnInit,
  ChangeDetectionStrategy,
  Output,
  EventEmitter,
  ChangeDetectorRef
} from "@angular/core";
import { TranslateService } from "@ngx-translate/core";
import { State } from "clarity-angular";

import { SystemInfo, SystemInfoService, HelmChartItem } from "../service/index";
import { ErrorHandler } from "../error-handler/error-handler";
import { toPromise, DEFAULT_PAGE_SIZE } from "../utils";
import { HelmChartService } from "../service/helm-chart.service";

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
  @Input() hasProjectAdminRole: boolean;
  @Output() chartClickEvt = new EventEmitter<any>();
  @Output() chartDownloadEve = new EventEmitter<string>();

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

  constructor(
    private errorHandler: ErrorHandler,
    private translateService: TranslateService,
    private systemInfoService: SystemInfoService,
    private helmChartService: HelmChartService,
    private cdr: ChangeDetectorRef,
  ) {}

  public get registryUrl(): string {
    return this.systemInfo ? this.systemInfo.registry_url : "";
  }

  ngOnInit(): void {
    // Get system info for tag views
    toPromise<SystemInfo>(this.systemInfoService.getSystemInfo())
      .then(systemInfo => (this.systemInfo = systemInfo))
      .catch(error => this.errorHandler.error(error));
    this.lastFilteredChartName = "";
    this.refresh();

  }

  updateFilterValue(value: string) {
    this.lastFilteredChartName = value;
    this.refresh();
  }

  refresh() {
    this.loading = true;
    this.helmChartService
    .getHelmCharts(this.projectName)
    .finally(() => {
        let hnd = setInterval(() => this.cdr.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 3000);
        this.loading = false;
    })
    .subscribe(
      charts => {
        this.charts = charts.filter(x => x.name.includes(this.lastFilteredChartName));
        this.chartsCopy = charts.map(x => Object.assign({}, x));
      },
      err => {
        this.errorHandler.error(err);
      }
    );
  }

  onChartClick(item: HelmChartItem) {
    this.chartClickEvt.emit(item.name);
  }

  onChartUpload() {
    this.isUploadModalOpen = true;
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
      })
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
}
