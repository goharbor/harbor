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
/* tslint:disable:no-unused-variable */

import { TestBed, async, ComponentFixture } from '@angular/core/testing';
import { Title } from '@angular/platform-browser';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CookieService } from 'ngx-cookie';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { SessionService } from './shared/session.service';
import { AppConfigService } from './services/app-config.service';
import { AppComponent } from './app.component';
import { ClarityModule } from "@clr/angular";
import { APP_BASE_HREF } from "@angular/common";

describe('AppComponent', () => {
    let fixture: ComponentFixture<any>;
    let compiled: any;
    let fakeCookieService = null;
    let fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        }
    };
    let fakeAppConfigService = {
        isIntegrationMode: function () {
            return true;
        }
    };
    let fakeTitle = {
        setTitle: function () {
        }
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [
                AppComponent
            ],
            imports: [
                ClarityModule,
                TranslateModule.forRoot()
            ],
            providers: [
                TranslateService,
                { provide: APP_BASE_HREF, useValue: '/' },
                { provide: CookieService, useValue: fakeCookieService },
                { provide: SessionService, useValue: fakeSessionService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: Title, useValue: fakeTitle },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA]
        });

        fixture = TestBed.createComponent(AppComponent);
        fixture.detectChanges();
        compiled = fixture.nativeElement;


    });

    afterEach(() => {
        fixture.destroy();
    });

    it('should create the app', async(() => {
        expect(compiled).toBeTruthy();
    }));


});
