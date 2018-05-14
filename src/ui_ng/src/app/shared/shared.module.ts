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
import { CookieService } from "ngx-cookie";
import {
  IServiceConfig,
  SERVICE_CONFIG,
  ErrorHandler,
  HarborLibraryModule
} from "harbor-ui";

import { SessionService } from "../shared/session.service";
import { MessageService } from "../global-message/message.service";
import { MessageComponent } from "../global-message/message.component";
import { DateValidatorDirective } from "../shared/date-validator.directive";
import { CoreModule } from "../core/core.module";

import { AuthCheckGuard } from "./route/auth-user-activate.service";
import { SignInGuard } from "./route/sign-in-guard-activate.service";
import { SystemAdminGuard } from "./route/system-admin-activate.service";
import { MemberGuard } from "./route/member-guard-activate.service";
import { LeavingConfigRouteDeactivate } from "./route/leaving-config-deactivate.service";
import { LeavingRepositoryRouteDeactivate } from "./route/leaving-repository-deactivate.service";

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

const uiLibConfig: IServiceConfig = {
  enablei18Support: true,
  langCookieKey: "harbor-lang",
  langMessageLoader: "http",
  langMessagePathForHttpLoader: "i18n/lang/",
  langMessageFileSuffixForHttpLoader: "-lang.json"
};

@NgModule({
  imports: [
    CoreModule,
    TranslateModule,
    RouterModule,
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
    DateValidatorDirective
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
    DateValidatorDirective
  ],
  providers: [
    SessionService,
    MessageService,
    CookieService,
    ConfirmationDialogService,
    SystemAdminGuard,
    AuthCheckGuard,
    SignInGuard,
    LeavingConfigRouteDeactivate,
    LeavingRepositoryRouteDeactivate,
    MemberGuard,
    MessageHandlerService,
    StatisticHandler
  ]
})
export class SharedModule {}
