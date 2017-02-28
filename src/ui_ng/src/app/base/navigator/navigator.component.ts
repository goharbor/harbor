import { Component, Output, EventEmitter, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';

import { ModalEvent } from '../modal-event';
import { SearchEvent } from '../search-event';
import { modalAccountSettings, modalPasswordSetting } from '../modal-events.const';

import { SessionUser } from '../../shared/session-user';
import { SessionService } from '../../shared/session.service';
import { CookieService } from 'angular2-cookie/core';

import { supportedLangs, enLang, languageNames } from '../../shared/shared.const';

@Component({
    selector: 'navigator',
    templateUrl: "navigator.component.html",
    styleUrls: ["navigator.component.css"]
})

export class NavigatorComponent implements OnInit {
    // constructor(private router: Router){}
    @Output() showAccountSettingsModal = new EventEmitter<ModalEvent>();
    @Output() searchEvt = new EventEmitter<SearchEvent>();
    @Output() showPwdChangeModal = new EventEmitter<ModalEvent>();

    private sessionUser: SessionUser = null;
    private selectedLang: string = enLang;

    constructor(
        private session: SessionService,
        private router: Router,
        private translate: TranslateService,
        private cookie: CookieService) { }

    ngOnInit(): void {
        this.sessionUser = this.session.getCurrentUser();
        this.selectedLang = this.translate.currentLang;
        this.translate.onLangChange.subscribe(langChange => {
            this.selectedLang = langChange.lang;
            //Keep in cookie for next use
            this.cookie.put("harbor-lang", langChange.lang);
        });
    }

    public get isSessionValid(): boolean {
        return this.sessionUser != null;
    }

    public get accountName(): string {
        return this.sessionUser ? this.sessionUser.username : "";
    }

    public get currentLang(): string {
        return languageNames[this.selectedLang];
    }

    matchLang(lang: string): boolean {
        return lang.trim() === this.selectedLang;
    }

    //Open the account setting dialog
    openAccountSettingsModal(): void {
        this.showAccountSettingsModal.emit({
            modalName: modalAccountSettings,
            modalFlag: true
        });
    }

    //Open change password dialog
    openChangePwdModal(): void {
        this.showPwdChangeModal.emit({
            modalName: modalPasswordSetting,
            modalFlag: true
        });
    }

    //Only transfer the search event to the parent shell
    transferSearchEvent(evt: SearchEvent): void {
        this.searchEvt.emit(evt);
    }

    //Log out system
    logOut(): void {
        this.session.signOff()
            .then(() => {
                this.sessionUser = null;
                //Naviagte to the sign in route
                this.router.navigate(["/sign-in"]);
            })
            .catch()//TODO:
    }

    //Switch languages
    switchLanguage(lang: string): void {
        if (supportedLangs.find(supportedLang => supportedLang === lang.trim())){
            this.translate.use(lang);
        }else{
            this.translate.use(enLang);//Use default
            //TODO:
            console.error('Language '+lang.trim()+' is not suppoted');
        }
    }

    //Handle the home action
    homeAction(): void {
        if(this.sessionUser != null){
            //Navigate to default page
            this.router.navigate(['harbor','projects']);
        }else{
            //Naviagte to signin page
            this.router.navigate(['sign-in']);
        }
    }
}