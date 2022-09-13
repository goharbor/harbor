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
import { RouterModule } from '@angular/router';
import { TranslateModule, TranslateStore } from '@ngx-translate/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { DateValidatorDirective } from './directives/date-validator.directive';
import { PortValidatorDirective } from './directives/port.directive';
import { MaxLengthExtValidatorDirective } from './directives/max-length-ext.directive';
import { ErrorHandler } from './units/error-handler';
import { ClarityIconsApi } from '@clr/icons/clr-icons-api';
import { ClarityModule } from '@clr/angular';
import { MarkdownModule } from 'ngx-markdown';
import { CommonModule } from '@angular/common';
import { ClipboardModule } from './components/third-party/ngx-clipboard';
import { InlineAlertComponent } from './components/inline-alert/inline-alert.component';
import { NewUserFormComponent } from './components/new-user-form/new-user-form.component';
import { MessageComponent } from './components/global-message/message.component';
import { NavigatorComponent } from './components/navigator/navigator.component';
import { SearchResultComponent } from './components/global-search/search-result.component';
import { GlobalSearchComponent } from './components/global-search/global-search.component';
import { AboutDialogComponent } from './components/about-dialog/about-dialog.component';
import {
    LabelDefaultService,
    LabelService,
    ProjectDefaultService,
    ProjectService,
    ReplicationDefaultService,
    ReplicationService,
    ScanningResultDefaultService,
    ScanningResultService,
    SystemInfoDefaultService,
    SystemInfoService,
    UserPermissionDefaultService,
    UserPermissionService,
} from './services';
import { FilterComponent } from './components/filter/filter.component';
import { GaugeComponent } from './components/gauge/gauge.component';
import { ConfirmationDialogComponent } from './components/confirmation-dialog';
import { ListRepositoryROComponent } from './components/list-repository-ro/list-repository-ro.component';
import { OperationComponent } from './components/operation/operation.component';
import { ViewTokenComponent } from './components/view-token/view-token.component';
import { PushImageButtonComponent } from './components/push-image/push-image.component';
import { CopyInputComponent } from './components/push-image/copy-input.component';
import { ListProjectROComponent } from './components/list-project-ro/list-project-ro.component';
import {
    CronScheduleComponent,
    CronTooltipComponent,
} from './components/cron-schedule';
import { LabelComponent } from './components/label/label.component';
import { LabelSignPostComponent } from './components/label/label-signpost/label-signpost.component';
import { LabelPieceComponent } from './components/label/label-piece/label-piece.component';
import { CreateEditLabelComponent } from './components/label/create-edit-label/create-edit-label.component';
import { ListChartVersionRoComponent } from './components/list-chart-version-ro/list-chart-version-ro.component';
import { DatePickerComponent } from './components/datetime-picker/datetime-picker.component';
import {
    EndpointDefaultService,
    EndpointService,
} from './services/endpoint.service';
import { ImageNameInputComponent } from './components/image-name-input/image-name-input.component';
import {
    HelmChartDefaultService,
    HelmChartService,
} from '../base/project/helm-chart/helm-chart-detail/helm-chart.service';
import { MessageHandlerService } from './services/message-handler.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HarborDatetimePipe } from './pipes/harbor-datetime.pipe';
import { RemainingTimeComponent } from './components/remaining-time/remaining-time.component';

import { registerLocaleData } from '@angular/common';
import locale_en from '@angular/common/locales/en';
import locale_zh_CN from '@angular/common/locales/zh-Hans';
import locale_zh_TW from '@angular/common/locales/zh-Hans-HK';
import locale_es from '@angular/common/locales/es';
import locale_fr from '@angular/common/locales/fr';
import locale_pt from '@angular/common/locales/pt-PT';
import locale_tr from '@angular/common/locales/tr';
import locale_de from '@angular/common/locales/de';
import { SupportedLanguage } from './entities/shared.const';

const localesForSupportedLangs: Record<SupportedLanguage, unknown[]> = {
    'en-us': locale_en,
    'zh-cn': locale_zh_CN,
    'zh-tw': locale_zh_TW,
    'es-es': locale_es,
    'fr-fr': locale_fr,
    'pt-br': locale_pt,
    'tr-tr': locale_tr,
    'de-de': locale_de,
};
for (const [lang, locale] of Object.entries(localesForSupportedLangs)) {
    registerLocaleData(locale, lang);
}

// ClarityIcons is publicly accessible from the browser's window object.
declare const ClarityIcons: ClarityIconsApi;

// Add custom icons to ClarityIcons
// Add robot head icon
ClarityIcons.add({
    'robot-head': `
<svg xmlns="http://www.w3.org/2000/svg" width="36" height="36" viewBox="0 0 36 36">
<defs><style>.cls-1{fill:none;}</style></defs><g id="Layer_2" data-name="Layer 2">
<circle cx="12.62" cy="18.6" r="1.5"/><circle cx="23.5" cy="18.5" r="1.5"/>
<path d="M22,28H14a1,1,0,0,1,0-2h8a1,1,0,0,1,0,2Z"/>
<path d="M35,25.22a1,1,0,0,1-1-1V19.38a1,1,0,1,1,2,0v4.84A1,1,0,0,1,35,25.22Z"/>
<path d="M1,25a1,1,0,0,1-1-1V19a1,1,0,0,1,2,0v5A1,1,0,0,1,1,25Z"/>
<path d="M19,8.26A3.26,3.26,0,1,1,22.26,5,3.26,3.26,0,0,1,19,8.26Zm0-4.92A1.66,1.66,0,1,0,20.66,5,1.67,1.67,0,0,0,19,3.34Z"/>
<path d="M29.1,10.49a1,1,0,0,0-.86-.49H20V7.58H18V12h9.67A19.51,19.51,0,0,1,30,21.42,
19.06,19.06,0,0,1,27,32H9.05A19.06,19.06,0,0,1,6,21.42,
19.51,19.51,0,0,1,8.33,12H16V10H7.76a1,1,0,0,0-.86.49A21.18,
21.18,0,0,0,4,21.42,21,21,0,0,0,7.71,33.58a1,1,0,0,0,.81.42h19a1,1,0,0,0,
.81-.42A21,21,0,0,0,32,21.42,21.18,21.18,0,0,0,29.1,10.49Z"/>
<rect class="cls-1" width="36" height="36"/></g></svg>`,
});

@NgModule({
    imports: [
        TranslateModule.forChild({
            extend: true,
        }),
        FormsModule,
        CommonModule,
        ClarityModule,
        MarkdownModule.forRoot(),
        RouterModule,
        ReactiveFormsModule,
        ClipboardModule,
    ],
    declarations: [
        MaxLengthExtValidatorDirective,
        PortValidatorDirective,
        DateValidatorDirective,
        InlineAlertComponent,
        NewUserFormComponent,
        MessageComponent,
        NavigatorComponent,
        SearchResultComponent,
        GlobalSearchComponent,
        AboutDialogComponent,
        FilterComponent,
        GaugeComponent,
        ConfirmationDialogComponent,
        ListRepositoryROComponent,
        OperationComponent,
        ViewTokenComponent,
        PushImageButtonComponent,
        CopyInputComponent,
        ListProjectROComponent,
        CronTooltipComponent,
        LabelComponent,
        LabelSignPostComponent,
        LabelPieceComponent,
        CreateEditLabelComponent,
        CronScheduleComponent,
        ListChartVersionRoComponent,
        DatePickerComponent,
        ImageNameInputComponent,
        HarborDatetimePipe,
        RemainingTimeComponent,
    ],
    exports: [
        TranslateModule,
        FormsModule,
        ReactiveFormsModule,
        ClarityModule,
        CommonModule,
        ClipboardModule,
        MarkdownModule,
        MaxLengthExtValidatorDirective,
        PortValidatorDirective,
        DateValidatorDirective,
        InlineAlertComponent,
        NewUserFormComponent,
        MessageComponent,
        NavigatorComponent,
        SearchResultComponent,
        GlobalSearchComponent,
        AboutDialogComponent,
        FilterComponent,
        GaugeComponent,
        ConfirmationDialogComponent,
        ListRepositoryROComponent,
        OperationComponent,
        ViewTokenComponent,
        PushImageButtonComponent,
        CopyInputComponent,
        ListProjectROComponent,
        CronTooltipComponent,
        LabelComponent,
        LabelSignPostComponent,
        LabelPieceComponent,
        CreateEditLabelComponent,
        CronScheduleComponent,
        ListChartVersionRoComponent,
        DatePickerComponent,
        ImageNameInputComponent,
        HarborDatetimePipe,
        RemainingTimeComponent,
    ],
    providers: [
        { provide: EndpointService, useClass: EndpointDefaultService },
        { provide: ReplicationService, useClass: ReplicationDefaultService },
        { provide: LabelService, useClass: LabelDefaultService },
        { provide: SystemInfoService, useClass: SystemInfoDefaultService },
        {
            provide: ScanningResultService,
            useClass: ScanningResultDefaultService,
        },
        { provide: HelmChartService, useClass: HelmChartDefaultService },
    ],
})
export class SharedModule {}

// this module is only for testing, you should only import this module in *.spec.ts files
@NgModule({
    imports: [
        BrowserAnimationsModule,
        SharedModule,
        HttpClientTestingModule,
        RouterTestingModule,
    ],
    exports: [
        BrowserAnimationsModule,
        SharedModule,
        HttpClientTestingModule,
        RouterTestingModule,
    ],
    providers: [
        TranslateStore,
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: ErrorHandler, useClass: MessageHandlerService },
        {
            provide: UserPermissionService,
            useClass: UserPermissionDefaultService,
        },
    ],
})
export class SharedTestingModule {}
