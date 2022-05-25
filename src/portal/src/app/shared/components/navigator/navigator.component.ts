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
import { Component, Output, EventEmitter, OnInit } from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';
import { PlatformLocation } from '@angular/common';
import { ModalEvent } from '../../../base/modal-event';
import { modalEvents } from '../../../base/modal-events.const';
import { SessionService } from '../../services/session.service';
import { AppConfigService } from '../../../services/app-config.service';
import { SearchTriggerService } from '../global-search/search-trigger.service';
import { MessageHandlerService } from '../../services/message-handler.service';
import { SkinableConfig } from '../../../services/skinable-config.service';
import {
    CommonRoutes,
    DATETIME_RENDERINGS,
    DatetimeRendering,
    DEFAULT_DATETIME_RENDERING_LOCALSTORAGE_KEY,
    DEFAULT_LANG_LOCALSTORAGE_KEY,
    DefaultDatetimeRendering,
    DeFaultLang,
    LANGUAGES,
    SupportedLanguage,
} from '../../entities/shared.const';
import {
    CustomStyle,
    HAS_STYLE_MODE,
    StyleMode,
} from '../../../services/theme';
import { getDatetimeRendering } from '../../units/shared.utils';

@Component({
    selector: 'navigator',
    templateUrl: 'navigator.component.html',
    styleUrls: ['navigator.component.scss'],
})
export class NavigatorComponent implements OnInit {
    @Output() showAccountSettingsModal = new EventEmitter<ModalEvent>();
    @Output() showDialogModalAction = new EventEmitter<ModalEvent>();

    readonly guiLanguages = Object.entries(LANGUAGES);
    readonly guiDatetimeRenderings = Object.entries(DATETIME_RENDERINGS);
    selectedLang: SupportedLanguage = DeFaultLang;
    selectedDatetimeRendering: DatetimeRendering = DefaultDatetimeRendering;
    appTitle: string = 'APP_TITLE.HARBOR';
    customStyle: CustomStyle;
    constructor(
        private session: SessionService,
        private router: Router,
        private location: PlatformLocation,
        private translate: TranslateService,
        private appConfigService: AppConfigService,
        private msgHandler: MessageHandlerService,
        private searchTrigger: SearchTriggerService,
        private skinableConfig: SkinableConfig
    ) {}

    ngOnInit(): void {
        // custom skin
        this.customStyle = this.skinableConfig.getSkinConfig();
        this.selectedLang = this.translate.currentLang as SupportedLanguage;
        this.selectedDatetimeRendering = getDatetimeRendering();
        if (this.appConfigService.isIntegrationMode()) {
            this.appTitle = 'APP_TITLE.VIC';
        }

        if (this.appConfigService.getConfig().read_only) {
            this.msgHandler.handleReadOnly();
        }
    }

    public get isSessionValid(): boolean {
        return this.session.getCurrentUser() != null;
    }

    public get accountName(): string {
        return this.session.getCurrentUser()
            ? this.session.getCurrentUser().username
            : 'N/A';
    }

    public get currentLang(): string {
        return LANGUAGES[this.selectedLang];
    }

    public get currentDatetimeRendering(): string {
        return DATETIME_RENDERINGS[this.selectedDatetimeRendering];
    }

    public get admiralLink(): string {
        return this.appConfigService.getAdmiralEndpoint(window.location.href);
    }

    public get isIntegrationMode(): boolean {
        return this.appConfigService.isIntegrationMode();
    }

    public get canDownloadCert(): boolean {
        return (
            this.session.getCurrentUser() &&
            this.session.getCurrentUser().has_admin_role &&
            this.appConfigService.getConfig() &&
            this.appConfigService.getConfig().has_ca_root
        );
    }

    public get canChangePassword(): boolean {
        let user = this.session.getCurrentUser();
        let config = this.appConfigService.getConfig();

        return (
            user &&
            ((config &&
                !(
                    config.auth_mode === 'ldap_auth' ||
                    config.auth_mode === 'uaa_auth' ||
                    config.auth_mode === 'oidc_auth'
                )) ||
                (user.user_id === 1 && user.username === 'admin'))
        );
    }

    matchLang(lang: SupportedLanguage): boolean {
        return lang === this.selectedLang;
    }

    matchDatetimeRendering(datetime: DatetimeRendering): boolean {
        return datetime === this.selectedDatetimeRendering;
    }

    // Open the account setting dialog
    openAccountSettingsModal(): void {
        this.showAccountSettingsModal.emit({
            modalName: modalEvents.USER_PROFILE,
            modalFlag: true,
        });
    }

    // Open change password dialog
    openChangePwdModal(): void {
        this.showDialogModalAction.emit({
            modalName: modalEvents.CHANGE_PWD,
            modalFlag: true,
        });
    }

    // Open about dialog
    openAboutDialog(): void {
        this.showDialogModalAction.emit({
            modalName: modalEvents.ABOUT,
            modalFlag: true,
        });
    }

    // Log out system
    logOut(): void {
        // Naviagte to the sign in router-guard
        // Appending 'signout' means destroy session cache
        let signout = true;
        let redirect_url = this.location.pathname;
        let navigatorExtra: NavigationExtras = {
            queryParams: { signout, redirect_url },
        };
        this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], navigatorExtra);
        // Confirm search result panel is close
        this.searchTrigger.closeSearch(true);
    }

    // Switch languages
    switchLanguage(lang: SupportedLanguage): void {
        this.selectedLang = lang;
        localStorage.setItem(DEFAULT_LANG_LOCALSTORAGE_KEY, lang);
        // due to the bug(https://github.com/ngx-translate/core/issues/1258) of translate module
        // have to reload
        this.translate.use(lang).subscribe(() => window.location.reload());
    }

    switchDatetimeRendering(datetime: DatetimeRendering): void {
        this.selectedDatetimeRendering = datetime;
        localStorage.setItem(
            DEFAULT_DATETIME_RENDERING_LOCALSTORAGE_KEY,
            datetime
        );
        // have to reload,as HarborDatetimePipe is pure pipe
        window.location.reload();
    }

    // Handle the home action
    homeAction(): void {
        if (this.session.getCurrentUser() != null) {
            // Navigate to default page
            this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
        } else {
            // Naviagte to signin page
            this.router.navigate([CommonRoutes.HARBOR_ROOT]);
        }

        // Confirm search result panel is close
        this.searchTrigger.closeSearch(true);
    }

    registryAction(): void {
        this.searchTrigger.closeSearch(true);
    }

    getBgColor(): string {
        if (
            this.customStyle &&
            this.customStyle.headerBgColor &&
            localStorage
        ) {
            if (localStorage.getItem(HAS_STYLE_MODE) === StyleMode.LIGHT) {
                return `background-color:${this.customStyle.headerBgColor.lightMode} !important`;
            }
            if (localStorage.getItem(HAS_STYLE_MODE) === StyleMode.DARK) {
                return `background-color:${this.customStyle.headerBgColor.darkMode} !important`;
            }
        }
        return null;
    }
}
