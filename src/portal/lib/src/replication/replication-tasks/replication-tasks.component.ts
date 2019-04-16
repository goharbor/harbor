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
const taskStatus = 'InProgress';
@Component({
  selector: 'replication-tasks',
  templateUrl: './replication-tasks.component.html',
  styleUrls: ['./replication-tasks.component.scss']
})
export class ReplicationTasksComponent implements OnInit, OnDestroy {
  isOpenFilterTag: boolean;
  selectedRow: [];
  currentPage: number = 1;
  currentPagePvt: number = 0;
  totalCount: number = 0;
  pageSize: number = DEFAULT_PAGE_SIZE;
  currentState: State;
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
  }

  getExecutionDetail(): void {
    if (this.executionId) {
      this.replicationService.getExecutionById(this.executionId)
        .subscribe(res => {
          this.executions = res.data;
        },
        error => {
          this.errorHandler.error(error);
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

  clrLoadTasks(state: State): void {
      if (!state || !state.page) {
        return;
      }
      // Keep it for future filter
      this.currentState = state;

      let pageNumber: number = calculatePage(state);
      if (pageNumber !== this.currentPagePvt) {
        // load data
        let params: RequestQueryParams = new RequestQueryParams();
        params.set("page", '' + pageNumber);
        params.set("page_size", '' + this.pageSize);
        if (this.searchTask && this.searchTask !== "") {
            params.set(this.defaultFilter, this.searchTask);
        }

      this.loading = true;
      this.replicationService.getReplicationTasks(this.executionId, params)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe(res => {
        this.totalCount = res.length;
        this.tasks = res; // Keep the data
        this.taskItem = this.tasks.filter(tasks => tasks.resource_type !== "");
        if (!this.timerDelay) {
          this.timerDelay = timer(10000, 10000).subscribe(() => {
            let count: number = 0;
            this.tasks.forEach(tasks => {
              if (
                tasks.status === taskStatus
              ) {
                count++;
              }
            });
            if (count > 0) {
              this.clrLoadTasks(this.currentState);
            } else {
              this.timerDelay.unsubscribe();
              this.timerDelay = null;
            }
          });
        }
        this.taskItem = doFiltering<ReplicationTasks>(this.taskItem, state);

        this.taskItem = doSorting<ReplicationTasks>(this.taskItem, state);

        this.currentPagePvt = pageNumber;
      },
      error => {
        this.errorHandler.error(error);
      });
      } else {

        this.taskItem = this.tasks.filter(tasks => tasks.resource_type !== "");
        // Do customized filter
        this.taskItem = doFiltering<ReplicationTasks>(this.taskItem, state);

        // Do customized sorting
        this.taskItem = doSorting<ReplicationTasks>(this.taskItem, state);
      }
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
    this.searchTask = '';
    this.loading = true;
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
    if (!value) {
      return;
    }
    this.searchTask = value.trim();
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

        this.currentPagePvt = 0;

        this.clrLoadTasks(st);
    }
  }

  openFilter(isOpen: boolean): void {
    if (isOpen) {
        this.isOpenFilterTag = true;
    } else {
        this.isOpenFilterTag = false;
    }
  }

}
