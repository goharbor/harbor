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
import { RouterTestingModule } from '@angular/router/testing';
import { AccountSettingsModalComponent } from './account-settings-modal.component';
import { SessionService } from '../../shared/services/session.service';
import { MessageHandlerService } from '../../shared/services/message-handler.service';
import { SearchTriggerService } from '../../shared/components/global-search/search-trigger.service';
import { CUSTOM_ELEMENTS_SCHEMA, ChangeDetectorRef } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { FormsModule } from '@angular/forms';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { of } from 'rxjs';
import { Router } from '@angular/router';
import { clone } from '../../shared/units/utils';
import { ErrorHandler } from '../../shared/units/error-handler';
import { ConfirmationDialogComponent } from '../../shared/components/confirmation-dialog';
import { InlineAlertComponent } from '../../shared/components/inline-alert/inline-alert.component';
import { ConfirmationDialogService } from '../global-confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../global-confirmation-dialog/confirmation-message';
import { UserService } from '../../../../ng-swagger-gen/services/user.service';
import { ProjectService } from '../../../../ng-swagger-gen/services/project.service';
import { AppConfigService } from '../../services/app-config.service';

describe('AccountSettingsModalComponent', () => {
    let component: AccountSettingsModalComponent;
    let fixture: ComponentFixture<AccountSettingsModalComponent>;
    let userExisting = false;
    let oidcUserMeta: any = true;
    let oidcUserMeta1 = {
        id: 1,
        user_id: 1,
        secret: 'Asdf12345',
        subiss: 'string',
    };
    let fakeSessionService = {
        getCurrentUser: function () {
            return {
                has_admin_role: true,
                user_id: 1,
                username: 'admin',
                email: '',
                realname: 'admin',
                role_name: 'admin',
                role_id: 1,
                comment: 'string',
                oidc_user_meta: oidcUserMeta,
            };
        },
        checkUserExisting: () => of(userExisting),
        updateAccountSettings: () => of(null),
        renameAdmin: () => of(null),
    };
    let fakeMessageHandlerService = {
        showSuccess: () => {},
    };
    let fakeSearchTriggerService = {
        closeSearch: () => {},
    };
    let fakeConfirmationDialogService = {
        cancel: () => of(null),
        confirm: () => of(null),
        confirmationAnnouced$: of(
            new ConfirmationMessage('null', 'null', 'null', 'null', null, null)
        ),
    };
    let fakeRouter = {
        navigate: () => {},
    };

    const fakedUserService = {
        getCurrentUserInfo() {
            return of({});
        },
        setCliSecret() {
            return of(null);
        },
        ListPersonalAccessTokens(_options: any) {
            return of([]);
        },
        CreatePersonalAccessToken(_options: any) {
            return of({
                id: 1,
                name: 'test',
                secret: 'hbr_pat_test123',
                expires_at: -1,
            });
        },
    };

    const fakedProjectService = {
        listProjects(_options: any) {
            return of([
                { project_id: 1, name: 'library' },
                { project_id: 2, name: 'my-project' },
            ]);
        },
    };

    const MockedAppConfigService = {
        getConfig() {
            return { self_registration: true };
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [
                AccountSettingsModalComponent,
                InlineAlertComponent,
                ConfirmationDialogComponent,
            ],
            imports: [
                RouterTestingModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                BrowserAnimationsModule,
            ],
            providers: [
                ChangeDetectorRef,
                TranslateService,
                ErrorHandler,
                { provide: SessionService, useValue: fakeSessionService },
                {
                    provide: MessageHandlerService,
                    useValue: fakeMessageHandlerService,
                },
                {
                    provide: SearchTriggerService,
                    useValue: fakeSearchTriggerService,
                },
                {
                    provide: UserService,
                    useValue: fakedUserService,
                },
                {
                    provide: ProjectService,
                    useValue: fakedProjectService,
                },
                { provide: Router, useValue: fakeRouter },
                {
                    provide: ConfirmationDialogService,
                    useValue: fakeConfirmationDialogService,
                },
                { provide: AppConfigService, useValue: MockedAppConfigService },
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AccountSettingsModalComponent);
        component = fixture.componentInstance;
        component.inlineAlert =
            TestBed.createComponent(InlineAlertComponent).componentInstance;
        component.error = true;
        component.open();
        // // component.confirmationDialogComponent.ope;
        oidcUserMeta = true;
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should input right email', async () => {
        await fixture.whenStable();
        // Update the title input
        userExisting = true;
        let emailInput: HTMLInputElement = fixture.nativeElement.querySelector(
            '#account_settings_email'
        );
        emailInput.value = 'email@qq.com';
        emailInput.dispatchEvent(new Event('input'));
        fixture.detectChanges();
        expect(emailInput.value).toEqual('email@qq.com');
        expect(component.emailTooltip).toEqual('TOOLTIP.EMAIL');
        emailInput.dispatchEvent(new Event('blur'));
        expect(component.emailTooltip).toEqual('TOOLTIP.EMAIL_EXISTING');
        emailInput.dispatchEvent(new Event('blur'));
        expect(component.emailTooltip).toEqual('TOOLTIP.EMAIL_EXISTING');
        userExisting = false;
        emailInput.value = '123@qq.com';
        emailInput.dispatchEvent(new Event('blur'));
        expect(emailInput.value).toEqual('123@qq.com');
    });

    it('should update settings', async () => {
        await fixture.whenStable();
        let emailInput: HTMLInputElement = fixture.nativeElement.querySelector(
            '#account_settings_email'
        );
        emailInput.value = 'email@qq.com';
        emailInput.dispatchEvent(new Event('input'));
        let fullNameInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#account_settings_full_name');
        fullNameInput.value = 'system guest';
        fullNameInput.dispatchEvent(new Event('input'));
        let submitBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#submit-btn');
        submitBtn.dispatchEvent(new Event('click'));
        const emailInput1: HTMLInputElement =
            fixture.nativeElement.querySelector('#account_settings_email');
        expect(emailInput1).toEqual(null);
    });
    it('admin should rename', async () => {
        await fixture.whenStable();
        let renameBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#rename-btn');
        renameBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const userNameInput: HTMLInputElement =
            fixture.nativeElement.querySelector('#account_settings_username');
        expect(userNameInput.value).toEqual('admin@harbor.local');
        expect(component.RenameOnGoing).toEqual(true);
    });
    it('admin should save when it click save button 2 times after rename', async () => {
        await fixture.whenStable();
        let renameBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#rename-btn');
        renameBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        let emailInput: HTMLInputElement = fixture.nativeElement.querySelector(
            '#account_settings_email'
        );
        emailInput.value = 'email@qq.com';
        emailInput.dispatchEvent(new Event('input'));
        let submitBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#submit-btn');
        submitBtn.dispatchEvent(new Event('click'));
        const alertTextElement: HTMLSpanElement =
            fixture.nativeElement.querySelector('.alert-text');
        expect(alertTextElement.innerText).toEqual(
            ' PROFILE.RENAME_CONFIRM_INFO '
        );

        submitBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const emailInput1: HTMLInputElement =
            fixture.nativeElement.querySelector('#account_settings_email');
        // rename success
        expect(emailInput1).toEqual(null);
    });
    it('should click cancel and close when has data change and no rename', async () => {
        await fixture.whenStable();
        let emailInput: HTMLInputElement = fixture.nativeElement.querySelector(
            '#account_settings_email'
        );
        emailInput.value = 'email@qq.com';
        emailInput.dispatchEvent(new Event('input'));
        let cancelBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#cancel-btn');
        cancelBtn.dispatchEvent(new Event('click'));
        const alertTextElement: HTMLSpanElement =
            fixture.nativeElement.querySelector('.alert-text');
        expect(alertTextElement.innerText).toEqual(
            ' ALERT.FORM_CHANGE_CONFIRMATION '
        );
    });
    it('should click cancel and close when has data change and has rename', async () => {
        await fixture.whenStable();
        let renameBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#rename-btn');
        renameBtn.dispatchEvent(new Event('click'));
        let cancelBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#cancel-btn');
        cancelBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        expect(component.RenameOnGoing).toEqual(false);
        const emailInput1: HTMLInputElement =
            fixture.nativeElement.querySelector('#account_settings_email');
        expect(emailInput1).toEqual(null);
    });
    it('should click cancel and close when has no data change', async () => {
        await fixture.whenStable();
        let cancelBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#cancel-btn');
        cancelBtn.dispatchEvent(new Event('click'));
        await fixture.whenStable();
        const emailInput1: HTMLInputElement =
            fixture.nativeElement.querySelector('#account_settings_email');
        expect(emailInput1).toEqual(null);
    });
    // CLI secret test omitted - UI elements not present in this template version

    // Scope section
    it('should load scope projects on openCreatePATModal', async () => {
        spyOn(fakedProjectService, 'listProjects').and.callThrough();
        component.openCreatePATModal();
        expect(component.scopeProjects.length).toBe(2);
        expect(component.scopeProjects[0].project_name).toBe('library');
        expect(component.scopeProjects[0].pull).toBeTrue();
        expect(component.scopeProjects[0].push).toBeTrue();
    });

    it('selectAllScope should select all', () => {
        component.scopeProjects = [
            { project_id: 1, project_name: 'a', pull: false, push: false },
            { project_id: 2, project_name: 'b', pull: false, push: false },
        ];
        component.selectAllScope();
        expect(component.scopeProjects.every(p => p.pull && p.push)).toBeTrue();
    });

    it('deselectAllScope should deselect all', () => {
        component.scopeProjects = [
            { project_id: 1, project_name: 'a', pull: true, push: true },
            { project_id: 2, project_name: 'b', pull: true, push: true },
        ];
        component.deselectAllScope();
        expect(
            component.scopeProjects.every(p => !p.pull && !p.push)
        ).toBeTrue();
    });

    it('allScopeSelected returns true when all selected', () => {
        component.scopeProjects = [
            { project_id: 1, project_name: 'a', pull: true, push: true },
            { project_id: 2, project_name: 'b', pull: true, push: true },
        ];
        expect(component.allScopeSelected).toBeTrue();
    });

    it('allScopeSelected returns false when some deselected', () => {
        component.scopeProjects = [
            { project_id: 1, project_name: 'a', pull: true, push: false },
            { project_id: 2, project_name: 'b', pull: true, push: true },
        ];
        expect(component.allScopeSelected).toBeFalse();
    });

    it('buildScopeJson returns undefined when all selected', () => {
        component.scopeProjects = [
            { project_id: 1, project_name: 'library', pull: true, push: true },
        ];
        expect(component.buildScopeJson()).toBeUndefined();
    });

    it('buildScopeJson returns scope JSON when some deselected', () => {
        component.scopeProjects = [
            { project_id: 1, project_name: 'library', pull: true, push: false },
        ];
        const json = component.buildScopeJson();
        expect(json).toBeDefined();
        const parsed = JSON.parse(json!);
        expect(parsed.length).toBe(1);
        expect(parsed[0].project_id).toBe(1);
        expect(parsed[0].access[0].actions).toEqual(['pull']);
    });

    it('buildScopeJson omits projects with no permissions', () => {
        component.scopeProjects = [
            { project_id: 1, project_name: 'lib', pull: false, push: false },
            { project_id: 2, project_name: 'proj', pull: true, push: true },
        ];
        const json = component.buildScopeJson();
        const parsed = JSON.parse(json!);
        expect(parsed.length).toBe(1);
        expect(parsed[0].project_id).toBe(2);
    });

    it('buildScopeJson includes both pull and push', () => {
        component.scopeProjects = [
            { project_id: 1, project_name: 'lib', pull: true, push: true },
            { project_id: 2, project_name: 'other', pull: false, push: false },
        ];
        const json = component.buildScopeJson();
        const parsed = JSON.parse(json!);
        expect(parsed[0].access[0].actions).toContain('pull');
        expect(parsed[0].access[0].actions).toContain('push');
    });

    it('createPAT includes scope when scope is customized', () => {
        component.account = { user_id: 1 } as any;
        component.newPATForm = {
            name: 'test',
            expiresInDays: 0,
            description: '',
        };
        component.scopeProjects = [
            { project_id: 1, project_name: 'lib', pull: true, push: false },
        ];
        spyOn(fakedUserService, 'CreatePersonalAccessToken').and.callThrough();
        component.createPAT();
        expect(fakedUserService.CreatePersonalAccessToken).toHaveBeenCalledWith(
            jasmine.objectContaining({
                request: jasmine.objectContaining({
                    scope: jasmine.any(String),
                }),
            })
        );
    });

    it('createPAT omits scope when all selected (auto-compute)', () => {
        component.account = { user_id: 1 } as any;
        component.newPATForm = {
            name: 'test2',
            expiresInDays: 0,
            description: '',
        };
        component.scopeProjects = [
            { project_id: 1, project_name: 'lib', pull: true, push: true },
        ];
        spyOn(fakedUserService, 'CreatePersonalAccessToken').and.callThrough();
        component.createPAT();
        expect(fakedUserService.CreatePersonalAccessToken).toHaveBeenCalledWith(
            jasmine.objectContaining({
                request: jasmine.objectContaining({ scope: undefined }),
            })
        );
    });
});
