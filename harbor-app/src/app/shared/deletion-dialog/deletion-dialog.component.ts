import { Component } from '@angular/core';

import { DeletionDialogService } from './deletion-dialog.service';

@Component({
    selector: 'deletion-dialog',
    templateUrl: 'deletion-dialog.component.html',
    styleUrls: ['deletion-dialog.component.css']
})

export class DeletionDialogComponent{
    opened: boolean = false;
    dialogTitle: string = "";
    dialogContent: string = "";
    data: any;

    constructor(private delService: DeletionDialogService){
       delService.deletionAnnouced$.subscribe(msg => {
           this.dialogTitle = msg.title;
           this.dialogContent = msg.message;
           this.data = msg.data;
           //Open dialog
           this.open();
       });
    }

    open(): void {
        this.opened = true;
    }

    close(): void {
        this.opened = false;
    }

    confirm(): void {
        this.delService.confirmDeletion(this.data);
        this.close();
    }
}