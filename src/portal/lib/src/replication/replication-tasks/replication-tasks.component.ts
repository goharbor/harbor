import { Component, OnInit, Input } from '@angular/core';
import { Router } from '@angular/router';
import { ReplicationService } from "../../service/replication.service";
import { map, catchError } from "rxjs/operators";
import { Observable, forkJoin, throwError as observableThrowError } from "rxjs";
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
  tasks: ReplicationTasks[] = [];
  stopOnGoing: boolean;
  executions: string = 'InProgress';
  @Input() executionId: string;
  startTimeComparator: Comparator<ReplicationJob> = new CustomComparator<
  ReplicationJob
  >("start_time", "date");
  endTimeComparator: Comparator<ReplicationJob> = new CustomComparator<
    ReplicationJob
  >("end_time", "date");

  constructor(
    private router: Router,
    private replicationService: ReplicationService,
    private errorHandler: ErrorHandler,
  ) { }

  ngOnInit(): void {
    // this.getExecutions();
    this.getTasks();
    // this.executions.status = 'success';
  }

  // getExecutions(): void {
  //   if (this.executionId) {
  //     toPromise<ReplicationJob>(
  //       this.replicationService.getExecutions(this.executionId)
  //       )
  //       .then(executions => {
  //         console.log(executions);
  //       })
  //       .catch(error => {
  //           this.errorHandler.error(error);
  //       });
  //   }
  // }

  stopJob() {
    this.stopOnGoing = true;
    this.replicationService.stopJobs(this.executionId)
    .subscribe(res => {
      this.stopOnGoing = false;
       // this.getExecutions();
    },
    error => {
      this.errorHandler.error(error);
    });
  }

  viewLog(taskId: number | string): string {
    return this.replicationService.getJobBaseUrl() + "/" + this.executionId + "/tasks/" + taskId + "/log";
  }

  getTasks(): void {
      this.replicationService.getReplicationTasks(this.executionId)
      .subscribe(tasks => {
        this.tasks = tasks.map(x => Object.assign({}, x));
      },
      error => {
        this.errorHandler.error(error);
      });
  }
  onBack(): void {
    this.router.navigate(["harbor", "replications"]);
  }

  openFilter(isOpen: boolean): void {
    if (isOpen) {
        this.isOpenFilterTag = true;
    } else {
        this.isOpenFilterTag = false;
    }
}

}
