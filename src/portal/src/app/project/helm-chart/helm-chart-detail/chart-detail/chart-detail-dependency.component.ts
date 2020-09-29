import {
  Component,
  OnInit,
  Input,
  ChangeDetectionStrategy
} from "@angular/core";

import { HelmChartDependency } from "../../helm-chart.interface.service";

@Component({
  selector: "hbr-chart-detail-dependency",
  templateUrl: "./chart-detail-dependency.component.html",
  styleUrls: ["./chart-detail-dependency.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ChartDetailDependencyComponent implements OnInit {
  @Input() dependencies: HelmChartDependency;

  constructor() {}

  ngOnInit(): void {
  }

}
