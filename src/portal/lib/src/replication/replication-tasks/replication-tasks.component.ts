import { Component, OnInit, Input, OnDestroy } from '@angular/core';
import { Router } from '@angular/router';
import { ReplicationService } from "../../service/replication.service";
import { TranslateService } from '@ngx-translate/core';
import { finalize } from "rxjs/operators";
import { Subscription, timer } from "rxjs";
import { ErrorHandler } from "../../error-handler/error-handler";
import { ReplicationJob, ReplicationTasks, Comparator, ReplicationJobItem, State } from "../../service/interface";
import { CustomComparator, DEFAULT_PAGE_SIZE, calculatePage, doFiltering, doSorting } from "../../utils";
import { RequestQueryParams } from "../../service/RequestQueryParams";
const executionStatus = 'InProgress';
@Component({
  selector: 'replication-tasks',
  templateUrl: './replication-tasks.component.html',
  styleUrls: ['./replication-tasks.component.scss']
})
export class ReplicationTasksComponent implements OnInit, OnDestroy {
  isOpenFilterTag: boolean;
  inProgress: boolean = false;
  currentPage: number = 1;
  selectedRow: [];
  pageSize: number = DEFAULT_PAGE_SIZE;
  loading = true;
  searchTask: string;
  defaultFilter = "resource_type";
  tasks: ReplicationTasks;
  taskItem: ReplicationTasks[] = [];
  tasksCopy: ReplicationTasks[] = [];
  stopOnGoing: boolean;
  executions: ReplicationJobItem[];
  timerDelay: Subscription;
  @Input() executionId: string;
  startTimeComparator: Comparator<ReplicationJob> = new CustomComparator<
  ReplicationJob
  >("start_time", "date");
  endTimeComparator: Comparator<ReplicationJob> = new CustomComparator<
    ReplicationJob
  >("end_time", "date");

  constructor(
    private translate: TranslateService,
    private router: Router,
    private replicationService: ReplicationService,
    private errorHandler: ErrorHandler,
  ) { }

  ngOnInit(): void {
    this.searchTask = '';
    this.getExecutionDetail();
    this.clrLoadTasks();
  }

  getExecutionDetail(): void {
    this.inProgress = true;
    if (this.executionId) {
      this.replicationService.getExecutionById(this.executionId)
        .pipe(finalize(() => (this.inProgress = false)))
        .subscribe(res => {
          this.executions = res.data;
          this.clrLoadPage();
        },
        error => {
          this.errorHandler.error(error);
        });
    }
  }

  clrLoadPage(): void {
    if (!this.timerDelay) {
      this.timerDelay = timer(10000, 10000).subscribe(() => {
        let count: number = 0;
          if (this.executions['status'] === executionStatus) {
            count++;
          }
        if (count > 0) {
          this.getExecutionDetail();
          this.clrLoadTasks();
        } else {
          this.timerDelay.unsubscribe();
          this.timerDelay = null;
        }
      });
    }
  }

  public get trigger(): string {
    return this.executions && this.executions['trigger']
      ? this.executions['trigger']
      : "";
  }

  public get startTime(): Date {
    return this.executions && this.executions['start_time']
      ? this.executions['start_time']
      : null;
  }

  public get successNum(): string {
    return this.executions && this.executions['succeed'];
  }

  public get failedNum(): string {
    return this.executions && this.executions['failed'];
  }

  public get progressNum(): string {
    return this.executions && this.executions['in_progress'];
  }

  public get stoppedNum(): string {
    return this.executions && this.executions['stopped'];
  }

  stopJob() {
    this.stopOnGoing = true;
    this.replicationService.stopJobs(this.executionId)
    .subscribe(response => {
      this.stopOnGoing = false;
       this.getExecutionDetail();
       this.translate.get("REPLICATION.STOP_SUCCESS", { param: this.executionId }).subscribe((res: string) => {
          this.errorHandler.info(res);
       });
    },
    error => {
      this.errorHandler.error(error);
    });
  }

  viewLog(taskId: number | string): string {
    return this.replicationService.getJobBaseUrl() + "/executions/" + this.executionId + "/tasks/" + taskId + "/log";
  }

  ngOnDestroy() {
    if (this.timerDelay) {
      this.timerDelay.unsubscribe();
    }
  }

  clrLoadTasks(): void {
      this.loading = true;
      let params: RequestQueryParams = new RequestQueryParams();
      if (this.searchTask && this.searchTask !== "") {
        params.set(this.defaultFilter, this.searchTask);
      }
      this.replicationService.getReplicationTasks(this.executionId, params)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe(res => {
        this.tasks = res; // Keep the data
      },
      error => {
        this.errorHandler.error(error);
      });
  }
  onBack(): void {
    this.router.navigate(["harbor", "replications"]);
  }

  selectFilter($event: any): void {
    this.defaultFilter = $event['target'].value;
    this.doSearch(this.searchTask);
  }

  // refresh icon
  refreshTasks(): void {
    this.loading = true;
    this.currentPage = 1;
    this.replicationService.getReplicationTasks(this.executionId)
    .subscribe(res => {
      this.tasks = res;
      this.loading = false;
    },
    error => {
      this.loading = false;
      this.errorHandler.error(error);
    });
  }

  public doSearch(value: string): void {
    this.searchTask = value.trim();
    this.loading = true;
    this.clrLoadTasks();
  }

  openFilter(isOpen: boolean): void {
    if (isOpen) {
        this.isOpenFilterTag = true;
    } else {
        this.isOpenFilterTag = false;
    }
  }

}
