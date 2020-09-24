import {
  Component,
  OnInit,
  ChangeDetectionStrategy,
  Input
} from "@angular/core";
import { HelmChartMetaData, HelmChartSecurity } from "./../../helm-chart.interface.service";
import { HelmChartService } from "../../helm-chart.service";
import { Label } from "../../../../../lib/services";
import { ErrorHandler } from "../../../../../lib/utils/error-handler";
import { downloadFile } from "../../../../../lib/utils/utils";

@Component({
  selector: "hbr-chart-detail-summary",
  templateUrl: "./chart-detail-summary.component.html",
  styleUrls: ["./chart-detail-summary.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ChartDetailSummaryComponent implements OnInit {
  @Input() summary: HelmChartMetaData;
  @Input() security: HelmChartSecurity;
  @Input() repoURL: string;
  @Input() projectName: string;
  @Input() chartName: string;
  @Input() chartVersion: string;
  @Input() readme: string;
  @Input() labels: Label[];

  copiedCMD = '';
  addCMD: string;
  installCMD: string;
  verifyCMD: string;

  constructor(
    private errorHandler: ErrorHandler,
    private helmChartService: HelmChartService
  ) {}

  ngOnInit(): void {
    this.addCMD = `helm repo add --ca-file <ca file> --cert-file <cert file> --key-file <key file> \
    --username <username> --password <password> <repo name> ${this.repoURL}/chartrepo/${this.projectName}`;
    this.installCMD = `helm install --ca-file <ca file> --cert-file <cert file> --key-file <key file> \
    --username=<username> --password=<password> --version ${this.chartVersion} <repo name>/${this.chartName}`;
    this.verifyCMD = `helm verify --keyring <key path> ${this.chartName}-${this.chartVersion}.tgz`;
  }

  isCopied(cmd: string) {
    return this.copiedCMD === cmd;
  }

  onCopySuccess(e: Event, cmd: string) {
    this.copiedCMD = cmd;
  }

  public get prov_ready() {
    return this.security && this.security.signature && this.security.signature.signed;
  }

  downloadChart() {
    if (!this.summary ||
      !this.summary.urls ||
      this.summary.urls.length < 1) {
      return;
    }
    let filename = `${this.summary.urls[0]}.prov`;

    this.helmChartService.downloadChart(this.projectName, filename).subscribe(
      res => {
        downloadFile(res);
      },
      error => {
        this.errorHandler.error(error);
      },
    );
  }

}
