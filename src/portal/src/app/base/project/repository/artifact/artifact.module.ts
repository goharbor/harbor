import { NgModule } from '@angular/core';
import { RouterModule, Routes } from "@angular/router";
import { SharedModule } from "../../../../shared/shared.module";
import { ArtifactListPageComponent } from "./artifact-list-page/artifact-list-page.component";
import { ArtifactListComponent } from "./artifact-list-page/artifact-list/artifact-list.component";
import { ArtifactListTabComponent } from "./artifact-list-page/artifact-list/artifact-list-tab/artifact-list-tab.component";
import { ArtifactSummaryComponent } from "./artifact-summary.component";
import { ArtifactTagComponent } from "./artifact-tag/artifact-tag.component";
import { ArtifactCommonPropertiesComponent } from "./artifact-common-properties/artifact-common-properties.component";
import { ArtifactAdditionsComponent } from "./artifact-additions/artifact-additions.component";
import { ValuesComponent } from "./artifact-additions/values/values.component";
import { SummaryComponent } from "./artifact-additions/summary/summary.component";
import { DependenciesComponent } from "./artifact-additions/dependencies/dependencies.component";
import { BuildHistoryComponent } from "./artifact-additions/build-history/build-history.component";
import { ArtifactVulnerabilitiesComponent } from "./artifact-additions/artifact-vulnerabilities/artifact-vulnerabilities.component";
import { ArtifactDefaultService, ArtifactService } from "./artifact.service";
import { ArtifactDetailRoutingResolverService } from "../../../../services/routing-resolvers/artifact-detail-routing-resolver.service";
import { ResultTipComponent } from "./vulnerability-scanning/result-tip.component";
import { ResultBarChartComponent } from "./vulnerability-scanning/result-bar-chart.component";
import { ResultTipHistogramComponent } from "./vulnerability-scanning/result-tip-histogram/result-tip-histogram.component";
import { HistogramChartComponent } from "./vulnerability-scanning/histogram-chart/histogram-chart.component";

const routes: Routes = [
  {
    path: ':repo',
    component: ArtifactListPageComponent,
  },
  {
    path: ':repo/depth/:depth',
    component: ArtifactListPageComponent,
  },
  {
    path: ':repo/artifacts/:digest',
    component: ArtifactSummaryComponent,
    resolve: {
      artifactResolver: ArtifactDetailRoutingResolverService
    }
  },
  {
    path: ':repo/depth/:depth/artifacts/:digest',
    component: ArtifactSummaryComponent,
    resolve: {
      artifactResolver: ArtifactDetailRoutingResolverService
    }
  },
];
@NgModule({
  declarations: [
    ArtifactListPageComponent,
    ArtifactListComponent,
    ArtifactListTabComponent,
    ArtifactSummaryComponent,
    ArtifactTagComponent,
    ArtifactCommonPropertiesComponent,
    ArtifactAdditionsComponent,
    ValuesComponent,
    SummaryComponent,
    DependenciesComponent,
    BuildHistoryComponent,
    ArtifactVulnerabilitiesComponent,
    ResultTipComponent,
    ResultBarChartComponent,
    ResultTipHistogramComponent,
    HistogramChartComponent
  ],
  imports: [
    RouterModule.forChild(routes),
    SharedModule
  ],
  providers: [
    ArtifactDefaultService,
    {provide: ArtifactService, useClass: ArtifactDefaultService },
  ]
})
export class ArtifactModule { }
