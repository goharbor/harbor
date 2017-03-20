import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';

import { ModalEvent } from '../modal-event';
import { modalEvents } from '../modal-events.const';

import { AccountSettingsModalComponent } from '../../account/account-settings/account-settings-modal.component';
import { SearchResultComponent } from '../global-search/search-result.component';
import { PasswordSettingComponent } from '../../account/password/password-setting.component';
import { NavigatorComponent } from '../navigator/navigator.component';
import { SessionService } from '../../shared/session.service';

import { AboutDialogComponent } from '../../shared/about-dialog/about-dialog.component';
import { StartPageComponent } from '../start-page/start.component';

import { SearchTriggerService } from '../global-search/search-trigger.service';

import { Subscription } from 'rxjs/Subscription';

import { CommonRoutes } from '../../shared/shared.const';

@Component({
    selector: 'harbor-shell',
    templateUrl: 'harbor-shell.component.html',
    styleUrls: ["harbor-shell.component.css"]
})

export class HarborShellComponent implements OnInit, OnDestroy {

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

    @ViewChild(StartPageComponent)
    private searchSatrt: StartPageComponent;

    //To indicator whwther or not the search results page is displayed
    //We need to use this property to do some overriding work
    private isSearchResultsOpened: boolean = false;

    private searchSub: Subscription;
    private searchCloseSub: Subscription;

    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private session: SessionService,
        private searchTrigger: SearchTriggerService) { }

    ngOnInit() {
        this.searchSub = this.searchTrigger.searchTriggerChan$.subscribe(searchEvt => {
            this.doSearch(searchEvt);
        });

        this.searchCloseSub = this.searchTrigger.searchCloseChan$.subscribe(close => {
            if (close) {
                this.searchClose();
            }else{
                this.watchClickEvt();//reuse
            }
        });
    }

    ngOnDestroy(): void {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
        }

        if (this.searchCloseSub) {
            this.searchCloseSub.unsubscribe();
        }
    }

    public get shouldOverrideContent(): boolean {
        return this.router.routerState.snapshot.url.toString().startsWith(CommonRoutes.EMBEDDED_SIGN_IN);
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
    doSearch(event: string): void {
        if (event === "") {
            if (!this.isSearchResultsOpened) {
                //Will not open search result panel if term is empty
                return;
            } else {
                //If opened, then close the search result panel
                this.isSearchResultsOpened = false;
                this.searchResultComponet.close();
                return;
            }
        }
        //Once this method is called
        //the search results page must be opened
        this.isSearchResultsOpened = true;

        //Call the child component to do the real work
        this.searchResultComponet.doSearch(event);
    }

    //Search results page closed
    //remove the related ovevriding things
    searchClose(): void {
        this.isSearchResultsOpened = false;
    }

    //Close serch result panel if existing
    watchClickEvt(): void {
        this.searchResultComponet.close();
        this.isSearchResultsOpened = false;
    }
}