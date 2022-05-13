import { Component, Input, OnDestroy } from '@angular/core';
import { finalize } from 'rxjs/operators';
import { TagRetentionComponent } from '../../tag-retention.component';
import { TagRetentionService } from '../../tag-retention.service';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { PENDING, RUNNING, TIMEOUT } from '../../retention';

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
        this.tagRetentionService
            .getExecutionHistory(
                this.retentionId,
                this.executionId,
                this.page,
                this.pageSize
            )
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                (response: any) => {
                    // Get total count
                    if (response.headers) {
                        let xHeader: string =
                            response.headers.get('x-total-count');
                        if (xHeader) {
                            this.total = parseInt(xHeader, 0);
                        }
                    }
                    this.tasks = response.body as Array<any>;
                    TagRetentionComponent.calculateDuration(this.tasks);
                    this.loopGettingTasks();
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
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
                this.tagRetentionService
                    .getExecutionHistory(
                        this.retentionId,
                        this.executionId,
                        this.page,
                        this.pageSize
                    )
                    .pipe(finalize(() => (this.loading = false)))
                    .subscribe(res => {
                        // Get total count
                        if (res.headers) {
                            let xHeader: string =
                                res.headers.get('x-total-count');
                            if (xHeader) {
                                this.total = parseInt(xHeader, 0);
                            }
                        }
                        this.tasks = res.body as Array<any>;
                        TagRetentionComponent.calculateDuration(this.tasks);
                        this.loopGettingTasks();
                    });
            }, TIMEOUT);
        }
    }
}
