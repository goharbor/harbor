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
import { AddMemberComponent } from './add-member.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { UserService } from '../../../left-side-nav/user/user.service';
import { of } from 'rxjs';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { ActivatedRoute } from '@angular/router';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { MemberService } from 'ng-swagger-gen/services/member.service';

describe('AddMemberComponent', () => {
    let component: AddMemberComponent;
    let fixture: ComponentFixture<AddMemberComponent>;
    const mockMemberService = {
        listProjectMembers: () => {
            return of([]);
        },
    };
    const mockUserService = {
        searchUsers: () => {
            return of([[], []]);
        },
    };

    const mockMessageHandlerService = {
        showSuccess: () => {},
        handleError: () => {},
        isAppLevel: () => {},
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [AddMemberComponent],
            providers: [
                { provide: MemberService, useValue: mockMemberService },
                { provide: UserService, useValue: mockUserService },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
                {
                    provide: ActivatedRoute,
                    useValue: {
                        RouterparamMap: of({ get: key => 'value' }),
                        snapshot: {
                            parent: {
                                parent: {
                                    params: { id: 1 },
                                    data: null,
                                },
                            },
                        },
                    },
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddMemberComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('groups mixed roles and allows selection across categories', async () => {
        const guest = { id: 3, name: 'guest', is_builtin: true };
        const custom = { id: 10, name: 'Audit', is_builtin: false };
        component.builtinRoles = [guest];
        component.customRoles = [custom];
        component.rolesLoaded = true;
        component.addMemberOpened = true;
        fixture.detectChanges();
        await fixture.whenStable();

        const groups = fixture.nativeElement.querySelectorAll(
            '.role-radio-list fieldset'
        );
        expect(groups.length).toBe(2);
        const builtinRadio = groups[0].querySelector('input');
        const customRadio = groups[1].querySelector('input');
        expect(builtinRadio.id).toBe('add-member-role-3');
        expect(customRadio.id).toBe('add-member-role-10');
        builtinRadio.click();
        customRadio.click();
        expect(component.roleId).toBe(custom.id);
        expect(builtinRadio.checked).toBeFalse();
        expect(component.hasValidRole()).toBeTrue();

        component.customRoles = [];
        fixture.detectChanges();
        expect(fixture.nativeElement.querySelectorAll('fieldset').length).toBe(
            1
        );
        expect(component.hasValidRole()).toBeFalse();
    });
});
