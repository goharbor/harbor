import { Injectable } from '@angular/core'
import { Subject } from 'rxjs/Subject';

import { DeletionMessage } from './deletion-message';

@Injectable()
export class DeletionDialogService {
    private deletionAnnoucedSource = new Subject<DeletionMessage>();
    private deletionConfirmSource = new Subject<DeletionMessage>();

    deletionAnnouced$ = this.deletionAnnoucedSource.asObservable();
    deletionConfirm$ = this.deletionConfirmSource.asObservable();

    confirmDeletion(message: any): void {
        this.deletionConfirmSource.next(message);
    }

    openComfirmDialog(message: DeletionMessage): void {
        this.deletionAnnoucedSource.next(message);
    }
}