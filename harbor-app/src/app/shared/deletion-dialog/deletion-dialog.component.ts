import { Component, OnDestroy } from '@angular/core';
import { Subscription } from 'rxjs/Subscription';
import { TranslateService } from '@ngx-translate/core';

import { DeletionDialogService } from './deletion-dialog.service';
import { DeletionMessage } from './deletion-message';

@Component({
    selector: 'deletion-dialog',
    templateUrl: 'deletion-dialog.component.html',
    styleUrls: ['deletion-dialog.component.css']
})

export class DeletionDialogComponent implements OnDestroy{
    opened: boolean = false;
    dialogTitle: string = "";
    dialogContent: string = "";
    message: DeletionMessage;
    private annouceSubscription: Subscription;

    constructor(
        private delService: DeletionDialogService,
        private translate: TranslateService) {
        this.annouceSubscription = delService.deletionAnnouced$.subscribe(msg => {
            this.dialogTitle = msg.title;
            this.dialogContent = msg.message;
            this.message = msg;

            this.translate.get(this.dialogTitle).subscribe((res: string) => this.dialogTitle = res);
            this.translate.get(this.dialogContent, { 'param': msg.param }).subscribe((res: string) => this.dialogContent = res);
            //Open dialog
            this.open();
        });
    }

    ngOnDestroy(): void {
        if(this.annouceSubscription){
            this.annouceSubscription.unsubscribe();
        }
    }

    open(): void {
        this.opened = true;
    }

    close(): void {
        this.opened = false;
    }

    confirm(): void {
        this.delService.confirmDeletion(this.message);
        this.close();
    }
}