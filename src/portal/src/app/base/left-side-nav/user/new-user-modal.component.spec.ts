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
import { ClarityModule } from '@clr/angular';
import { TranslateModule } from '@ngx-translate/core';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { UserService } from './user.service';
import { SessionService } from '../../../shared/services/session.service';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { SharedTestingModule } from '../../../shared/shared.module';
import { NewUserModalComponent } from './new-user-modal.component';

describe('NewUserModalComponent', () => {
    let component: NewUserModalComponent;
    let fixture: ComponentFixture<NewUserModalComponent>;
    let fakeSessionService = null;
    let fakeUserService = null;
    let fakeMessageHandlerService = {
        handleError: function () {},
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [NewUserModalComponent],
            imports: [
                ClarityModule,
                SharedTestingModule,
                TranslateModule.forRoot(),
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            providers: [
                {
                    provide: MessageHandlerService,
                    useValue: fakeMessageHandlerService,
                },
                { provide: UserService, useValue: fakeUserService },
                { provide: SessionService, useValue: fakeSessionService },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(NewUserModalComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
