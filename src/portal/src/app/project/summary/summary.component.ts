import { Component, OnInit, Input } from '@angular/core';
import { ProjectService, clone, QuotaUnits, getSuitableUnit, ErrorHandler, GetIntegerAndUnit } from '@harbor/ui';
import { Router, ActivatedRoute } from '@angular/router';
import { forkJoin } from 'rxjs';
import { AppConfigService } from "../../app-config.service";
export const riskRatio = 0.9;
@Component({
  selector: 'summary',
  templateUrl: './summary.component.html',
  styleUrls: ['./summary.component.scss']
})
export class SummaryComponent implements OnInit {
  projectId: number;
  summaryInformation: any;
  constructor(
    private projectService: ProjectService,
    private errorHandler: ErrorHandler,
    private appConfigService: AppConfigService,
    private route: ActivatedRoute
    ) { }

  ngOnInit() {
    this.projectId = this.route.snapshot.parent.params['id'];
    this.projectService.getProjectSummary(this.projectId).subscribe(res => {
      this.summaryInformation = res;
    }, error => {
      this.errorHandler.error(error);
    });
  }
  getSuitableUnit(value) {
    const QuotaUnitsCopy = clone(QuotaUnits);
    return getSuitableUnit(value, QuotaUnitsCopy);
  }
  getIntegerAndUnit(hardValue, usedValue) {
    return GetIntegerAndUnit(hardValue, clone(QuotaUnits), usedValue, clone(QuotaUnits));
  }
  public get withHelmChart(): boolean {
    return this.appConfigService.getConfig().with_chartmuseum;
  }

}
