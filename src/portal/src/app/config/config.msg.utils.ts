import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../shared/confirmation-dialog/confirmation-message';
import { ConfirmationTargets } from '../shared/shared.const';
import { Injectable } from '@angular/core';

@Injectable()
export class ConfirmMessageHandler {
     constructor(private confirmService: ConfirmationDialogService) {
     }

     public confirmUnsavedChanges(changes: any) {
        let msg = new ConfirmationMessage(
            'CONFIG.CONFIRM_TITLE',
            'CONFIG.CONFIRM_SUMMARY',
            '',
            changes,
            ConfirmationTargets.CONFIG
        );
        this.confirmService.openComfirmDialog(msg);
     }
}

