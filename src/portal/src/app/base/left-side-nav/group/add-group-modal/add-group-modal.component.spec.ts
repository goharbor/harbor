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
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { SessionService } from '../../../../shared/services/session.service';
import { AppConfigService } from '../../../../services/app-config.service';
import { AddGroupModalComponent } from './add-group-modal.component';
import { UsergroupService } from '../../../../../../ng-swagger-gen/services/usergroup.service';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('AddGroupModalComponent', () => {
    let component: AddGroupModalComponent;
    let fixture: ComponentFixture<AddGroupModalComponent>;
    let fakeSessionService = {
        getCurrentUser: function () {
            return { has_admin_role: true };
        },
    };
    let fakeGroupService = null;
    let fakeAppConfigService = {
        isLdapMode: function () {
            return true;
        },
        isHttpAuthMode: function () {
            return false;
        },
        isOidcMode: function () {
            return false;
        },
    };
    let fakeMessageHandlerService = null;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [AddGroupModalComponent],
            imports: [SharedTestingModule],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                {
                    provide: MessageHandlerService,
                    useValue: fakeMessageHandlerService,
                },
                { provide: SessionService, useValue: fakeSessionService },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                { provide: UsergroupService, useValue: fakeGroupService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddGroupModalComponent);
        component = fixture.componentInstance;
        component.open();
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
