import { Component, Input, Output, EventEmitter } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

import { errorHandler } from '../shared.utils';

@Component({
    selector: 'inline-alert',
    templateUrl: "inline-alert.component.html"
})
export class InlineAlertComponent {
    private inlineAlertType: string = 'alert-danger';
    private inlineAlertClosable: boolean = true;
    private alertClose: boolean = true;
    private displayedText: string = "";
    private showCancelAction: boolean = false;
    private useAppLevelStyle: boolean = false;

    @Output() confirmEvt = new EventEmitter<boolean>();

    constructor(private translate: TranslateService){}

    public get errorMessage(): string {
        return this.displayedText;
    }

    //Show error message inline
    public showInlineError(error: any): void {
        this.displayedText = errorHandler(error);

        this.inlineAlertType = 'alert-danger';
        this.showCancelAction = false;
        this.inlineAlertClosable = true;
        this.alertClose = false;
        this.useAppLevelStyle = false;
    }

    //Show confirmation info with action button
    public showInlineConfirmation(warning: any): void {
        this.displayedText = "";
        if(warning && warning.message){
            this.translate.get(warning.message).subscribe((res: string) => this.displayedText = res);
        }
        this.inlineAlertType = 'alert-warning';
        this.showCancelAction = true;
        this.inlineAlertClosable = true;
        this.alertClose = false;
        this.useAppLevelStyle = false;
    }

    //Show inline sccess info
    public showInlineSuccess(info: any): void {
        this.displayedText = "";
        if(info && info.message){
            this.translate.get(info.message).subscribe((res: string) => this.displayedText = res);
        }
        this.inlineAlertType = 'alert-success';
        this.showCancelAction = false;
        this.inlineAlertClosable = true;
        this.alertClose = false;
        this.useAppLevelStyle = false;
    }

    //Close alert
    public close(): void {
        this.alertClose = true;
    }

    private confirmCancel(): void {
        this.confirmEvt.emit(true);
    }
}