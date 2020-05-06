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
// import {  map } from 'rxjs/operators';
import { PlatformLocation } from '@angular/common';
import { ModalEvent } from '../modal-event';
import { modalEvents } from '../modal-events.const';
import { SessionService } from '../../shared/session.service';
import { CookieService, CookieOptions } from 'ngx-cookie';
import { supportedLangs, enLang, languageNames } from '../../shared/shared.const';
import { AppConfigService } from '../../services/app-config.service';
import { SearchTriggerService } from '../global-search/search-trigger.service';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { SkinableConfig } from "../../services/skinable-config.service";
import { CommonRoutes } from "../../../lib/entities/shared.const";


@Component({
    selector: 'navigator',
    templateUrl: "navigator.component.html",
    styleUrls: ["navigator.component.scss"]
})

export class NavigatorComponent implements OnInit {
    // constructor(private router: Router){}
    @Output() showAccountSettingsModal = new EventEmitter<ModalEvent>();
    @Output() showDialogModalAction = new EventEmitter<ModalEvent>();

    selectedLang: string = enLang;
    appTitle: string = 'APP_TITLE.HARBOR';
    customStyle: { [key: string]: any };
    customProjectName: { [key: string]: any };
    constructor(
        private session: SessionService,
        private router: Router,
        private location: PlatformLocation,
        private translate: TranslateService,
        private cookie: CookieService,
        private appConfigService: AppConfigService,
        private msgHandler: MessageHandlerService,
        private searchTrigger: SearchTriggerService,
        private skinableConfig: SkinableConfig) {
    }

    ngOnInit(): void {
        // custom skin
        let customSkinObj = this.skinableConfig.getSkinConfig();
        if (customSkinObj) {
            if (customSkinObj.product) {
                this.customProjectName = customSkinObj.product;
            }
            this.customStyle = customSkinObj;
        }

        this.selectedLang = this.translate.currentLang;
        this.translate.onLangChange.subscribe((langChange: { lang: string }) => {
            this.selectedLang = langChange.lang;
            // Keep in cookie for next use
            let opt: CookieOptions = { path: '/', expires: new Date(Date.now() + 3600 * 1000 * 24 * 31) };
            this.cookie.put("harbor-lang", langChange.lang, opt);
        });
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
        return this.session.getCurrentUser() ? this.session.getCurrentUser().username : "N/A";
    }

    public get currentLang(): string {
        return languageNames[this.selectedLang];
    }

    public get admiralLink(): string {
        return this.appConfigService.getAdmiralEndpoint(window.location.href);
    }

    public get isIntegrationMode(): boolean {
        return this.appConfigService.isIntegrationMode();
    }

    public get canDownloadCert(): boolean {
        return this.session.getCurrentUser() &&
            this.session.getCurrentUser().has_admin_role &&
            this.appConfigService.getConfig() &&
            this.appConfigService.getConfig().has_ca_root;
    }

    public get canChangePassword(): boolean {
        let user = this.session.getCurrentUser();
        let config = this.appConfigService.getConfig();

        return user && ((config && !(config.auth_mode === "ldap_auth" || config.auth_mode === "uaa_auth"
        || config.auth_mode === "oidc_auth")) || (user.user_id === 1 && user.username === "admin"));
    }

    matchLang(lang: string): boolean {
        return lang.trim() === this.selectedLang;
    }

    // Open the account setting dialog
    openAccountSettingsModal(): void {
        this.showAccountSettingsModal.emit({
            modalName: modalEvents.USER_PROFILE,
            modalFlag: true
        });
    }

    // Open change password dialog
    openChangePwdModal(): void {
        this.showDialogModalAction.emit({
            modalName: modalEvents.CHANGE_PWD,
            modalFlag: true
        });
    }

    // Open about dialog
    openAboutDialog(): void {
        this.showDialogModalAction.emit({
            modalName: modalEvents.ABOUT,
            modalFlag: true
        });
    }

    // Log out system
    logOut(): void {
        // Naviagte to the sign in route
        // Appending 'signout' means destroy session cache
        let signout = true;
        let redirect_url = this.location.pathname;
        let navigatorExtra: NavigationExtras = {
            queryParams: {signout, redirect_url}
        };
        this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], navigatorExtra);
        // Confirm search result panel is close
        this.searchTrigger.closeSearch(true);
    }

    // Switch languages
    switchLanguage(lang: string): void {
        let selectedLang: string = enLang; // Default
        if (supportedLangs.find(supportedLang => supportedLang === lang.trim())) {
            selectedLang = lang;
        } else {
            console.error('Language ' + lang.trim() + ' is not suppoted yet');
        }

        this.translate.use(selectedLang).subscribe(() => window.location.reload());
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
}
