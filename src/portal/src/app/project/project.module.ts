// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { NgModule } from '@angular/core';
import { RouterModule } from '@angular/router';
import { SharedModule } from '../shared/shared.module';
import { ReplicationModule } from '../replication/replication.module';
import { SummaryModule } from './summary/summary.module';
import { TagFeatureIntegrationModule } from './tag-feature-integration/tag-feature-integration.module';
import { LogModule } from '../log/log.module';
import { ProjectComponent } from './project.component';
import { CreateProjectComponent } from './create-project/create-project.component';
import { ListProjectComponent } from './list-project/list-project.component';
import { ProjectDetailComponent } from './project-detail/project-detail.component';
import { MemberComponent } from './member/member.component';
import { AddMemberComponent } from './member/add-member/add-member.component';
import { AddGroupComponent } from './member/add-group/add-group.component';
import { MemberService } from './member/member.service';
import { RobotService } from './robot-account/robot-account.service';
import { ProjectRoutingResolver } from './project-routing-resolver.service';
import { TargetExistsValidatorDirective } from '../shared/target-exists-directive';
import { HelmChartModule } from './helm-chart/helm-chart.module';
import { RobotAccountComponent } from './robot-account/robot-account.component';
import { AddRobotComponent } from './robot-account/add-robot/add-robot.component';
import { AddHttpAuthGroupComponent } from './member/add-http-auth-group/add-http-auth-group.component';
import { WebhookService } from './webhook/webhook.service';
import { WebhookComponent } from './webhook/webhook.component';
import { AddWebhookComponent } from './webhook/add-webhook/add-webhook.component';
import { AddWebhookFormComponent } from './webhook/add-webhook-form/add-webhook-form.component';
import { ScannerComponent } from "./scanner/scanner.component";
import { ConfigScannerService } from "../config/scanner/config-scanner.service";
import { RepositoryGridviewComponent } from "./repository/repository-gridview.component";
import { ResultTipHistogramComponent } from "./repository/vulnerability-scanning/result-tip-histogram/result-tip-histogram.component";
import { ResultGridComponent } from "./repository/vulnerability-scanning/result-grid.component";
import { ResultBarChartComponent } from "./repository/vulnerability-scanning/result-bar-chart.component";
import { HistogramChartComponent } from "./repository/vulnerability-scanning/histogram-chart/histogram-chart.component";
import { ResultTipComponent } from "./repository/vulnerability-scanning/result-tip.component";
import { ArtifactListPageComponent } from "./repository/artifact-list-page/artifact-list-page.component";
import { ProjectLabelComponent } from "./project-label/project-label.component";
import { ArtifactListComponent } from "./repository/artifact-list-page/artifact-list/artifact-list.component";
import { ArtifactTagComponent } from "./repository/artifact/artifact-tag/artifact-tag.component";
import { ArtifactCommonPropertiesComponent } from "./repository/artifact/artifact-common-properties/artifact-common-properties.component";
import { ArtifactAdditionsComponent } from "./repository/artifact/artifact-additions/artifact-additions.component";
import { ArtifactSummaryComponent } from "./repository/artifact/artifact-summary.component";
import { ArtifactListTabComponent } from "./repository/artifact-list-page/artifact-list/artifact-list-tab/artifact-list-tab.component";
import { BuildHistoryComponent } from "./repository/artifact/artifact-additions/build-history/build-history.component";
import { DependenciesComponent } from "./repository/artifact/artifact-additions/dependencies/dependencies.component";
import { SummaryComponent } from "./repository/artifact/artifact-additions/summary/summary.component";
import { ValuesComponent } from "./repository/artifact/artifact-additions/values/values.component";
import {
  ArtifactVulnerabilitiesComponent
} from "./repository/artifact/artifact-additions/artifact-vulnerabilities/artifact-vulnerabilities.component";
import { RepositoryDefaultService, RepositoryService } from "./repository/repository.service";
import { ArtifactDefaultService, ArtifactService } from "./repository/artifact/artifact.service";
import { GridViewComponent } from "./repository/gridview/grid-view.component";

@NgModule({
  imports: [
    SharedModule,
    ReplicationModule,
    LogModule,
    RouterModule,
    HelmChartModule,
    SummaryModule,
    TagFeatureIntegrationModule,
  ],
  declarations: [
    ProjectComponent,
    CreateProjectComponent,
    ListProjectComponent,
    ProjectDetailComponent,
    MemberComponent,
    AddMemberComponent,
    TargetExistsValidatorDirective,
    ProjectLabelComponent,
    AddGroupComponent,
    RobotAccountComponent,
    AddRobotComponent,
    AddHttpAuthGroupComponent,
    WebhookComponent,
    AddWebhookComponent,
    AddWebhookFormComponent,
    ScannerComponent,
    RepositoryGridviewComponent,
    HistogramChartComponent,
    ResultTipHistogramComponent,
    ResultBarChartComponent,
    ResultGridComponent,
    ResultTipComponent,
    ArtifactListPageComponent,
    ArtifactListComponent,
    ArtifactListTabComponent,
    ArtifactSummaryComponent,
    ArtifactCommonPropertiesComponent,
    ArtifactTagComponent,
    ArtifactAdditionsComponent,
    BuildHistoryComponent,
    DependenciesComponent,
    SummaryComponent,
    ValuesComponent,
    ArtifactVulnerabilitiesComponent,
    GridViewComponent,
  ],
  exports: [ProjectComponent, ListProjectComponent],
  providers: [
    ProjectRoutingResolver,
    MemberService,
    RobotService,
    WebhookService,
    ConfigScannerService,
    RepositoryDefaultService,
    ArtifactDefaultService,
    { provide: RepositoryService, useClass: RepositoryDefaultService },
    { provide: ArtifactService, useClass: ArtifactDefaultService },
  ]
})
export class ProjectModule {

}
