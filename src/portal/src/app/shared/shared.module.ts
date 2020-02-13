// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { NgModule } from "@angular/core";
import { RouterModule } from "@angular/router";
import { TranslateModule } from "@ngx-translate/core";
import { FormsModule, ReactiveFormsModule } from "@angular/forms";
import { CookieService } from "ngx-cookie";
import { SessionService } from "../shared/session.service";
import { MessageService } from "../global-message/message.service";
import { MessageComponent } from "../global-message/message.component";
import { DateValidatorDirective } from "./date-validator.directive";
import { CoreModule } from "../core/core.module";

import { AuthCheckGuard } from "./route/auth-user-activate.service";
import { SignInGuard } from "./route/sign-in-guard-activate.service";
import { SystemAdminGuard } from "./route/system-admin-activate.service";
import { MemberGuard } from "./route/member-guard-activate.service";
import { ArtifactGuard } from "./route/artifact-guard-activate.service";
import { MemberPermissionGuard } from "./route/member-permission-guard-activate.service";
import { OidcGuard } from "./route/oidc-guard-active.service";
import { LeavingRepositoryRouteDeactivate } from "./route/leaving-repository-deactivate.service";
import { LeavingArtifactSummaryRouteDeactivate } from "./route/leaving-artifact-summary-deactivate.service";

import { PortValidatorDirective } from "./port.directive";
import { MaxLengthExtValidatorDirective } from "./max-length-ext.directive";

import { StatisticHandler } from "./statictics/statistic-handler.service";
import { StatisticsComponent } from "./statictics/statistics.component";
import { StatisticsPanelComponent } from "./statictics/statistics-panel.component";
import { ListProjectROComponent } from "./list-project-ro/list-project-ro.component";
import { ListRepositoryROComponent } from "./list-repository-ro/list-repository-ro.component";
import { NewUserFormComponent } from "./new-user-form/new-user-form.component";
import { InlineAlertComponent } from "./inline-alert/inline-alert.component";
import { PageNotFoundComponent } from "./not-found/not-found.component";
import { AboutDialogComponent } from "./about-dialog/about-dialog.component";
import { GaugeComponent } from "./gauge/gauge.component";
import { ConfirmationDialogComponent } from "./confirmation-dialog/confirmation-dialog.component";
import { ConfirmationDialogService } from "./confirmation-dialog/confirmation-dialog.service";
import { MessageHandlerService } from "./message-handler/message-handler.service";
import { ListChartVersionRoComponent } from "./list-chart-version-ro/list-chart-version-ro.component";
import { IServiceConfig, SERVICE_CONFIG } from "../../lib/entities/service.config";
import { ErrorHandler } from "../../lib/utils/error-handler";
import { HarborLibraryModule } from "../../lib/harbor-library.module";

const uiLibConfig: IServiceConfig = {
  enablei18Support: true,
  langCookieKey: "harbor-lang",
  langMessageLoader: "http",
  langMessagePathForHttpLoader: "i18n/lang/",
  langMessageFileSuffixForHttpLoader: "-lang.json",
  systemInfoEndpoint: "/api/systeminfo",
  repositoryBaseEndpoint: "/api/repositories",
  logBaseEndpoint: "/api/logs",
  targetBaseEndpoint: "/api/registries",
  replicationBaseEndpoint: "/api/replication",
  replicationRuleEndpoint: "/api/replication/policies",
  vulnerabilityScanningBaseEndpoint: "/api/repositories",
  projectPolicyEndpoint: "/api/projects/configs",
  projectBaseEndpoint: "/api/projects",
  localI18nMessageVariableMap: {},
  configurationEndpoint: "/api/configurations",
  scanJobEndpoint: "/api/jobs/scan",
  labelEndpoint: "/api/labels",
  helmChartEndpoint: "/api/chartrepo",
  downloadChartEndpoint: "/chartrepo",
  gcEndpoint: "/api/system/gc",
  ScanAllEndpoint: "/api/system/scanAll",
  quotaUrl: "/api/quotas"
};

@NgModule({
  imports: [
    CoreModule,
    TranslateModule,
    RouterModule,
    FormsModule,
    ReactiveFormsModule,
    HarborLibraryModule.forRoot({
      config: { provide: SERVICE_CONFIG, useValue: uiLibConfig },
      errorHandler: { provide: ErrorHandler, useClass: MessageHandlerService }
    })
  ],
  declarations: [
    MessageComponent,
    MaxLengthExtValidatorDirective,
    ConfirmationDialogComponent,
    NewUserFormComponent,
    InlineAlertComponent,
    PortValidatorDirective,
    PageNotFoundComponent,
    AboutDialogComponent,
    StatisticsComponent,
    StatisticsPanelComponent,
    ListProjectROComponent,
    ListRepositoryROComponent,
    GaugeComponent,
    DateValidatorDirective,
    ListChartVersionRoComponent
  ],
  exports: [
    CoreModule,
    HarborLibraryModule,
    MessageComponent,
    MaxLengthExtValidatorDirective,
    TranslateModule,
    ConfirmationDialogComponent,
    NewUserFormComponent,
    InlineAlertComponent,
    PortValidatorDirective,
    PageNotFoundComponent,
    AboutDialogComponent,
    StatisticsComponent,
    StatisticsPanelComponent,
    ListProjectROComponent,
    ListRepositoryROComponent,
    GaugeComponent,
    DateValidatorDirective,
    FormsModule,
    ReactiveFormsModule,
    ListChartVersionRoComponent
  ],
  providers: [
    SessionService,
    MessageService,
    CookieService,
    ConfirmationDialogService,
    SystemAdminGuard,
    AuthCheckGuard,
    SignInGuard,
    LeavingRepositoryRouteDeactivate,
    LeavingArtifactSummaryRouteDeactivate,
    MemberGuard,
    ArtifactGuard,
    MemberPermissionGuard,
    OidcGuard,
    MessageHandlerService,
    StatisticHandler
  ]
})
export class SharedModule {}
