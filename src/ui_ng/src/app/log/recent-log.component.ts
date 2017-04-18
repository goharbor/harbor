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
import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';

import { AuditLog } from './audit-log';
import { SessionUser } from '../shared/session-user';

import { AuditLogService } from './audit-log.service';
import { SessionService } from '../shared/session.service';
import { MessageService } from '../global-message/message.service';
import { AlertType } from '../shared/shared.const';
import { errorHandler, accessErrorHandler } from '../shared/shared.utils';

@Component({
    selector: 'recent-log',
    templateUrl: './recent-log.component.html',
    styleUrls: ['recent-log.component.css']
})

export class RecentLogComponent implements OnInit {
    private sessionUser: SessionUser = null;
    private recentLogs: AuditLog[];
    private logsCache: AuditLog[];
    private onGoing: boolean = false;
    private lines: number = 10; //Support 10, 25 and 50
    currentTerm: string;

    constructor(
        private session: SessionService,
        private msgService: MessageService,
        private logService: AuditLogService) {
        this.sessionUser = this.session.getCurrentUser();//Initialize session
    }

    ngOnInit(): void {
        this.retrieveLogs();
    }

    private handleOnchange($event: any) {
        this.currentTerm = '';
        if ($event && $event.target && $event.target["value"]) {
            this.lines = $event.target["value"];
            if (this.lines < 10) {
                this.lines = 10;
            }
            this.retrieveLogs();
        }
    }

    public get logNumber(): number {
        return this.recentLogs?this.recentLogs.length:0;
    }

    public get inProgress(): boolean {
        return this.onGoing;
    }

    public doFilter(terms: string): void {
        if (terms.trim() === "") {
            this.recentLogs = this.logsCache.filter(log => log.username != "");
            return;
        }
        this.currentTerm = terms;
        this.recentLogs = this.logsCache.filter(log => this.isMatched(terms, log));
    }

    public refresh(): void {
        this.retrieveLogs();
    }

    private retrieveLogs(): void {
        if (this.lines < 10) {
            this.lines = 10;
        }

        this.onGoing = true;
        this.logService.getRecentLogs(this.lines)
            .subscribe(
            response => {
                this.onGoing = false;
                this.logsCache = response; //Keep the data
                this.recentLogs = this.logsCache.filter(log => log.username != "");//To display
            },
            error => {
                this.onGoing = false;
                if (!accessErrorHandler(error, this.msgService)) {
                    this.msgService.announceMessage(error.status, errorHandler(error), AlertType.DANGER);
                }
            }
            );
    }

    private isMatched(terms: string, log: AuditLog): boolean {
        let reg = new RegExp('.*' + terms + '.*', 'i');
        return reg.test(log.username) ||
            reg.test(log.repo_name) ||
            reg.test(log.operation) ||
            reg.test(log.repo_tag);
    }
}