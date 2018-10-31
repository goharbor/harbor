import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../shared/confirmation-dialog/confirmation-message';
import { ConfirmationTargets } from '../shared/shared.const';

export function confirmUnsavedChanges(changes: any) {
    let confirmService = new ConfirmationDialogService();
    let msg = new ConfirmationMessage(
        'CONFIG.CONFIRM_TITLE',
        'CONFIG.CONFIRM_SUMMARY',
        '',
        changes,
        ConfirmationTargets.CONFIG
    );

    confirmService.openComfirmDialog(msg);
}

