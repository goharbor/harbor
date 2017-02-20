import { Component } from '@angular/core';

@Component({
    selector: "account-settings-modal",
    templateUrl: "account-settings-modal.component.html"
})

export class AccountSettingsModalComponent{
    opened:boolean = false;
    staticBackdrop: boolean = true;

    open() {
        this.opened = true;
    }

    close() {
        this.opened = false;
    }

    submit() {
        console.info("ok here!");
        this.close();
    }

}