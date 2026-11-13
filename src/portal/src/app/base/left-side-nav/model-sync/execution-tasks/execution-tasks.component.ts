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
import { Component, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ClrDatagridStateInterface } from '@clr/angular';
import { TranslateService } from '@ngx-translate/core';
import { finalize } from 'rxjs/operators';
import { ModelSyncService } from '../../../../../../ng-swagger-gen/services/model-sync.service';
import { ModelSyncExecution } from '../../../../../../ng-swagger-gen/models/model-sync-execution';
import { ModelSyncTask } from '../../../../../../ng-swagger-gen/models/model-sync-task';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import {
    CURRENT_BASE_HREF,
    formatSize,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import {
    PAGE_SIZE_OPTIONS,
    REFRESH_TIME_DIFFERENCE,
} from '../../../../shared/entities/shared.const';
import { isExecutionInProgress, isTaskInProgress } from '../model-sync';

const STATUS_MAP = {
    Succeed: 'Succeeded',
};

@Component({
    selector: 'model-sync-execution-tasks',
    templateUrl: './execution-tasks.component.html',
    styleUrls: ['./execution-tasks.component.scss'],
    standalone: false,
})
export class ModelSyncExecutionTasksComponent implements OnInit, OnDestroy {
    clrPageSizeOptions: number[] = PAGE_SIZE_OPTIONS;
    executionId: number;
    execution: ModelSyncExecution;
    executionLoading: boolean = false;
    stopOnGoing: boolean = false;

    tasks: ModelSyncTask[] = [];
    loading: boolean = true;
    currentPage: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.MODEL_SYNC_TASKS_COMPONENT
    );
    totalCount: number = 0;
    refreshTimeout: any;

    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private modelSyncService: ModelSyncService,
        private errorHandler: ErrorHandler,
        private translate: TranslateService
    ) {}

    ngOnInit(): void {
        this.executionId = +this.route.snapshot.params['id'];
        this.loadExecution();
    }

    ngOnDestroy(): void {
        this.clearRefresh();
    }

    clearRefresh(): void {
        if (this.refreshTimeout) {
            clearTimeout(this.refreshTimeout);
            this.refreshTimeout = null;
        }
    }

    loadExecution(): void {
        this.executionLoading = true;
        this.modelSyncService
            .getModelSyncExecution({ executionId: this.executionId })
            .pipe(finalize(() => (this.executionLoading = false)))
            .subscribe(
                res => (this.execution = res),
                error => this.errorHandler.error(error)
            );
    }

    clrLoadTasks(withLoading: boolean, state: ClrDatagridStateInterface): void {
        if (!state || !state.page || !this.executionId) {
            return;
        }
        if (state.page.size) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.MODEL_SYNC_TASKS_COMPONENT,
                this.pageSize
            );
        }
        this.clearRefresh();
        if (withLoading) {
            this.loading = true;
        }
        this.modelSyncService
            .listModelSyncTasksResponse({
                executionId: this.executionId,
                page: this.currentPage,
                pageSize: this.pageSize,
                sort: getSortingString(state),
            })
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                res => {
                    const total = res.headers.get('X-Total-Count');
                    this.totalCount = total ? parseInt(total, 10) : 0;
                    this.tasks = res.body || [];
                    if (
                        this.tasks.some(t => isTaskInProgress(t)) ||
                        isExecutionInProgress(this.execution)
                    ) {
                        this.refreshTimeout = setTimeout(() => {
                            this.loadExecution();
                            this.clrLoadTasks(false, { page: {} });
                        }, REFRESH_TIME_DIFFERENCE);
                    }
                },
                error => this.errorHandler.error(error)
            );
    }

    refresh(): void {
        this.currentPage = 1;
        this.loadExecution();
        this.clrLoadTasks(true, { page: {} });
    }

    canStop(): boolean {
        return isExecutionInProgress(this.execution) && !this.stopOnGoing;
    }

    stop(): void {
        if (!this.canStop()) {
            return;
        }
        this.stopOnGoing = true;
        this.modelSyncService
            .stopModelSync({ executionId: this.executionId })
            .pipe(finalize(() => (this.stopOnGoing = false)))
            .subscribe(
                () => {
                    this.translate
                        .get('MODEL_SYNC.STOP_SUCCESS', {
                            param: this.executionId,
                        })
                        .subscribe(res => this.errorHandler.info(res));
                    this.refresh();
                },
                error => this.errorHandler.error(error)
            );
    }

    viewLog(taskId: number): string {
        return `${CURRENT_BASE_HREF}/model-sync/executions/${this.executionId}/tasks/${taskId}/log`;
    }

    onBack(): void {
        this.router.navigate(['harbor', 'model-sync']);
    }

    getStatusStr(status: string): string {
        return STATUS_MAP[status] || status;
    }

    size(bytes: number): string {
        return formatSize(String(bytes || 0));
    }
}
