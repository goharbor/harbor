import {
  Component,
  OnInit,
  ChangeDetectionStrategy,
  Input
} from "@angular/core";

import { HelmChartMetaData, HelmChartSecurity } from "./../../service/interface";

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

  copiedCMD = '';

  constructor() {}

  ngOnInit(): void {
  }

  isCopied(cmd: string) {
    return this.copiedCMD === cmd;
  }

  onCopySuccess(e: Event, cmd: string) {
    this.copiedCMD = cmd;
  }


  public get addCMD() {
    return `helm repo add --ca-file <ca file> --cert-file <cert file> --key-file <key file> --username <username> --password <password> <repo name> ${this.repoURL}/chartrepo/${this.projectName}`;
  }

  public get installCMD() {
      return `helm install --ca-file <ca file> --cert-file <cert file> --key-file <key file> --username=<username> --password=<password> --version ${this.chartVersion} <repo name>/${this.chartName}`;
  }

  public get verifyCMD() {
    return `helm verify --keyring <key path> ${this.chartName}-${this.chartVersion}.tgz`;
}

  public get prov_ready() {
    return this.security && this.security.signature && this.security.signature.signed;
  }

}
