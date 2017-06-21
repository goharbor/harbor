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
import { Router } from '@angular/router';
import {
    AccessLogService,
    AccessLog,
    AccessLogItem,
    RequestQueryParams
} from '../service/index';
import { ErrorHandler } from '../error-handler/index';
import { Observable } from 'rxjs/Observable';
import { toPromise, CustomComparator } from '../utils';
import { LOG_TEMPLATE, LOG_STYLES } from './recent-log.template';
import { DEFAULT_PAGE_SIZE } from '../utils';

import { Comparator, State } from 'clarity-angular';

@Component({
    selector: 'hbr-log',
    styles: [LOG_STYLES],
    template: LOG_TEMPLATE
})

export class RecentLogComponent implements OnInit {
    recentLogs: AccessLogItem[] = [];
    logsCache: AccessLog;
    loading: boolean = true;
    currentTerm: string;
    @Input() withTitle: boolean = false;

    pageSize: number = DEFAULT_PAGE_SIZE;
    currentPage: number = 1;//Double bound to pagination component
    currentPagePvt: number = 0; //Used to confirm whether page is changed
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
        this.currentTerm = terms.trim();
        //Trigger data loading and start from first page
        this.loading = true;
        this.currentPage = 1;
        if (this.currentPagePvt === 1) {
            //Force reloading
            let st: State = this.currentState;
            if (!st) {
                st = {
                    page: {}
                };
            }
            st.page.from = 0;
            st.page.to = this.pageSize - 1;
            st.page.size = this.pageSize;

            this.currentPagePvt = 0;//Reset pvt

            this.load(st);
        }
    }

    public refresh(): void {
        this.doFilter("");
    }

    load(state: State) {
        //Keep it for future filter
        this.currentState = state;

        let pageNumber: number = this._calculatePage(state);
        if (pageNumber !== this.currentPagePvt) {
            //load data
            let params: RequestQueryParams = new RequestQueryParams();
            params.set("page", '' + pageNumber);
            params.set("page_size", '' + this.pageSize);
            if (this.currentTerm && this.currentTerm !== "") {
                params.set('repository', this.currentTerm);
            }

            this.loading = true;
            toPromise<AccessLog>(this.logService.getRecentLogs(params))
                .then(response => {
                    this.logsCache = response; //Keep the data
                    this.recentLogs = this.logsCache.data.filter(log => log.username != "");//To display

                    //Do customized filter
                    this._doFilter(state);

                    //Do customized sorting
                    this._doSorting(state);

                    this.currentPagePvt = pageNumber;

                    this.loading = false;
                })
                .catch(error => {
                    this.loading = false;
                    this.errorHandler.error(error);
                });
        } else {
            //Column sorting and filtering

            this.recentLogs = this.logsCache.data.filter(log => log.username != "");//Reset data

            //Do customized filter
            this._doFilter(state);

            //Do customized sorting
            this._doSorting(state);
        }
    }

    isMatched(terms: string, log: AccessLogItem): boolean {
        let reg = new RegExp('.*' + terms + '.*', 'i');
        return reg.test(log.username) ||
            reg.test(log.repo_name) ||
            reg.test(log.operation) ||
            reg.test(log.repo_tag);
    }

    _calculatePage(state: State): number {
        if (!state || !state.page) {
            return 1;
        }

        return Math.ceil((state.page.to + 1) / state.page.size);
    }

    _doFilter(state: State): void {
        if (!this.recentLogs || this.recentLogs.length === 0) {
            return;
        }

        if (!state || !state.filters || state.filters.length === 0) {
            return;
        }

        state.filters.forEach((filter: {
            property: string;
            value: string;
        }) => {
            this.recentLogs = this.recentLogs.filter(logItem => this._regexpFilter(filter["value"], logItem[filter["property"]]));
        });
    }

    _regexpFilter(terms: string, testedValue: any): boolean {
        let reg = new RegExp('.*' + terms + '.*', 'i');
        return reg.test(testedValue);
    }

    _doSorting(state: State): void {
        if (!this.recentLogs || this.recentLogs.length === 0) {
            return;
        }

        if (!state || !state.sort) {
            return;
        }

        this.recentLogs = this.recentLogs.sort((a: AccessLogItem, b: AccessLogItem) => {
            let comp: number = 0;
            if (typeof state.sort.by !== "string") {
                comp = state.sort.by.compare(a, b);
            } else {
                let propA = a[state.sort.by.toString()], propB = b[state.sort.by.toString()];
                if (typeof propA === "string") {
                    comp = propA.localeCompare(propB);
                } else {
                    if (propA > propB) {
                        comp = 1;
                    } else if (propA < propB) {
                        comp = -1;
                    }
                }
            }

            if (state.sort.reverse) {
                comp = -comp;
            }

            return comp;
        });
    }
}