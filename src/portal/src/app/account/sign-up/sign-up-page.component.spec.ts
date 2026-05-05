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
import { UserService } from '../../base/left-side-nav/user/user.service';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { MessageService } from '../../shared/components/global-message/message.service';
import { RouterTestingModule } from '@angular/router/testing';
import { SignUpPageComponent } from './sign-up-page.component';
import { FormsModule } from '@angular/forms';
import { NewUserFormComponent } from '../../shared/components/new-user-form/new-user-form.component';
import { SessionService } from '../../shared/services/session.service';

describe('SignUpPageComponent', () => {
    let component: SignUpPageComponent;
    let fixture: ComponentFixture<SignUpPageComponent>;
    let fakeUserService = null;
    let fakeSessionService = null;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [SignUpPageComponent, NewUserFormComponent],
            imports: [
                FormsModule,
                RouterTestingModule,
                TranslateModule.forRoot(),
            ],
            providers: [
                MessageService,
                TranslateService,
                { provide: UserService, useValue: fakeUserService },
                { provide: SessionService, useValue: fakeSessionService },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(SignUpPageComponent);
        component = fixture.componentInstance;
        component.newUserForm =
            TestBed.createComponent(NewUserFormComponent).componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
});
