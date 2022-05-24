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
import { Injectable } from '@angular/core';
import { Subject } from 'rxjs';
import { Message } from './message';
import { AlertType } from '../../entities/shared.const';

@Injectable({
    providedIn: 'root',
})
export class MessageService {
    messageAnnouncedSource = new Subject<Message>();
    appLevelAnnouncedSource = new Subject<Message>();
    clearSource = new Subject<boolean>();

    messageAnnounced$ = this.messageAnnouncedSource.asObservable();
    appLevelAnnounced$ = this.appLevelAnnouncedSource.asObservable();
    clearChan$ = this.clearSource.asObservable();

    announceMessage(statusCode: number, message: string, alertType: AlertType) {
        this.messageAnnouncedSource.next(
            Message.newMessage(statusCode, message, alertType)
        );
    }

    announceAppLevelMessage(
        statusCode: number,
        message: string,
        alertType: AlertType
    ) {
        this.appLevelAnnouncedSource.next(
            Message.newMessage(statusCode, message, alertType)
        );
    }

    clear() {
        this.clearSource.next(true);
    }
}
