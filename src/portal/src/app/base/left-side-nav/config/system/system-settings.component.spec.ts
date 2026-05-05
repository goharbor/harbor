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
import { SystemSettingsComponent } from './system-settings.component';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { delay, of, Subscription } from 'rxjs';
import { Configuration } from '../config';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { ConfigService } from '../config.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { AuditlogService } from 'ng-swagger-gen/services';
import { HttpHeaders, HttpResponse } from '@angular/common/http';

describe('SystemSettingsComponent', () => {
    let component: SystemSettingsComponent;
    let fixture: ComponentFixture<SystemSettingsComponent>;
    const fakeConfigService = {
        config: new Configuration(),
        getConfig() {
            return this.config;
        },
        setConfig(c) {
            this.config = c;
        },
        getOriginalConfig() {
            return new Configuration();
        },
        getLoadingConfigStatus() {
            return false;
        },
        confirmUnsavedChanges() {},
        updateConfig() {},
        resetConfig() {},
        saveConfiguration() {
            return of(null);
        },
    };
    const fakedAppConfigService = {
        getConfig() {
            return {};
        },
        load() {
            return of(null);
        },
    };
    const mockedAuditLogs = [
        {
            event_type: 'create_artifact',
        },
        {
            event_type: 'delete_artifact',
        },
        {
            event_type: 'pull_artifact',
        },
    ];
    const fakeAuditlogService = {
        listAuditLogEventTypesResponse() {
            return of(
                new HttpResponse({
                    body: mockedAuditLogs,
                    headers: new HttpHeaders({
                        'x-total-count': '18',
                    }),
                })
            ).pipe(delay(0));
        },
    };
    const fakedErrorHandler = {
        error() {
            return undefined;
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [
                { provide: AppConfigService, useValue: fakedAppConfigService },
                { provide: ErrorHandler, useValue: fakedErrorHandler },
                { provide: ConfigService, useValue: fakeConfigService },
                { provide: AuditlogService, useValue: fakeAuditlogService },
            ],
            declarations: [SystemSettingsComponent],
        }).compileComponents();
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(SystemSettingsComponent);
        component = fixture.componentInstance;
        component.selectedLogEventTypes = ['create_artifact'];
        fixture.autoDetectChanges(true);
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('cancel button should work', () => {
        const spy: jasmine.Spy = spyOn(component, 'cancel').and.returnValue(
            undefined
        );
        const cancel: HTMLButtonElement = fixture.nativeElement.querySelector(
            '#config_system_cancel'
        );
        cancel.dispatchEvent(new Event('click'));
        expect(spy.calls.count()).toEqual(1);
    });
    it('save button should work', () => {
        const input = fixture.nativeElement.querySelector('#robotNamePrefix');
        input.value = 'test';
        input.dispatchEvent(new Event('input'));
        const save: HTMLButtonElement = fixture.nativeElement.querySelector(
            '#config_system_save'
        );
        save.dispatchEvent(new Event('click'));
        expect(input.value).toEqual('test');
    });
});
