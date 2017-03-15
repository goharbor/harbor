import { Component, Output, EventEmitter, OnInit, Inject } from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';

import { ModalEvent } from '../modal-event';
import { modalEvents } from '../modal-events.const';

import { SessionUser } from '../../shared/session-user';
import { SessionService } from '../../shared/session.service';
import { CookieService } from 'angular2-cookie/core';

import { supportedLangs, enLang, languageNames, CommonRoutes } from '../../shared/shared.const';

import { AppConfigService } from '../../app-config.service';
import { AppConfig } from '../../app-config';

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
    private appConfig: AppConfig = new AppConfig();

    constructor(
        private session: SessionService,
        private router: Router,
        private translate: TranslateService,
        private cookie: CookieService,
        private appConfigService: AppConfigService) { }

    ngOnInit(): void {
        this.selectedLang = this.translate.currentLang;
        this.translate.onLangChange.subscribe(langChange => {
            this.selectedLang = langChange.lang;
            //Keep in cookie for next use
            this.cookie.put("harbor-lang", langChange.lang);
        });

        this.appConfig = this.appConfigService.getConfig();
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

    public get isIntegrationMode(): boolean {
        return this.appConfig.with_admiral && this.appConfig.admiral_endpoint.trim() != "";
    }

    public get admiralLink(): string {
        let routeSegments = [this.appConfig.admiral_endpoint,
            "?registry_url=",
        encodeURIComponent(window.location.href)
        ];

        return routeSegments.join("");
    }

    public get showSignUpLink(): boolean {
        return this.appConfig.auth_mode === "db_auth" &&
            this.appConfig.self_registration &&
            !this.isSessionValid &&
            !this.isSignInRoute();
    }

    public get showSignInLink(): boolean {
        return !this.isSignInRoute() && !this.isSessionValid;
    }

    public get loginUrl(): string {
        if (this.isIntegrationMode) {
            return CommonRoutes.EMBEDDED_SIGN_IN;
        } else {
            return CommonRoutes.SIGN_IN;
        }
    }

    private isSignInRoute(): boolean {
        return this.router.routerState.snapshot.url.toString().startsWith(CommonRoutes.EMBEDDED_SIGN_IN);
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
        this.session.signOff()
            .then(() => {
                //Naviagte to the sign in route
                if (this.isIntegrationMode) {
                    this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN]);
                } else {
                    this.router.navigate([CommonRoutes.SIGN_IN]);
                }
            })
            .catch(error => {
                console.error("Log out with error: ", error);
            });
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
        //Try to switch backend lang
        //this.session.switchLanguage(lang).catch(error => console.error(error));
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
    }

    openSignUp(): void {
        let navigatorExtra: NavigationExtras = {
            queryParams: { "sign_up": true }
        };

        if (this.isIntegrationMode) {
            this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], navigatorExtra);
        } else {
            this.router.navigate([CommonRoutes.SIGN_IN], navigatorExtra);
        }
    }
}