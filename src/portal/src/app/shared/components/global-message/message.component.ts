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
import { Component, Input, OnInit, OnDestroy, ElementRef } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { Subscription } from 'rxjs';
import { TranslateService } from '@ngx-translate/core';
import { Message } from './message';
import { MessageService } from './message.service';
import { dismissInterval } from '../../entities/shared.const';

@Component({
    selector: 'global-message',
    templateUrl: 'message.component.html',
    styleUrls: ['message.component.scss'],
})
export class MessageComponent implements OnInit, OnDestroy {
    globalMessage: Message = new Message();
    globalMessageOpened: boolean = false;
    messageText: string = '';
    timer: any = null;
    msgSub: Subscription;
    constructor(
        private elementRef: ElementRef,
        private messageService: MessageService,
        private router: Router,
        private route: ActivatedRoute,
        private translate: TranslateService
    ) {}

    ngOnInit(): void {
        if (!this.msgSub) {
            this.msgSub = this.messageService.messageAnnounced$.subscribe(
                message => {
                    this.globalMessageOpened = true;
                    this.globalMessage = message;
                    this.messageText = message.message;
                    this.translateMessage(message);
                    // Make the message alert bar dismiss after several intervals.
                    // Only for this case
                    if (this.timer) {
                        clearTimeout(this.timer);
                        this.timer = null;
                    }
                    this.timer = setTimeout(
                        () => this.onClose(),
                        dismissInterval
                    );
                }
            );
        }
    }

    ngOnDestroy() {
        if (this.msgSub) {
            this.msgSub.unsubscribe();
            this.msgSub = null;
        }
    }

    // Translate or refactor the message shown to user
    translateMessage(msg: Message): void {
        let key = 'UNKNOWN_ERROR',
            param = '';
        if (msg && msg.message) {
            key =
                typeof msg.message === 'string'
                    ? msg.message.trim()
                    : msg.message;
            if (key === '') {
                key = 'UNKNOWN_ERROR';
            }
        }

        this.translate
            .get(key, { param: param })
            .subscribe((res: string) => (this.messageText = res));
    }
    // Show message text
    public get message(): string {
        return this.messageText;
    }
    onClose() {
        if (this.timer) {
            clearTimeout(this.timer);
        }
        this.globalMessageOpened = false;
    }
}
