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
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ProjectDetailComponent } from './project-detail.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import {
    BrowserAnimationsModule,
    NoopAnimationsModule,
} from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { of } from 'rxjs';
import { ActivatedRoute } from '@angular/router';
import { SessionService } from '../../../shared/services/session.service';
import { AppConfigService } from '../../../services/app-config.service';
import {
    ProjectService,
    UserPermissionService,
} from '../../../shared/services';
import { ErrorHandler } from '../../../shared/units/error-handler';

describe('ProjectDetailComponent', () => {
    let component: ProjectDetailComponent;
    let fixture: ComponentFixture<ProjectDetailComponent>;
    const mockSessionService = {
        getCurrentUser: () => {
            return of({
                user_id: 1,
            });
        },
    };
    const mockAppConfigService = {
        getConfig: () => {
            return {
                with_admiral: true,
                with_chartmuseum: true,
            };
        },
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    const mockProjectService = null;
    const mockErrorHandler = {
        error() {},
    };
    const mockActivatedRoute = {
        RouterparamMap: of({ get: key => 'value' }),
        snapshot: {
            params: { id: 1 },
            data: 1,
            children: [
                {
                    routeConfig: { path: '' },
                },
            ],
        },
        data: of({
            projectResolver: {
                ismember: true,
                role_name: 'maintainer',
            },
        }),
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HttpClientTestingModule,
            ],
            declarations: [ProjectDetailComponent],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: mockSessionService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                {
                    provide: UserPermissionService,
                    useValue: mockUserPermissionService,
                },
                { provide: ErrorHandler, useValue: mockErrorHandler },
                { provide: ProjectService, useValue: mockProjectService },
                {
                    provide: ActivatedRoute,
                    useValue: mockActivatedRoute,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectDetailComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
