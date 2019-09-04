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
import { Component, OnInit, Input } from '@angular/core';
import { Comparator, State } from '../service/interface';

import {
    AccessLogService,
    AccessLog,
    AccessLogItem,
    RequestQueryParams
} from '../service/index';
import { ErrorHandler } from '../error-handler/index';
import { CustomComparator } from '../utils';
import {
    DEFAULT_PAGE_SIZE,
    calculatePage,
    doFiltering,
    doSorting
} from '../utils';

@Component({
    selector: 'hbr-log',
    templateUrl: './recent-log.component.html',
    styleUrls: ['./recent-log.component.scss']
})

export class RecentLogComponent implements OnInit {
    recentLogs: AccessLogItem[] = [];
    logsCache: AccessLog;
    loading: boolean = true;
    currentTerm: string;
    defaultFilter = "username";
    isOpenFilterTag: boolean;
    @Input() withTitle: boolean = false;

    pageSize: number = DEFAULT_PAGE_SIZE;
    currentPage: number = 1; // Double bound to pagination component
    currentPagePvt: number = 0; // Used to confirm whether page is changed
    currentState: State;

    opTimeComparator: Comparator<AccessLogItem> = new CustomComparator<AccessLogItem>('op_time', 'date');

    constructor(
        private logService: AccessLogService,
        private errorHandler: ErrorHandler) { }

    ngOnInit(): void {
    }

    public get totalCount(): number {
        return this.logsCache && this.logsCache.metadata ? this.logsCache.metadata.xTotalCount : 0;
    }

    public get inProgress(): boolean {
        return this.loading;
    }

    public doFilter(terms: string): void {

        // allow search by null characters
        if (terms === undefined || terms === null) {
            return;
        }
        this.currentTerm = terms.trim();
        // Trigger data loading and start from first page
        this.loading = true;
        this.currentPage = 1;
        if (this.currentPagePvt === 1) {
            // Force reloading
            let st: State = this.currentState;
            if (!st) {
                st = {
                    page: {}
                };
            }
            st.page.from = 0;
            st.page.to = this.pageSize - 1;
            st.page.size = this.pageSize;

            this.currentPagePvt = 0; // Reset pvt

            this.load(st);
        }
    }

    public refresh(): void {
        this.doFilter("");
    }

    openFilter(isOpen: boolean): void {
        if (isOpen) {
            this.isOpenFilterTag = true;
        } else {
            this.isOpenFilterTag = false;
        }
    }

    selectFilterKey($event: any): void {
        this.defaultFilter = $event['target'].value;
        this.doFilter(this.currentTerm);
    }

    load(state: State) {
        if (!state || !state.page) {
            return;
        }
        // Keep it for future filter
        this.currentState = state;

        let pageNumber: number = calculatePage(state);
        if (pageNumber !== this.currentPagePvt) {
            // load data
            let params: RequestQueryParams = new RequestQueryParams().set("page", '' + pageNumber).set("page_size", '' + this.pageSize);
            if (this.currentTerm && this.currentTerm !== "") {
                params = params.set(this.defaultFilter, this.currentTerm);
            }

            this.loading = true;
            this.logService.getRecentLogs(params)
                .subscribe(response => {
                    this.logsCache = response; // Keep the data
                    this.recentLogs = this.logsCache.data.filter(log => log.username !== ""); // To display

                    // Do customized filter
                    this.recentLogs = doFiltering<AccessLogItem>(this.recentLogs, state);

                    // Do customized sorting
                    this.recentLogs = doSorting<AccessLogItem>(this.recentLogs, state);

                    this.currentPagePvt = pageNumber;

                    this.loading = false;
                }, error => {
                    this.loading = false;
                    this.errorHandler.error(error);
                });
        } else {
            // Column sorting and filtering

            this.recentLogs = this.logsCache.data.filter(log => log.username !== ""); // Reset data

            // Do customized filter
            this.recentLogs = doFiltering<AccessLogItem>(this.recentLogs, state);

            // Do customized sorting
            this.recentLogs = doSorting<AccessLogItem>(this.recentLogs, state);
        }
    }
    isMatched(terms: string, log: AccessLogItem): boolean {
        let reg = new RegExp('.*' + terms + '.*', 'i');
        return reg.test(log.username) ||
            reg.test(log.repo_name) ||
            reg.test(log.operation) ||
            reg.test(log.repo_tag);
    }
}
