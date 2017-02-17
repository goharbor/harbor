import { Component, OnInit, ViewChild } from '@angular/core';
import { Router } from '@angular/router';

import { SessionService } from '../../shared/session.service';
import { ModalEvent } from '../modal-event';

import { AccountSettingsModalComponent } from '../account-settings/account-settings-modal.component';

@Component({
    selector: 'harbor-shell',
    templateUrl: 'harbor-shell.component.html'
})

export class HarborShellComponent implements OnInit {

    @ViewChild(AccountSettingsModalComponent)
    private accountSettingsModal: AccountSettingsModalComponent;

    constructor(private session: SessionService) { }

    ngOnInit() {
        let cUser = this.session.getCurrentUser();
        if (!cUser) {
            //Try to update the session
            this.session.retrieveUser();
        }
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

    //Close the modal dialog
    closeModal(event: ModalEvent): void {
        switch (event.modalName) {
            case "account-settings":
                this.accountSettingsModal.close();
            default:
                break;
        }
    }
}