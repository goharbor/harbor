import { Component, OnInit, Input } from '@angular/core';
import { Router } from '@angular/router';
import { ReplicationService } from "../../service/replication.service";
import { TranslateService } from '@ngx-translate/core';
import { finalize } from "rxjs/operators";
import { ErrorHandler } from "../../error-handler/error-handler";
import { ReplicationJob, ReplicationTasks, Comparator, ReplicationJobItem } from "../../service/interface";
import { CustomComparator } from "../../utils";
@Component({
  selector: 'replication-tasks',
  templateUrl: './replication-tasks.component.html',
  styleUrls: ['./replication-tasks.component.scss']
})
export class ReplicationTasksComponent implements OnInit {
  isOpenFilterTag: boolean;
  selectedRow: [];
  loading = false;
  searchTask: string;
  defaultFilter = "recourceType";
  tasks: ReplicationTasks[] = [];
  tasksCopy: ReplicationTasks[] = [];
  stopOnGoing: boolean;
  executions: ReplicationJobItem[];
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
    this.clrLoadTasks();
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

  stopJob() {
    this.stopOnGoing = true;
    this.replicationService.stopJobs(this.executionId)
    .subscribe(response => {
      this.stopOnGoing = false;
       // this.getExecutions();
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

  clrLoadTasks(): void {
      this.loading = true;
      this.replicationService.getReplicationTasks(this.executionId)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe(tasks => {
        if (this.defaultFilter === 'recourceType') {
            this.tasks = tasks.filter(x =>
              x.resource_type.includes(this.searchTask)
            );
        } else if (this.defaultFilter === 'recource') {
            this.tasks = tasks.filter(x =>
              x.src_resource.includes(this.searchTask)
            );
        } else if (this.defaultFilter === 'destination') {
            this.tasks = tasks.filter(x =>
              x.dst_resource.includes(this.searchTask)
            );
        } else {
            this.tasks = tasks.filter(x =>
              x.status.includes(this.searchTask)
            );
        }

        this.tasksCopy = tasks.map(x => Object.assign({}, x));
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
    this.searchTask = '';
    this.clrLoadTasks();
  }

  doSearch(value: string): void {
    if (!value) {
      return;
    }
    this.searchTask = value.trim();
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
