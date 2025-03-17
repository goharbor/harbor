// Copyright Project Harbor Authors
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
import { Component, Input, OnDestroy } from '@angular/core';
import { finalize } from 'rxjs/operators';
import { TagRetentionComponent } from '../../tag-retention.component';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { PENDING, RUNNING, TIMEOUT } from '../../retention';
import { RetentionService } from '../../../../../../../../ng-swagger-gen/services/retention.service';
import { TagRetentionService } from '../../tag-retention.service';

@Component({
    selector: 'app-tag-retention-tasks',
    templateUrl: './tag-retention-tasks.component.html',
    styleUrls: ['./tag-retention-tasks.component.css'],
})
export class TagRetentionTasksComponent implements OnDestroy {
    @Input()
    retentionId;
    @Input()
    executionId;
    loading: boolean = true;
    page: number = 1;
    pageSize: number = 5;
    total: number = 0;
    tasks = [];
    tasksTimeout;
    constructor(
        private tagRetentionService: TagRetentionService,
        private retentionService: RetentionService,
        private errorHandler: ErrorHandler
    ) {}
    ngOnDestroy() {
        if (this.tasksTimeout) {
            clearTimeout(this.tasksTimeout);
            this.tasksTimeout = null;
        }
    }
    loadLog() {
        this.loading = true;
        this.retentionService
            .listRetentionTasksResponse({
                id: this.retentionId,
                eid: this.executionId,
                page: this.page,
                pageSize: this.pageSize,
            })
            .pipe(finalize(() => (this.loading = false)))
            .subscribe({
                next: res => {
                    this.handleResponse(res);
                },
                error: err => {
                    this.errorHandler.error(err);
                },
            });
    }
    seeLog(executionId, taskId) {
        this.tagRetentionService.seeLog(this.retentionId, executionId, taskId);
    }
    loopGettingTasks() {
        if (
            this.tasks &&
            this.tasks.length &&
            this.tasks.some(item => {
                return item.status === RUNNING || item.status === PENDING;
            })
        ) {
            this.tasksTimeout = setTimeout(() => {
                this.retentionService
                    .listRetentionTasksResponse({
                        id: this.retentionId,
                        eid: this.executionId,
                        page: this.page,
                        pageSize: this.pageSize,
                    })
                    .pipe(finalize(() => (this.loading = false)))
                    .subscribe(res => {
                        this.handleResponse(res);
                    });
            }, TIMEOUT);
        }
    }

    handleResponse(res: any) {
        // Get total count
        if (res.headers) {
            let xHeader: string = res.headers.get('x-total-count');
            if (xHeader) {
                this.total = parseInt(xHeader, 0);
            }
        }
        this.tasks = res.body as Array<any>;
        TagRetentionComponent.calculateDuration(this.tasks);
        this.loopGettingTasks();
    }
}
