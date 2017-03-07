import { Component, OnInit, ViewChild } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';

import { ModalEvent } from '../modal-event';
import { SearchEvent } from '../search-event';
import { modalEvents } from '../modal-events.const';

import { AccountSettingsModalComponent } from '../../account/account-settings/account-settings-modal.component';
import { SearchResultComponent } from '../global-search/search-result.component';
import { PasswordSettingComponent } from '../../account/password/password-setting.component';
import { NavigatorComponent } from '../navigator/navigator.component';
import { SessionService } from '../../shared/session.service';

import { AboutDialogComponent } from '../../shared/about-dialog/about-dialog.component'

@Component({
    selector: 'harbor-shell',
    templateUrl: 'harbor-shell.component.html',
    styleUrls: ["harbor-shell.component.css"]
})

export class HarborShellComponent implements OnInit {

    @ViewChild(AccountSettingsModalComponent)
    private accountSettingsModal: AccountSettingsModalComponent;

    @ViewChild(SearchResultComponent)
    private searchResultComponet: SearchResultComponent;

    @ViewChild(PasswordSettingComponent)
    private pwdSetting: PasswordSettingComponent;

    @ViewChild(NavigatorComponent)
    private navigator: NavigatorComponent;

    @ViewChild(AboutDialogComponent)
    private aboutDialog: AboutDialogComponent;

    //To indicator whwther or not the search results page is displayed
    //We need to use this property to do some overriding work
    private isSearchResultsOpened: boolean = false;

    constructor(
        private route: ActivatedRoute,
        private session: SessionService) { }

    ngOnInit() {
        this.route.data.subscribe(data => {
            //dummy
        });
    }

    public get showSearch(): boolean {
        return this.isSearchResultsOpened;
    }

    public get isSystemAdmin(): boolean {
        let account = this.session.getCurrentUser();
        return account != null && account.has_admin_role > 0;
    }

    public get isUserExisting(): boolean {
        let account = this.session.getCurrentUser();
        return account != null;
    }

    //Open modal dialog
    openModal(event: ModalEvent): void {
        switch (event.modalName) {
            case modalEvents.USER_PROFILE:
                this.accountSettingsModal.open();
                break;
            case modalEvents.CHANGE_PWD:
                this.pwdSetting.open();
                break;
            case modalEvents.ABOUT:
                this.aboutDialog.open();
                break;
            default:
                break;
        }
    }

    //Handle the global search event and then let the result page to trigger api
    doSearch(event: SearchEvent): void {
        //Once this method is called
        //the search results page must be opened
        this.isSearchResultsOpened = true;

        //Call the child component to do the real work
        this.searchResultComponet.doSearch(event.term);
    }

    //Search results page closed
    //remove the related ovevriding things
    searchClose(event: boolean): void {
        if (event) {
            this.isSearchResultsOpened = false;
        }
    }

    //Close serch result panel if existing
    watchClickEvt(): void {
        this.searchResultComponet.close();
        this.isSearchResultsOpened = false;
    }
}