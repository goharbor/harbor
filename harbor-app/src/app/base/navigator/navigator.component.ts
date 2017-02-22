import { Component, Output, EventEmitter, OnInit } from '@angular/core';
import { Router } from '@angular/router';

import { ModalEvent } from '../modal-event';
import { SearchEvent } from '../search-event';
import { modalAccountSettings, modalPasswordSetting } from '../modal-events.const';

import { SessionUser } from '../../shared/session-user';
import { SessionService } from '../../shared/session.service';

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

    constructor(private session: SessionService, private router: Router) { }

    ngOnInit(): void {
        this.sessionUser = this.session.getCurrentUser();
    }

    public get isSessionValid(): boolean {
        return this.sessionUser != null;
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
}