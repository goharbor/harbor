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
import { Component, OnDestroy } from '@angular/core';
import { Subscription } from 'rxjs';
import { TranslateService } from '@ngx-translate/core';
import { ConfirmationMessage } from './confirmation-message';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../shared/entities/shared.const';
import { ConfirmationDialogService } from './confirmation-dialog.service';
import { ConfirmationAcknowledgement } from './confirmation-state-message';

@Component({
    selector: 'global-confirmation-dialog',
    templateUrl: 'global-confirmation-dialog.component.html',
    styleUrls: ['global-confirmation-dialog.component.scss'],
})
export class GlobalConfirmationDialogComponent implements OnDestroy {
    opened: boolean = false;
    dialogTitle: string = '';
    dialogContent: string = '';
    message: ConfirmationMessage;
    annouceSubscription: Subscription;
    buttons: ConfirmationButtons;
    isDelete: boolean = false;
    constructor(
        private confirmationService: ConfirmationDialogService,
        private translate: TranslateService
    ) {
        this.annouceSubscription =
            confirmationService.confirmationAnnouced$.subscribe(msg => {
                this.dialogTitle = msg.title;
                this.dialogContent = msg.message;
                this.message = msg;
                this.translate
                    .get(this.dialogTitle)
                    .subscribe((res: string) => (this.dialogTitle = res));
                this.translate
                    .get(this.dialogContent, { param: msg.param })
                    .subscribe((res: string) => (this.dialogContent = res));
                // Open dialog
                this.buttons = msg.buttons;
                this.open();
            });
    }

    ngOnDestroy(): void {
        if (this.annouceSubscription) {
            this.annouceSubscription.unsubscribe();
        }
    }

    open(msg?: ConfirmationMessage): void {
        if (msg) {
            this.dialogTitle = msg.title;
            this.message = msg;
            this.translate
                .get(this.dialogTitle)
                .subscribe((res: string) => (this.dialogTitle = res));
            this.translate
                .get(msg.message, { param: msg.param })
                .subscribe((res: string) => {
                    this.dialogContent = res;
                });
            // Open dialog
            this.buttons = msg.buttons;
        }
        this.opened = true;
    }

    close(): void {
        this.opened = false;
    }

    cancel(): void {
        if (!this.message) {
            // Inproper condition
            this.close();
            return;
        }

        let data: any = this.message.data ? this.message.data : {};
        let target = this.message.targetId
            ? this.message.targetId
            : ConfirmationTargets.EMPTY;
        this.confirmationService.cancel(
            new ConfirmationAcknowledgement(
                ConfirmationState.CANCEL,
                data,
                target
            )
        );
        this.isDelete = false;
        this.close();
    }

    confirm(): void {
        if (!this.message) {
            // Inproper condition
            this.close();
            return;
        }

        let data: any = this.message.data ? this.message.data : {};
        let target = this.message.targetId
            ? this.message.targetId
            : ConfirmationTargets.EMPTY;
        this.confirmationService.confirm(
            new ConfirmationAcknowledgement(
                ConfirmationState.CONFIRMED,
                data,
                target
            )
        );
        this.close();
    }
}
