import { Component, Output, EventEmitter } from '@angular/core';
import { Router } from '@angular/router';

import { ModalEvent } from '../modal-event';
import { SearchEvent } from '../search-event';

@Component({
    selector: 'navigator',
    templateUrl: "navigator.component.html"
})
export class NavigatorComponent {
    // constructor(private router: Router){}
    @Output() showAccountSettingsModal = new EventEmitter<ModalEvent>();
    @Output() searchEvt = new EventEmitter<SearchEvent>();

    //Open the account setting dialog
    open():void {
        this.showAccountSettingsModal.emit({
            modalName:"account-settings",
            modalFlag: true
        });
    }

    //Only transfer the search event to the parent shell
    transferSearchEvent(evt: SearchEvent): void {
        this.searchEvt.emit(evt);
    }
}