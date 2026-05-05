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
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { ConfigurationService } from '../../../../services/config.service';
import { ConfigurationAuthComponent } from './config-auth.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { SystemInfoService } from '../../../../shared/services';
import { ConfigService } from '../config.service';
import { Configuration } from '../config';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('ConfigurationAuthComponent', () => {
    let component: ConfigurationAuthComponent;
    let fixture: ComponentFixture<ConfigurationAuthComponent>;
    let fakeMessageHandlerService = {
        showSuccess: () => null,
    };
    let fakeConfigurationService = {
        saveConfiguration: () => of(null),
        testLDAPServer: () => of(null),
        testOIDCServer: () => of(null),
    };
    let fakeAppConfigService = {
        load: () => of(null),
    };
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
        updateConfig() {},
        resetConfig() {},
    };
    let fakeSystemInfoService = {
        getSystemInfo: function () {
            return of({
                external_url: 'expectedUrl',
            });
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ConfigurationAuthComponent],
            providers: [
                {
                    provide: MessageHandlerService,
                    useValue: fakeMessageHandlerService,
                },
                {
                    provide: ConfigurationService,
                    useValue: fakeConfigurationService,
                },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: ConfigService, useValue: fakeConfigService },
                { provide: SystemInfoService, useValue: fakeSystemInfoService },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationAuthComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
