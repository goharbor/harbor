import { Injectable } from '@angular/core';
import { ConfirmationDialogService } from "../../global-confirmation-dialog/confirmation-dialog.service";
import { ConfirmationTargets } from "../../../shared/entities/shared.const";
import { ConfirmationMessage } from "../../global-confirmation-dialog/confirmation-message";

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

