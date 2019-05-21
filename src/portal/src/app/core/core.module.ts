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
