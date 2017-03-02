import { Component, Input, Output, EventEmitter } from '@angular/core';

import { errorHandler } from '../shared.utils';

@Component({
    selector: 'inline-alert',
    templateUrl: "inline-alert.component.html"
})
export class InlineAlertComponent {
    private inlineAlertType: string = 'alert-danger';
    private inlineAlertClosable: boolean = true;
    private alertClose: boolean = true;
    private errorText: string = "";
    private showCancelAction: boolean = false;
    private useAppLevelStyle: boolean = false;

    @Output() confirmEvt = new EventEmitter<boolean>();

    public get errorMessage(): string {
        return this.errorText;
    }

    //Show error message inline
    public showInlineError(error: any): void {
        this.errorText = errorHandler(error);

        this.inlineAlertType = 'alert-danger';
        this.showCancelAction = false;
        this.inlineAlertClosable = true;
        this.alertClose = false;
        this.useAppLevelStyle = false;
    }

    //Show confirmation info with action button
    public showInlineConfirmation(warning: any): void {
        this.errorText = errorHandler(warning);
        this.inlineAlertType = 'alert-warning';
        this.showCancelAction = true;
        this.inlineAlertClosable = true;
        this.alertClose = false;
        this.useAppLevelStyle = true;
    }

    //Close alert
    public close(): void {
        this.alertClose = true;
    }

    private confirmCancel(): void {
        this.confirmEvt.emit(true);
    }
}