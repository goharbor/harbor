import { Component, Output, EventEmitter, OnInit } from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';

import { ModalEvent } from '../modal-event';
import { modalEvents } from '../modal-events.const';

import { SessionUser } from '../../shared/session-user';
import { SessionService } from '../../shared/session.service';
import { CookieService } from 'angular2-cookie/core';

import { supportedLangs, enLang, languageNames, CommonRoutes } from '../../shared/shared.const';
import { AppConfigService } from '../../app-config.service';
import { SearchTriggerService } from '../global-search/search-trigger.service';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';

@Component({
    selector: 'navigator',
    templateUrl: "navigator.component.html",
    styleUrls: ["navigator.component.css"]
})

export class NavigatorComponent implements OnInit {
    // constructor(private router: Router){}
    @Output() showAccountSettingsModal = new EventEmitter<ModalEvent>();
    @Output() showPwdChangeModal = new EventEmitter<ModalEvent>();

    private selectedLang: string = enLang;
    private appTitle: string = 'APP_TITLE.HARBOR';

    constructor(
        private session: SessionService,
        private router: Router,
        private translate: TranslateService,
        private cookie: CookieService,
        private appConfigService: AppConfigService,
        private msgHandler: MessageHandlerService,
        private searchTrigger: SearchTriggerService) {

    }

    ngOnInit(): void {
        this.selectedLang = this.translate.currentLang;
        this.translate.onLangChange.subscribe(langChange => {
            this.selectedLang = langChange.lang;
            //Keep in cookie for next use
            this.cookie.put("harbor-lang", langChange.lang);
        });
        if (this.appConfigService.isIntegrationMode()) {
            this.appTitle = 'APP_TITLE.VIC';
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
            this.session.getCurrentUser().has_admin_role > 0 &&
            this.appConfigService.getConfig() &&
            this.appConfigService.getConfig().has_ca_root;
    }

    public get canChangePassword(): boolean {
        return this.session.getCurrentUser() &&
            this.appConfigService.getConfig() &&
            this.appConfigService.getConfig().auth_mode != 'ldap_auth';
    }

    matchLang(lang: string): boolean {
        return lang.trim() === this.selectedLang;
    }

    //Open the account setting dialog
    openAccountSettingsModal(): void {
        this.showAccountSettingsModal.emit({
            modalName: modalEvents.USER_PROFILE,
            modalFlag: true
        });
    }

    //Open change password dialog
    openChangePwdModal(): void {
        this.showPwdChangeModal.emit({
            modalName: modalEvents.CHANGE_PWD,
            modalFlag: true
        });
    }

    //Open about dialog
    openAboutDialog(): void {
        this.showPwdChangeModal.emit({
            modalName: modalEvents.ABOUT,
            modalFlag: true
        });
    }

    //Log out system
    logOut(): void {
        //Naviagte to the sign in route
        //Appending 'signout' means destroy session cache
        let navigatorExtra: NavigationExtras = {
            queryParams: { "signout": true }
        };
        this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], navigatorExtra);
        //Confirm search result panel is close
        this.searchTrigger.closeSearch(true);
    }

    //Switch languages
    switchLanguage(lang: string): void {
        if (supportedLangs.find(supportedLang => supportedLang === lang.trim())) {
            this.translate.use(lang);
        } else {
            this.translate.use(enLang);//Use default
            //TODO:
            console.error('Language ' + lang.trim() + ' is not suppoted');
        }
        setTimeout(() => {
            window.location.reload();
        }, 500);
    }

    //Handle the home action
    homeAction(): void {
        if (this.session.getCurrentUser() != null) {
            //Navigate to default page
            this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
        } else {
            //Naviagte to signin page
            this.router.navigate([CommonRoutes.HARBOR_ROOT]);
        }

        //Confirm search result panel is close
        this.searchTrigger.closeSearch(true);
    }

    registryAction(): void {
        this.searchTrigger.closeSearch(true);
    }
}