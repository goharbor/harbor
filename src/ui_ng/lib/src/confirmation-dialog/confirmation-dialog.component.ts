// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import {Component, EventEmitter, Input, Output} from '@angular/core';
import { TranslateService } from '@ngx-translate/core';

import { ConfirmationMessage } from './confirmation-message';
import { ConfirmationAcknowledgement } from './confirmation-state-message';
import { ConfirmationState, ConfirmationTargets, ConfirmationButtons } from '../shared/shared.const';

import {BatchInfo} from './confirmation-batch-message';

@Component({
    selector: 'confirmation-dialog',
    templateUrl: './confirmation-dialog.component.html',
    styleUrls: [ './confirmation-dialog.component.scss' ]
})

export class ConfirmationDialogComponent {
    opened = false;
    dialogTitle = '';
    dialogContent = '';
    message: ConfirmationMessage;
    buttons: ConfirmationButtons;

    @Output() confirmAction = new EventEmitter<ConfirmationAcknowledgement>();
    @Output() cancelAction = new EventEmitter<ConfirmationAcknowledgement>();
    @Input() batchInfors: BatchInfo[]  = [];
    isDelete = false;

    constructor(
        private translate: TranslateService) {}

    open(msg: ConfirmationMessage): void {
        this.dialogTitle = msg.title;
        this.dialogContent = msg.message;
        this.message = msg;
        this.translate.get(this.dialogTitle).subscribe((res: string) => this.dialogTitle = res);
        this.translate.get(this.dialogContent, { 'param': msg.param }).subscribe((res: string) => this.dialogContent = res);
        // Open dialog
        this.buttons = msg.buttons;
        this.opened = true;
    }

    get batchOverStatus(): boolean {
        if (this.batchInfors.length) {
            return this.batchInfors.every(item => item.loading === false);
        }
        return false;
    }

    colorChange(list: BatchInfo) {
        if (!list.loading && !list.errorState) {
            return 'green';
        }else if (!list.loading && list.errorState) {
            return 'red';
        }else {
            return '#666';
        }
    }

    toggleErrorTitle(errorSpan: any) {
        errorSpan.style.display = (errorSpan.style.display === 'none') ? 'block' : 'none';
    }

    close(): void {
        this.batchInfors = [];
        this.opened = false;
    }

    cancel(): void {
        if (!this.message) {
            // Inproper condition
            this.close();
            return;
        }

        let data: any = this.message.data ? this.message.data : {};
        let target = this.message.targetId ? this.message.targetId : ConfirmationTargets.EMPTY;
        this.cancelAction.emit(new ConfirmationAcknowledgement(
            ConfirmationState.CANCEL,
            data,
            target
        ));
        this.isDelete = false;
        this.close();
    }

    operate(): void {
        if (!this.message) {// Inproper condition
            this.close();
            return;
        }

        if (this.batchInfors.length) {
            this.batchInfors.every(item => item.loading = true);
            this.isDelete = true;
        }

        let data: any = this.message.data ? this.message.data : {};
        let target = this.message.targetId ? this.message.targetId : ConfirmationTargets.EMPTY;
        let message = new ConfirmationAcknowledgement(
            ConfirmationState.CONFIRMED,
            data,
            target
        );
        this.confirmAction.emit(message);
    }

    confirm(): void {
        if (!this.message) {// Inproper condition
            this.close();
            return;
        }

        let data: any = this.message.data ? this.message.data : {};
        let target = this.message.targetId ? this.message.targetId : ConfirmationTargets.EMPTY;
        let message = new ConfirmationAcknowledgement(
            ConfirmationState.CONFIRMED,
            data,
            target
        );
        this.confirmAction.emit(message);
        this.close();
    }
}
