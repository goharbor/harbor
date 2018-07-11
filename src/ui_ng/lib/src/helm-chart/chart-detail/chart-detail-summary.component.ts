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

  constructor() {}

  ngOnInit(): void {
  }

  public get addCMD() {
    return `helm repo add REPO_NAME ${this.repoURL}/chartrepo/${this.projectName}`;
  }

  public get installCMD() {
      return `helm install --version ${this.chartVersion} REPO_NAME/${this.chartName}`;
  }

}
