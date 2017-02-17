import { Component, Output, EventEmitter } from '@angular/core';
import { Router } from '@angular/router';

import { ModalEvent } from '../modal-event'; 

@Component({
    selector: 'navigator',
    templateUrl: "navigator.component.html"
})
export class NavigatorComponent {
    // constructor(private router: Router){}
    @Output() showAccountSettingsModal = new EventEmitter<ModalEvent>();

    //Open the account setting dialog
    open():void {
        this.showAccountSettingsModal.emit({
            modalName:"account-settings",
            modalFlag: true
        });
    }
}