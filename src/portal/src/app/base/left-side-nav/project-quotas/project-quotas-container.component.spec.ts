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
import { ProjectQuotasContainerComponent } from './project-quotas-container.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { SessionService } from '../../../shared/services/session.service';
import { SessionUser } from '../../../shared/entities/session-user';
import { ConfigurationService } from '../../../services/config.service';
import { of } from 'rxjs';
import { Configuration } from '../config/config';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('ProjectQuotasContainerComponent', () => {
    let component: ProjectQuotasContainerComponent;
    let fixture: ComponentFixture<ProjectQuotasContainerComponent>;
    const mockedUser: SessionUser = {
        user_id: 1,
        username: 'admin',
        email: 'harbor@vmware.com',
        realname: 'admin',
        has_admin_role: true,
        comment: 'no comment',
    };
    let mockedConfig: Configuration = new Configuration();
    mockedConfig.count_per_project.value = 10;
    const fakedSessionService = {
        getCurrentUser() {
            return mockedUser;
        },
    };
    const fakedConfigurationService = {
        getConfiguration() {
            return of(mockedConfig);
        },
    };
    const fakedMessageHandlerService = {
        handleError() {
            return;
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ProjectQuotasContainerComponent],
            providers: [
                {
                    provide: MessageHandlerService,
                    useValue: fakedMessageHandlerService,
                },
                { provide: SessionService, useValue: fakedSessionService },
                {
                    provide: ConfigurationService,
                    useValue: fakedConfigurationService,
                },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectQuotasContainerComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should get config', () => {
        expect(component.allConfig.count_per_project.value).toEqual(10);
    });
});
