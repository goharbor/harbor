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
import { Component } from '@angular/core';
import { ErrorHandler } from '../../../shared/units/error-handler';
import { finalize } from 'rxjs/operators';
import { AuditlogService } from '../../../../../ng-swagger-gen/services/auditlog.service';
import { AuditLog } from '../../../../../ng-swagger-gen/models/audit-log';
import { ClrDatagridStateInterface } from '@clr/angular';
import {
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../shared/units/utils';
import ListAuditLogsParams = AuditlogService.ListAuditLogsParams;

@Component({
    selector: 'hbr-log',
    templateUrl: './recent-log.component.html',
    styleUrls: ['./recent-log.component.scss'],
})
export class RecentLogComponent {
    recentLogs: AuditLog[] = [];
    loading: boolean = true;
    currentTerm: string;
    defaultFilter = 'username';
    isOpenFilterTag: boolean;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.SYSTEM_RECENT_LOG_COMPONENT
    );
    currentPage: number = 1; // Double bound to pagination component
    totalCount: number = 0;

    constructor(
        private logService: AuditlogService,
        private errorHandler: ErrorHandler
    ) {}

    public get inProgress(): boolean {
        return this.loading;
    }

    public doFilter(terms: string): void {
        // allow search by null characters
        if (terms === undefined || terms === null) {
            return;
        }
        this.currentTerm = terms.trim();
        this.loading = true;
        this.currentPage = 1;
        this.totalCount = 0;
        this.load();
    }

    public refresh(): void {
        this.doFilter('');
    }

    openFilter(isOpen: boolean): void {
        this.isOpenFilterTag = isOpen;
    }

    selectFilterKey($event: any): void {
        this.defaultFilter = $event['target'].value;
        this.doFilter(this.currentTerm);
    }

    load(state?: ClrDatagridStateInterface) {
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.SYSTEM_RECENT_LOG_COMPONENT,
                this.pageSize
            );
        }
        // Keep it for future filter
        // this.currentState = state;
        const params: ListAuditLogsParams = {
            page: this.currentPage,
            pageSize: this.pageSize,
        };
        if (this.currentTerm && this.currentTerm !== '') {
            params.q = encodeURIComponent(
                `${this.defaultFilter}=~${this.currentTerm}`
            );
        }
        this.loading = true;
        this.logService
            .listAuditLogsResponse(params)
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                response => {
                    // Get total count
                    if (response.headers) {
                        let xHeader: string =
                            response.headers.get('x-total-count');
                        if (xHeader) {
                            this.totalCount = parseInt(xHeader, 0);
                        }
                    }
                    this.recentLogs = response.body as AuditLog[];
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }
}
