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
import { DatePickerComponent } from './components/datetime-picker/datetime-picker.component';
import {
    EndpointDefaultService,
    EndpointService,
} from './services/endpoint.service';
import { ImageNameInputComponent } from './components/image-name-input/image-name-input.component';
import { MessageHandlerService } from './services/message-handler.service';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HarborDatetimePipe } from './pipes/harbor-datetime.pipe';
import { RemainingTimeComponent } from './components/remaining-time/remaining-time.component';
import { LabelSelectorComponent } from './components/label-selector/label-selector.component';
import { ScrollSectionDirective } from './directives/scroll/scroll-section.directive';
import { ScrollAnchorDirective } from './directives/scroll/scroll-anchor.directive';
import { AppLevelAlertsComponent } from './components/app-level-alerts/app-level-alerts.component';
// import echarts
import * as echarts from 'echarts/core';
import { PieChart } from 'echarts/charts';
import {
    TitleComponent,
    TooltipComponent,
    GridComponent,
    DatasetComponent,
    TransformComponent,
    LegendComponent,
} from 'echarts/components';
import { LabelLayout, UniversalTransition } from 'echarts/features';
import { CanvasRenderer } from 'echarts/renderers';
import { RobotPermissionsPanelComponent } from './components/robot-permissions-panel/robot-permissions-panel.component';
import { PreferenceSettingsComponent } from '../base/preference-settings/preference-settings.component';

// register necessary components
echarts.use([
    TitleComponent,
    TooltipComponent,
    GridComponent,
    DatasetComponent,
    TransformComponent,
    PieChart,
    LabelLayout,
    UniversalTransition,
    CanvasRenderer,
    LegendComponent,
]);

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
    sbom: `
<?xml version='1.0' encoding='utf-8'?>
<!-- Generator: imaengine 6.0   -->
<svg xmlns:xlink="http://www.w3.org/1999/xlink" xmlns="http://www.w3.org/2000/svg" viewBox="0,0,512,512" style="enable-background:new 0 0 512 512;" version="1.1">
<defs/>
<g id="layer0">
<g transform="matrix(1 0 0 1 0 0)">
<path d="M341.333,213.333L362.666,213.333L362.666,117.333C362.619,116.239 362.399,115.159 362.015,114.133C361.93,113.866 361.951,113.557 361.844,113.29C361.285,111.937 360.454,110.713 359.401,109.695L252.885,3.136C250.88,1.135 248.166,0.0079999 245.333,0L10.667,0C4.776,0 0,4.776 0,10.667L0,458.667C0,464.558 4.776,469.334 10.667,469.334L224,469.334L224,448L21.333,448L21.333,21.333L234.666,21.333L234.666,117.333C234.666,123.224 239.442,128 245.333,128L341.333,128L341.333,213.333L341.333,213.333ZM256,106.667L256,36.427L326.219,106.667L256,106.667L256,106.667Z" fill="#175975"/>
<path d="M501.333,341.333L480,341.333L480,330.666C479.988,289.429 446.55,256.009 405.312,256.02C402.195,256.021 399.081,256.217 395.989,256.607C358.752,261.151 330.666,294.047 330.666,333.119L330.666,341.332L309.333,341.332C303.442,341.332 298.666,346.108 298.666,351.999L298.666,501.332C298.666,507.223 303.442,511.999 309.333,511.999L501.333,511.999C507.224,511.999 512,507.223 512,501.332L512,352C512,346.109 507.224,341.333 501.333,341.333L501.333,341.333ZM352,333.131C352,304.822 372.021,281.035 398.571,277.792C427.788,274.057 454.502,294.715 458.237,323.932C458.523,326.165 458.666,328.415 458.667,330.666L458.667,341.333L352,341.333L352,333.131L352,333.131ZM490.667,490.667L320,490.667L320,362.667L490.667,362.667L490.667,490.667L490.667,490.667Z" fill="#175975"/>
<path d="M394.667,423.797L394.667,458.666C394.667,464.557 399.443,469.333 405.334,469.333C411.225,469.333 416,464.558 416,458.667L416,423.563C426.103,417.57 429.435,404.522 423.443,394.419C419.605,387.948 412.633,383.986 405.11,384C393.369,383.979 383.834,393.479 383.813,405.22C383.798,412.922 387.951,420.028 394.667,423.797L394.667,423.797Z" fill="#175975"/>
</g>
<text font-size="100" font-family="'MesloLGLDZForPowerline-Bold'" fill="#175975" transform="matrix(1.24404 0 0 1.04972 34.5897 178.002)">
<tspan x="0" y="112" textLength="240.82">
<![CDATA[SBOM]]></tspan>
</text>
</g>
</svg>`,
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
        ScrollSectionDirective,
        ScrollAnchorDirective,
        InlineAlertComponent,
        NewUserFormComponent,
        MessageComponent,
        PreferenceSettingsComponent,
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
        DatePickerComponent,
        ImageNameInputComponent,
        HarborDatetimePipe,
        RemainingTimeComponent,
        LabelSelectorComponent,
        AppLevelAlertsComponent,
        RobotPermissionsPanelComponent,
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
        ScrollSectionDirective,
        ScrollAnchorDirective,
        InlineAlertComponent,
        NewUserFormComponent,
        MessageComponent,
        PreferenceSettingsComponent,
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
        DatePickerComponent,
        ImageNameInputComponent,
        HarborDatetimePipe,
        RemainingTimeComponent,
        LabelSelectorComponent,
        AppLevelAlertsComponent,
        RobotPermissionsPanelComponent,
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
