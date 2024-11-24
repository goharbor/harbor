import { ComponentFixture, TestBed } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { SessionService } from '../../shared/services/session.service';
import { CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA } from '@angular/core';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { SearchTriggerService } from '../../shared/components/global-search/search-trigger.service';
import { HarborShellComponent } from './harbor-shell.component';
import { ClarityModule } from '@clr/angular';
import { of } from 'rxjs';
import { modalEvents } from '../modal-events.const';
import { PasswordSettingComponent } from '../password-setting/password-setting.component';
import { AboutDialogComponent } from '../../shared/components/about-dialog/about-dialog.component';
import { FormsModule } from '@angular/forms';
import { MessageHandlerService } from '../../shared/services/message-handler.service';
import { PasswordSettingService } from '../password-setting/password-setting.service';
import { SkinableConfig } from '../../services/skinable-config.service';
import { AppConfigService } from '../../services/app-config.service';
import { ErrorHandler } from '../../shared/units/error-handler';
import { AccountSettingsModalComponent } from '../account-settings/account-settings-modal.component';
import { PreferenceSettingsComponent } from '../preference-settings/preference-settings.component';
import { InlineAlertComponent } from '../../shared/components/inline-alert/inline-alert.component';
import { ScannerService } from '../../../../ng-swagger-gen/services/scanner.service';
import { UserService } from '../../../../ng-swagger-gen/services/user.service';

// Mocks
const fakeSessionService = {
    getCurrentUser: function () {
        return { has_admin_role: true };
    },
};

const fakeSearchTriggerService = {
    searchTriggerChan$: of('null'),
    searchCloseChan$: of(null),
};

const mockMessageHandlerService = null;

const mockPasswordSettingService = null;

const mockSkinableConfig = {
    getSkinConfig: function () {
        return {
            headerBgColor: {
                darkMode: '',
                lightMode: '',
            },
            loginBgImg: '',
            loginTitle: '',
            product: {
                name: '',
                logo: '',
                introduction: '',
            },
        };
    },
};

const fakeAppConfigService = {
    isLdapMode: function () {
        return true;
    },
    isHttpAuthMode: function () {
        return false;
    },
    isOidcMode: function () {
        return false;
    },
    getConfig: function () {
        return {
            with_trivy: true,
        };
    },
};

const fakedUserService = {
    getCurrentUserInfo() {
        return of({});
    },
    setCliSecret() {
        return of(null);
    },
};

describe('HarborShellComponent', () => {
    let component: HarborShellComponent;
    let fixture: ComponentFixture<HarborShellComponent>;

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [
                RouterTestingModule,
                TranslateModule.forRoot(), // Ensure TranslateModule is imported
                ClarityModule,
                BrowserAnimationsModule,
                FormsModule,
            ],
            declarations: [
                HarborShellComponent,
                AccountSettingsModalComponent,
                PreferenceSettingsComponent,
                PasswordSettingComponent,
                AboutDialogComponent,
                InlineAlertComponent,
            ],
            providers: [
                TranslateService,
                { provide: SessionService, useValue: fakeSessionService },
                {
                    provide: SearchTriggerService,
                    useValue: fakeSearchTriggerService,
                },
                { provide: AppConfigService, useValue: fakeAppConfigService },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
                { provide: UserService, useValue: fakedUserService },
                {
                    provide: PasswordSettingService,
                    useValue: mockPasswordSettingService,
                },
                { provide: SkinableConfig, useValue: mockSkinableConfig },
                ErrorHandler,
            ],
            schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(HarborShellComponent);
        component = fixture.componentInstance;
        component.accountSettingsModal = TestBed.createComponent(
            AccountSettingsModalComponent
        ).componentInstance;
        component.accountSettingsModal.inlineAlert =
            TestBed.createComponent(InlineAlertComponent).componentInstance;
        component.prefSetting = TestBed.createComponent(
            PreferenceSettingsComponent
        ).componentInstance;
        component.pwdSetting = TestBed.createComponent(
            PasswordSettingComponent
        ).componentInstance;
        component.aboutDialog =
            TestBed.createComponent(AboutDialogComponent).componentInstance;
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should open users profile', async () => {
        component.openModal({
            modalName: modalEvents.USER_PROFILE,
            modalFlag: false,
        });
        await fixture.whenStable();
        const accountSettingsUsernameInput =
            fixture.nativeElement.querySelector('#account_settings_username');
        expect(accountSettingsUsernameInput).toBeTruthy();
    });

    it('should open users preferences', async () => {
        component.openModal({
            modalName: modalEvents.PREFERENCES,
            modalFlag: false,
        });
        await fixture.whenStable();
        const dropdowns = fixture.nativeElement.querySelector('.dropdowns');
        expect(dropdowns).toBeTruthy();
    });

    it('should open users changePwd', async () => {
        component.openModal({
            modalName: modalEvents.CHANGE_PWD,
            modalFlag: false,
        });
        await fixture.whenStable();
        const oldPasswordInput =
            fixture.nativeElement.querySelector('#oldPassword');
        expect(oldPasswordInput).toBeTruthy();
    });

    it('should open users about-dialog', async () => {
        component.openModal({ modalName: modalEvents.ABOUT, modalFlag: false });
        await fixture.whenStable();
        const aboutVersionEl =
            fixture.nativeElement.querySelector('.about-version');
        expect(aboutVersionEl).toBeTruthy();
    });
});
