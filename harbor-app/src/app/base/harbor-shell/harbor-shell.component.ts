import { Component, OnInit, ViewChild } from '@angular/core';
import { Router } from '@angular/router';

import { SessionService } from '../../shared/session.service';
import { ModalEvent } from '../modal-event';
import { SearchEvent } from '../search-event';

import { AccountSettingsModalComponent } from '../account-settings/account-settings-modal.component';
import { SearchResultComponent } from '../global-search/search-result.component';

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

    //To indicator whwther or not the search results page is displayed
    //We need to use this property to do some overriding work
    private isSearchResultsOpened: boolean = false;

    constructor(private session: SessionService) { }

    ngOnInit() {
        let cUser = this.session.getCurrentUser();
        if (!cUser) {
            //Try to update the session
            this.session.retrieveUser();
        }
    }

    public get showSearch(): boolean {
        return this.isSearchResultsOpened;
    }

    //Open modal dialog
    openModal(event: ModalEvent): void {
        switch (event.modalName) {
            case "account-settings":
                this.accountSettingsModal.open();
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
        if(event){
            this.isSearchResultsOpened = false;
        }
    }
}