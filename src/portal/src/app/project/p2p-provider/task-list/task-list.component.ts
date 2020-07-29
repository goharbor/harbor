import { Component, OnDestroy, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';
import { finalize } from 'rxjs/operators';
import { clone, CustomComparator, DEFAULT_PAGE_SIZE, isEmptyObject } from '../../../../lib/utils/utils';
import { Task } from '../../../../../ng-swagger-gen/models/task';
import { MessageHandlerService } from '../../../shared/message-handler/message-handler.service';
import { Project } from '../../project';
import { ClrDatagridComparatorInterface, UserPermissionService, USERSTATICPERMISSION } from '../../../../lib/services';
import { Execution } from '../../../../../ng-swagger-gen/models/execution';
import { PreheatService } from '../../../../../ng-swagger-gen/services/preheat.service';
import { EXECUTION_STATUS, P2pProviderService, TIME_OUT } from '../p2p-provider.service';
import { forkJoin, Observable } from 'rxjs';
import { ClrLoadingState } from '@clr/angular';

@Component({
  selector: 'task-list',
  templateUrl: './task-list.component.html',
  styleUrls: ['./task-list.component.scss']
})
export class TaskListComponent implements OnInit, OnDestroy {
  projectId: number;
  projectName: string;
  isOpenFilterTag: boolean;
  inProgress: boolean = false;
  currentPage: number = 1;
  pageSize: number = DEFAULT_PAGE_SIZE;
  totalCount: number;
  loading = true;
  tasks: Task[];
  stopOnGoing: boolean;
  executionId: string;
  preheatPolicyName: string;
  startTimeComparator: ClrDatagridComparatorInterface<Task> = new CustomComparator<Task>("start_time", "date");
  execution: Execution;
  hasUpdatePermission: boolean = false;
  btnState: ClrLoadingState = ClrLoadingState.DEFAULT;
  timeout: any;
  timeoutForTaskList: any;
  constructor(
    private translate: TranslateService,
    private router: Router,
    private route: ActivatedRoute,
    private messageHandlerService: MessageHandlerService,
    private preheatService: PreheatService,
    private p2pProviderService: P2pProviderService,
    private userPermissionService: UserPermissionService,
  ) { }

  ngOnInit(): void {
    this.projectId = +this.route.snapshot.parent.parent.params['id'];
    const resolverData = this.route.snapshot.parent.parent.data;
    if (resolverData) {
      let project = <Project>(resolverData["projectResolver"]);
      this.projectName = project.name;
    }
    this.executionId = this.route.snapshot.params['executionId'];
    this.preheatPolicyName = this.route.snapshot.params['preheatPolicyName'];
    if (this.executionId && this.preheatPolicyName && this.projectName) {
      this.getExecutionDetail(true);
    }
    this.getPermissions();
  }
  ngOnDestroy(): void {
    if (this.timeout) {
      clearTimeout(this.timeout);
      this.timeout = null;
    }
    if (this.timeoutForTaskList) {
      clearTimeout(this.timeoutForTaskList);
      this.timeoutForTaskList = null;
    }
  }
  getPermissions() {
    const permissionsList: Observable<boolean>[] = [];
    permissionsList.push(this.userPermissionService.getPermission(this.projectId,
      USERSTATICPERMISSION.P2P_PROVIDER.KEY, USERSTATICPERMISSION.P2P_PROVIDER.VALUE.UPDATE));
    this.btnState = ClrLoadingState.LOADING;
    forkJoin(...permissionsList).subscribe(Rules => {
      [this.hasUpdatePermission, ] = Rules;
      this.btnState = ClrLoadingState.SUCCESS;
    }, error => {
      this.messageHandlerService.error(error);
      this.btnState = ClrLoadingState.ERROR;
    });
  }
  getExecutionDetail(withLoading: boolean): void {
    if (withLoading) {
      this.inProgress = true;
    }
    if (this.executionId) {
      this.preheatService.GetExecution({
        projectName: this.projectName,
        preheatPolicyName: this.preheatPolicyName,
        executionId: +this.executionId
      }).pipe(finalize(() => (this.inProgress = false)))
        .subscribe(res => {
          this.execution = res;
            if (!this.execution || this.p2pProviderService.willChangStatus(this.execution.status)) {
              if (!this.timeout) {
                this.timeout = setTimeout(() => {
                  this.getExecutionDetail(false);
                }, TIME_OUT);
              }
            }
        },
        error => {
          this.messageHandlerService.error(error);
        });
    }
  }
  trigger(): string {
    return this.execution && this.execution.trigger
      ? this.execution.trigger
      : "";
  }

  startTime(): string {
    return this.execution && this.execution.start_time
      ? this.execution.start_time
      : null;
  }

  successNum(): number {
    if (this.execution && this.execution.metrics) {
      return this.execution.metrics.success_task_count ? this.execution.metrics.success_task_count : 0;
    }
    return 0;
  }

  failedNum(): number {
    if (this.execution && this.execution.metrics) {
      return this.execution.metrics.error_task_count ? this.execution.metrics.error_task_count : 0;
    }
    return 0;
  }

  progressNum(): number {
    if (this.execution && this.execution.metrics) {
      const num: number = (this.execution.metrics.pending_task_count ? this.execution.metrics.pending_task_count : 0)
        + (this.execution.metrics.running_task_count ? this.execution.metrics.running_task_count : 0)
        + (this.execution.metrics.scheduled_task_count ? this.execution.metrics.scheduled_task_count : 0);
      return num ? num : 0;
    }
    return 0;
  }

  stoppedNum(): number {
    if (this.execution && this.execution.metrics) {
      return this.execution.metrics.stopped_task_count ? this.execution.metrics.stopped_task_count : 0;
    }
    return 0;
  }

  stopJob() {
    this.stopOnGoing = true;
    const execution: Execution = clone(this.execution);
    execution.status = EXECUTION_STATUS.STOPPED;
    this.preheatService.StopExecution({
      projectName: this.projectName,
      preheatPolicyName: this.preheatPolicyName,
      executionId: +this.executionId,
      execution: execution
    })
    .subscribe(response => {
      this.stopOnGoing = false;
       this.getExecutionDetail(true);
       this.translate.get("REPLICATION.STOP_SUCCESS", { param: this.executionId }).subscribe((res: string) => {
          this.messageHandlerService.showSuccess(res);
       });
    },
    error => {
      this.messageHandlerService.error(error);
    });
  }

  viewLog(taskId: number | string): string {
    return this.preheatService.rootUrl
      + `/projects/${this.projectName}/preheat/policies/${this.preheatPolicyName}/executions/${this.executionId}/tasks/${taskId}/logs`;
  }
  clrLoadTasks(withLoading): void {
      if (withLoading) {
        this.loading = true;
      }
      this.preheatService.ListTasks({
        projectName: this.projectName,
        preheatPolicyName: this.preheatPolicyName,
        executionId: +this.executionId
      })
      .pipe(finalize(() => {
        this.loading = false;
      }))
        .subscribe(res => {
            this.tasks = res;
            if (this.tasks && this.tasks.length) {
              this.totalCount = this.tasks.length;
              for (let i = 0; i < this.tasks.length; i++) {
                if (this.p2pProviderService.willChangStatus(this.tasks[i].status)) {
                  if (!this.timeoutForTaskList) {
                    this.timeoutForTaskList = setTimeout(() => {
                      this.clrLoadTasks(false);
                    }, TIME_OUT);
                  }
                }
              }
            }
          },
      error => {
        this.messageHandlerService.error(error);
      });
  }
  onBack(): void {
    this.router.navigate(["harbor", "projects", `${this.projectId}`, "p2p-provider", "policies"]);
  }
  // refresh icon
  refreshTasks(): void {
    this.currentPage = 1;
    this.totalCount = 0;
    this.clrLoadTasks(true);
  }
  getDuration(t: Task): string {
    return this.p2pProviderService.getDuration(t.start_time, t.end_time);
  }
  isInProgress(): boolean {
    return this.execution && this.p2pProviderService.willChangStatus(this.execution.status);
  }
  isSuccess(): boolean {
    return this.execution && this.execution.status === EXECUTION_STATUS.SUCCESS;
  }
  isFailed(): boolean {
    return this.execution && (this.execution.status === EXECUTION_STATUS.ERROR || this.execution.status === EXECUTION_STATUS.STOPPED);
  }
  canStop(): boolean {
    return this.execution && this.p2pProviderService.willChangStatus(this.execution.status);
  }
}
