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
import {
    ComponentFixture,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { AddRoleComponent } from './add-role.component';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { of, throwError } from 'rxjs';
import { Role } from '../../../../../../ng-swagger-gen/models/role';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { delay } from 'rxjs/operators';
import { RoleService } from '../../../../../../ng-swagger-gen/services/role.service';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { SharedTestingModule } from '../../../../shared/shared.module';

describe('AddRoleComponent', () => {
    let component: AddRoleComponent;
    let fixture: ComponentFixture<AddRoleComponent>;
    const existingRoles: Role[] = [
        { id: 2, name: 'developer', is_builtin: true },
        { id: 7, name: 'release,prod', is_builtin: false },
    ];
    let listRoleCalls: number;
    let listRoleFails: boolean;
    const fakedRoleService = {
        ListRole() {
            return of([]).pipe(delay(0));
        },
        ListRoleResponse() {
            listRoleCalls++;
            if (listRoleFails) {
                return throwError(() => new Error('boom'));
            }
            return of(
                new HttpResponse<Array<Role>>({
                    headers: new HttpHeaders({
                        'x-total-count': `${existingRoles.length}`,
                    }),
                    body: existingRoles,
                })
            );
        },
    };
    const fakedMessageHandlerService = {
        showSuccess() {},
        error() {},
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [AddRoleComponent],
            imports: [SharedTestingModule],
            providers: [
                OperationService,
                { provide: RoleService, useValue: fakedRoleService },
                {
                    provide: MessageHandlerService,
                    useValue: fakedMessageHandlerService,
                },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        listRoleCalls = 0;
        listRoleFails = false;
        fixture = TestBed.createComponent(AddRoleComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    describe('name check', () => {
        function type(name: string) {
            component.role.name = name;
            component.inputName();
            tick(500);
        }

        it('flags a name that an existing role already uses', fakeAsync(() => {
            type('developer');
            expect(component.isNameExisting).toBeTrue();
        }));

        it('accepts a name that is only a prefix of an existing one', fakeAsync(() => {
            type('dev');
            expect(component.isNameExisting).toBeFalse();
        }));

        it('accepts a name holding characters that a q filter would not survive', fakeAsync(() => {
            type('release,pro');
            expect(component.isNameExisting).toBeFalse();
            type('release,prod');
            expect(component.isNameExisting).toBeTrue();
        }));

        it('keeps checking later names after a failed request', fakeAsync(() => {
            listRoleFails = true;
            type('whatever');
            expect(component.isNameExisting).toBeFalse();
            listRoleFails = false;
            type('developer');
            expect(component.isNameExisting).toBeTrue();
            expect(listRoleCalls).toBe(2);
        }));
    });
});
