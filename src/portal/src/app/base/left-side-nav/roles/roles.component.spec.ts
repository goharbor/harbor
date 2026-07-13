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
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { RolesComponent } from './roles.component';
import { RoleService } from '../../../../../ng-swagger-gen/services/role.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { of, Subscription } from 'rxjs';
import { delay } from 'rxjs/operators';
import { Role } from '../../../../../ng-swagger-gen/models/role';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CommonModule } from '@angular/common';
import { ClarityModule } from '@clr/angular';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { SysteminfoService } from '../../../../../ng-swagger-gen/services/systeminfo.service';
import { PermissionsService } from '../../../../../ng-swagger-gen/services/permissions.service';

const builtinRole: Role = {
    id: 1,
    name: 'projectAdmin',
    is_builtin: true,
    permissions: [],
};

const customRole: Role = {
    id: 10,
    name: 'myCustomRole',
    is_builtin: false,
    permissions: [],
};

const fakedRoleService = {
    ListRoleResponse() {
        const res: HttpResponse<Array<Role>> = new HttpResponse<Array<Role>>({
            headers: new HttpHeaders({ 'x-total-count': '2' }),
            body: [builtinRole, customRole],
        });
        return of(res).pipe(delay(0));
    },
};

const fakedMessageHandlerService = { showSuccess() {}, error() {} };
const fakedSystemInfoService = { getSystemInfo: () => of({}) };
const fakedPermissionsService = { getPermissions: () => of({}) };

describe('RolesComponent', () => {
    let component: RolesComponent;
    let fixture: ComponentFixture<RolesComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                TranslateModule.forRoot(),
                CommonModule,
                ClarityModule,
                HttpClientTestingModule,
                RouterTestingModule,
                BrowserAnimationsModule,
            ],
            declarations: [RolesComponent],
            providers: [
                TranslateService,
                ConfirmationDialogService,
                OperationService,
                {
                    provide: MessageHandlerService,
                    useValue: fakedMessageHandlerService,
                },
                { provide: RoleService, useValue: fakedRoleService },
                {
                    provide: SysteminfoService,
                    useValue: fakedSystemInfoService,
                },
                {
                    provide: PermissionsService,
                    useValue: fakedPermissionsService,
                },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(RolesComponent);
        component = fixture.componentInstance;
        component.searchSub = new Subscription();
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    describe('isBuiltinSelected', () => {
        it('returns false when nothing is selected', () => {
            component.selectedRows = [];
            expect(component.isBuiltinSelected()).toBeFalse();
        });

        it('returns false when only custom roles are selected', () => {
            component.selectedRows = [customRole];
            expect(component.isBuiltinSelected()).toBeFalse();
        });

        it('returns true when a built-in role is selected', () => {
            component.selectedRows = [builtinRole];
            expect(component.isBuiltinSelected()).toBeTrue();
        });

        it('returns true when a mix of built-in and custom roles is selected', () => {
            component.selectedRows = [builtinRole, customRole];
            expect(component.isBuiltinSelected()).toBeTrue();
        });
    });

    describe('action button disabled state', () => {
        it('edit button is disabled when a built-in role is selected', async () => {
            fixture.autoDetectChanges();
            await fixture.whenStable();
            component.selectedRows = [builtinRole];
            fixture.detectChanges();
            const editBtn: HTMLButtonElement = fixture.nativeElement
                .querySelector('#system-robot-edit')
                ?.closest('button');
            expect(editBtn?.disabled).toBeTrue();
        });

        it('edit button is enabled when a custom role is selected', async () => {
            fixture.autoDetectChanges();
            await fixture.whenStable();
            component.selectedRows = [customRole];
            fixture.detectChanges();
            const editBtn: HTMLButtonElement = fixture.nativeElement
                .querySelector('#system-robot-edit')
                ?.closest('button');
            expect(editBtn?.disabled).toBeFalse();
        });

        it('delete button is disabled when a built-in role is selected', async () => {
            fixture.autoDetectChanges();
            await fixture.whenStable();
            component.selectedRows = [builtinRole];
            fixture.detectChanges();
            const deleteBtn: HTMLButtonElement = fixture.nativeElement
                .querySelector('#system-robot-delete')
                ?.closest('button');
            expect(deleteBtn?.disabled).toBeTrue();
        });

        it('delete button is disabled when a mix of built-in and custom roles is selected', async () => {
            fixture.autoDetectChanges();
            await fixture.whenStable();
            component.selectedRows = [builtinRole, customRole];
            fixture.detectChanges();
            const deleteBtn: HTMLButtonElement = fixture.nativeElement
                .querySelector('#system-robot-delete')
                ?.closest('button');
            expect(deleteBtn?.disabled).toBeTrue();
        });
    });
});
