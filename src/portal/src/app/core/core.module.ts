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
import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { CookieModule } from 'ngx-cookie';
import { MarkdownModule } from 'ngx-markdown';
import { ClarityIconsApi } from '@clr/icons/clr-icons-api';

// ClarityIcons is publicly accessible from the browser's window object.
declare const ClarityIcons: ClarityIconsApi;

// Add custom icons to ClarityIcons
// Add robot head icon
ClarityIcons.add({"robot-head": `
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
<rect class="cls-1" width="36" height="36"/></g></svg>`});

@NgModule({
    imports: [
        BrowserModule,
        FormsModule,
        ClarityModule,
        CookieModule.forRoot(),
        MarkdownModule.forRoot(),
        BrowserAnimationsModule
    ],
    exports: [
        BrowserModule,
        FormsModule,
        ClarityModule,
        BrowserAnimationsModule,
        MarkdownModule
    ]
})
export class CoreModule {
}
