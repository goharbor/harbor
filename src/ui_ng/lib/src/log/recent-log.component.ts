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
import {
    AccessLogService,
    AccessLog
} from '../service/index';
import { ErrorHandler } from '../error-handler/index';
import { Observable } from 'rxjs/Observable';
import { toPromise, CustomComparator } from '../utils';
import { LOG_TEMPLATE, LOG_STYLES } from './recent-log.template';

import { Comparator } from 'clarity-angular';

@Component({
    selector: 'hbr-log',
    styles: [LOG_STYLES],
    template: LOG_TEMPLATE
})

export class RecentLogComponent implements OnInit {
    recentLogs: AccessLog[];
    logsCache: AccessLog[];
    onGoing: boolean = false;
    lines: number = 10; //Support 10, 25 and 50
    currentTerm: string;

    loading: boolean;

    opTimeComparator: Comparator<AccessLog> = new CustomComparator<AccessLog>('op_time', 'date');

    constructor(
        private logService: AccessLogService,
        private errorHandler: ErrorHandler) { }

    ngOnInit(): void {
        this.retrieveLogs();
    }

    handleOnchange($event: any) {
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
        return this.recentLogs ? this.recentLogs.length : 0;
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

    retrieveLogs(): void {
        if (this.lines < 10) {
            this.lines = 10;
        }

        this.onGoing = true;
        this.loading = true;
        toPromise<AccessLog[]>(this.logService.getRecentLogs(this.lines))
            .then(response => {
                this.onGoing = false;
                this.loading = false;
                this.logsCache = response; //Keep the data
                this.recentLogs = this.logsCache.filter(log => log.username != "");//To display
            })
            .catch(error => {
                this.onGoing = false;
                this.loading = false;
                this.errorHandler.error(error);
            });
    }

    isMatched(terms: string, log: AccessLog): boolean {
        let reg = new RegExp('.*' + terms + '.*', 'i');
        return reg.test(log.username) ||
            reg.test(log.repo_name) ||
            reg.test(log.operation) ||
            reg.test(log.repo_tag);
    }
}