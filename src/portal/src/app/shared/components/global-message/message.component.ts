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
import {
    CommonRoutes,
    dismissInterval,
    httpStatusCode,
} from '../../entities/shared.const';
import { delUrlParam } from '../../units/utils';
import { UN_LOGGED_PARAM, YES } from '../../../account/sign-in/sign-in.service';
import { SessionService } from '../../services/session.service';

@Component({
    selector: 'global-message',
    templateUrl: 'message.component.html',
    styleUrls: ['message.component.scss'],
})
export class MessageComponent implements OnInit, OnDestroy {
    @Input() isAppLevel: boolean;
    globalMessage: Message = new Message();
    globalMessageOpened: boolean = false;
    messageText: string = '';
    timer: any = null;

    appLevelMsgSub: Subscription;
    msgSub: Subscription;
    clearSub: Subscription;

    constructor(
        private elementRef: ElementRef,
        private messageService: MessageService,
        private router: Router,
        private route: ActivatedRoute,
        private translate: TranslateService,
        private session: SessionService
    ) {}

    ngOnInit(): void {
        // Only subscribe application level message
        if (this.isAppLevel) {
            this.appLevelMsgSub =
                this.messageService.appLevelAnnounced$.subscribe(message => {
                    this.globalMessageOpened = true;
                    this.globalMessage = message;
                    this.checkLoginStatus();
                    this.messageText = message.message;

                    this.translateMessage(message);
                });
        } else {
            // Only subscribe general messages
            this.msgSub = this.messageService.messageAnnounced$.subscribe(
                message => {
                    this.globalMessageOpened = true;
                    this.globalMessage = message;
                    this.checkLoginStatus();
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

                    // Hack the Clarity Alert style with native dom
                    setTimeout(() => {
                        let nativeDom: any = this.elementRef.nativeElement;
                        let queryDoms: any[] =
                            nativeDom.getElementsByClassName('alert');
                        if (queryDoms && queryDoms.length > 0) {
                            let hackDom: any = queryDoms[0];
                            hackDom.className +=
                                ' alert-global alert-global-align';
                        }
                    }, 0);
                }
            );
        }

        this.clearSub = this.messageService.clearChan$.subscribe(clear => {
            this.onClose();
        });
    }

    ngOnDestroy() {
        if (this.appLevelMsgSub) {
            this.appLevelMsgSub.unsubscribe();
        }

        if (this.msgSub) {
            this.msgSub.unsubscribe();
        }

        if (this.clearSub) {
            this.clearSub.unsubscribe();
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

    public get needAuth(): boolean {
        return this.globalMessage
            ? this.globalMessage.statusCode === httpStatusCode.Unauthorized
            : false;
    }

    // Show message text
    public get message(): string {
        return this.messageText;
    }

    signIn(): void {
        // remove queryParam UN_LOGGED_PARAM of redirect url
        const url = delUrlParam(this.router.url, UN_LOGGED_PARAM);
        this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], {
            queryParams: { redirect_url: url },
        });
    }

    onClose() {
        if (this.timer) {
            clearTimeout(this.timer);
        }
        this.globalMessageOpened = false;
    }
    // if navigate from global search(un-logged users visit public project)
    isFromGlobalSearch(): boolean {
        return this.route.snapshot.queryParams[UN_LOGGED_PARAM] === YES;
    }
    checkLoginStatus() {
        if (this.globalMessage.statusCode === httpStatusCode.Unauthorized) {
            // User session timed out, then redirect to sign-in page
            if (
                this.session.getCurrentUser() &&
                !this.isSignInUrl() &&
                this.route.snapshot.queryParams[UN_LOGGED_PARAM] !== YES
            ) {
                const url = delUrlParam(this.router.url, UN_LOGGED_PARAM);
                this.session.clear(); // because of SignInGuard, must clear user session before navigating to sign-in page
                this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], {
                    queryParams: { redirect_url: url },
                });
            }
        }
    }
    isSignInUrl(): boolean {
        const url: string =
            this.router.url?.indexOf('?') === -1
                ? this.router.url
                : this.router.url?.split('?')[0];
        return url === CommonRoutes.EMBEDDED_SIGN_IN;
    }
}
