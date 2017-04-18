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
import { NgModule } from '@angular/core';
import { CoreModule } from '../core/core.module';
import { CookieService } from 'angular2-cookie/core';

import { SessionService } from '../shared/session.service';
import { MessageComponent } from '../global-message/message.component';

import { MessageService } from '../global-message/message.service';
import { MaxLengthExtValidatorDirective } from './max-length-ext.directive';
import { FilterComponent } from './filter/filter.component';
import { TranslateModule } from "@ngx-translate/core";

import { RouterModule } from '@angular/router';

import { ConfirmationDialogComponent } from './confirmation-dialog/confirmation-dialog.component';
import { ConfirmationDialogService } from './confirmation-dialog/confirmation-dialog.service';
import { SystemAdminGuard } from './route/system-admin-activate.service';
import { NewUserFormComponent } from './new-user-form/new-user-form.component';
import { InlineAlertComponent } from './inline-alert/inline-alert.component';

import { ListPolicyComponent } from './list-policy/list-policy.component';
import { CreateEditPolicyComponent } from './create-edit-policy/create-edit-policy.component';

import { PortValidatorDirective } from './port.directive';

import { PageNotFoundComponent } from './not-found/not-found.component';
import { AboutDialogComponent } from './about-dialog/about-dialog.component';

import { AuthCheckGuard } from './route/auth-user-activate.service';

import { StatisticsComponent } from './statictics/statistics.component';
import { StatisticsPanelComponent } from './statictics/statistics-panel.component';
import { SignInGuard } from './route/sign-in-guard-activate.service';
import { LeavingConfigRouteDeactivate } from './route/leaving-config-deactivate.service';
import { MemberGuard } from './route/member-guard-activate.service';

import { ListProjectROComponent } from './list-project-ro/list-project-ro.component';
import { ListRepositoryROComponent } from './list-repository-ro/list-repository-ro.component';

import { MessageHandlerService } from './message-handler/message-handler.service';
import { EmailValidatorDirective } from './email.directive';
import { GaugeComponent } from './gauge/gauge.component';
import { StatisticHandler } from './statictics/statistic-handler.service';

@NgModule({
  imports: [
    CoreModule,
    TranslateModule,
    RouterModule
  ],
  declarations: [
    MessageComponent,
    MaxLengthExtValidatorDirective,
    FilterComponent,
    ConfirmationDialogComponent,
    NewUserFormComponent,
    InlineAlertComponent,
    ListPolicyComponent,
    CreateEditPolicyComponent,
    PortValidatorDirective,
    PageNotFoundComponent,
    AboutDialogComponent,
    StatisticsComponent,
    StatisticsPanelComponent,
    ListProjectROComponent,
    ListRepositoryROComponent,
    EmailValidatorDirective,
    GaugeComponent
  ],
  exports: [
    CoreModule,
    MessageComponent,
    MaxLengthExtValidatorDirective,
    FilterComponent,
    TranslateModule,
    ConfirmationDialogComponent,
    NewUserFormComponent,
    InlineAlertComponent,
    ListPolicyComponent,
    CreateEditPolicyComponent,
    PortValidatorDirective,
    PageNotFoundComponent,
    AboutDialogComponent,
    StatisticsComponent,
    StatisticsPanelComponent,
    ListProjectROComponent,
    ListRepositoryROComponent,
    EmailValidatorDirective,
    GaugeComponent
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
    MemberGuard,
    MessageHandlerService,
    StatisticHandler
  ]
})
export class SharedModule {

}