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
import {
  Component,
  OnInit,
  ViewChild,
  Input,
  Output,
  OnDestroy,
  EventEmitter
} from "@angular/core";
import { Comparator, State } from "../service/interface";
import { finalize, catchError, map } from "rxjs/operators";
import { Subscription, forkJoin, timer, Observable, throwError } from "rxjs";
import { TranslateService } from "@ngx-translate/core";

import { ListReplicationRuleComponent } from "../list-replication-rule/list-replication-rule.component";
import { CreateEditRuleComponent } from "../create-edit-rule/create-edit-rule.component";
import { ErrorHandler } from "../error-handler/error-handler";

import { ReplicationService } from "../service/replication.service";
import { RequestQueryParams } from "../service/RequestQueryParams";
import {
  ReplicationRule,
  ReplicationJob,
  ReplicationJobItem
} from "../service/interface";

import {
  CustomComparator,
  DEFAULT_PAGE_SIZE,
  doFiltering,
  doSorting,
  calculatePage
} from "../utils";

import {
  ConfirmationTargets,
  ConfirmationButtons,
  ConfirmationState
} from "../shared/shared.const";
import { ConfirmationMessage } from "../confirmation-dialog/confirmation-message";
import { ConfirmationDialogComponent } from "../confirmation-dialog/confirmation-dialog.component";
import { ConfirmationAcknowledgement } from "../confirmation-dialog/confirmation-state-message";
import {
  operateChanges,
  OperationState,
  OperateInfo
} from "../operation/operate";
import { OperationService } from "../operation/operation.service";
import { Router } from "@angular/router";
const ONE_HOUR_SECONDS: number = 3600;
const ONE_MINUTE_SECONDS: number = 60;
const ONE_DAY_SECONDS: number = 24 * ONE_HOUR_SECONDS;

const ruleStatus: { [key: string]: any } = [
  { key: "all", description: "REPLICATION.ALL_STATUS" },
  { key: "1", description: "REPLICATION.ENABLED" },
  { key: "0", description: "REPLICATION.DISABLED" }
];

export class SearchOption {
  ruleId: number | string;
  ruleName: string = "";
  trigger: string = "";
  status: string = "";
  page: number = 1;
  pageSize: number = DEFAULT_PAGE_SIZE;
}

@Component({
  selector: "hbr-replication",
  templateUrl: "./replication.component.html",
  styleUrls: ["./replication.component.scss"]
})
export class ReplicationComponent implements OnInit, OnDestroy {
  @Input() projectId: number | string;
  @Input() projectName: string;
  @Input() isSystemAdmin: boolean;
  @Input() withAdmiral: boolean;
  @Input() withReplicationJob: boolean;
  @Input() hasCreateReplicationPermission: boolean;
  @Input() hasUpdateReplicationPermission: boolean;
  @Input() hasDeleteReplicationPermission: boolean;
  @Input() hasExecuteReplicationPermission: boolean;

  @Output() redirect = new EventEmitter<ReplicationRule>();
  @Output() openCreateRule = new EventEmitter<any>();
  @Output() openEdit = new EventEmitter<string | number>();
  @Output() goToRegistry = new EventEmitter<any>();

  search: SearchOption = new SearchOption();
  isOpenFilterTag: boolean;
  ruleStatus = ruleStatus;
  currentRuleStatus: { key: string; description: string };

  currentTerm: string;
  defaultFilter = "trigger";

  changedRules: ReplicationRule[];

  selectedRow: ReplicationJobItem[] = [];
  rules: ReplicationRule[];
  loading: boolean;
  isStopOnGoing: boolean;
  hiddenJobList = true;

  jobs: ReplicationJobItem[];

  @ViewChild(ListReplicationRuleComponent)
  listReplicationRule: ListReplicationRuleComponent;

  @ViewChild(CreateEditRuleComponent)
  createEditPolicyComponent: CreateEditRuleComponent;

  @ViewChild("replicationConfirmDialog")
  replicationConfirmDialog: ConfirmationDialogComponent;

  @ViewChild("StopConfirmDialog")
  StopConfirmDialog: ConfirmationDialogComponent;

  creationTimeComparator: Comparator<ReplicationJob> = new CustomComparator<
    ReplicationJob
  >("start_time", "date");
  updateTimeComparator: Comparator<ReplicationJob> = new CustomComparator<
    ReplicationJob
  >("end_time", "date");

  // Server driven pagination
  currentPage: number = 1;
  totalCount: number = 0;
  pageSize: number = DEFAULT_PAGE_SIZE;
  currentState: State;
  jobsLoading: boolean = false;
  timerDelay: Subscription;

  constructor(
    private router: Router,
    private errorHandler: ErrorHandler,
    private replicationService: ReplicationService,
    private operationService: OperationService,
    private translateService: TranslateService
  ) {}

  public get showPaginationIndex(): boolean {
    return this.totalCount > 0;
  }

  ngOnInit() {
    this.currentRuleStatus = this.ruleStatus[0];
  }

  ngOnDestroy() {
    if (this.timerDelay) {
      this.timerDelay.unsubscribe();
    }
  }

  // open replication rule
  openModal(): void {
    this.createEditPolicyComponent.openCreateEditRule();
  }

  // edit replication rule
  openEditRule(rule: ReplicationRule) {
    if (rule) {
      this.createEditPolicyComponent.openCreateEditRule(rule.id);
    }
  }

  goRegistry(): void {
    this.goToRegistry.emit();
  }

  goToLink(exeId: number): void {
    let linkUrl = ["harbor", "replications", exeId, "tasks"];
    this.router.navigate(linkUrl);
  }

  // Server driven data loading
  clrLoadJobs(state: State): void {
    if (!state || !state.page || !this.search.ruleId) {
      return;
    }
    this.currentState = state;

    let pageNumber: number = calculatePage(state);
    if (pageNumber <= 0) {
      pageNumber = 1;
    }

    let params: RequestQueryParams = new RequestQueryParams();
    // Pagination
    params.set("page", "" + pageNumber);
    params.set("page_size", "" + this.pageSize);

    if (this.currentTerm && this.currentTerm !== "") {
      params.set(this.defaultFilter, this.currentTerm);
    }

    this.jobsLoading = true;

    this.replicationService.getExecutions(this.search.ruleId, params).subscribe(
      response => {
        this.totalCount = response.metadata.xTotalCount;
        this.jobs = response.data;
        if (!this.timerDelay) {
          this.timerDelay = timer(10000, 10000).subscribe(() => {
            let count: number = 0;
            this.jobs.forEach(job => {
              if (
                job.status === "InProgress"
              ) {
                count++;
              }
            });
            if (count > 0) {
              this.clrLoadJobs(this.currentState);
            } else {
              this.timerDelay.unsubscribe();
              this.timerDelay = null;
            }
          });
        }
        // Do filtering and sorting
        this.jobs = doFiltering<ReplicationJobItem>(this.jobs, state);
        this.jobs = doSorting<ReplicationJobItem>(this.jobs, state);

        this.jobsLoading = false;
      },
      error => {
        this.jobsLoading = false;
        this.errorHandler.error(error);
      }
    );
  }
  public doSearchExecutions(terms: string): void {
    if (!terms) {
      return;
    }
    this.currentTerm = terms.trim();
    // Trigger data loading and start from first page
    this.jobsLoading = true;
    this.currentPage = 1;
    this.jobsLoading = true;
    // Force reloading
    this.loadFirstPage();
  }
  loadFirstPage(): void {
    let st: State = this.currentState;
    if (!st) {
      st = {
        page: {}
      };
    }
    st.page.size = this.pageSize;
    st.page.from = 0;
    st.page.to = this.pageSize - 1;

    this.clrLoadJobs(st);
  }

  selectOneRule(rule: ReplicationRule) {
    if (rule && rule.id) {
      this.hiddenJobList = false;
      this.search.ruleId = rule.id || "";
      this.loadFirstPage();
    }
  }

  replicateManualRule(rule: ReplicationRule) {
    if (rule) {
      let replicationMessage = new ConfirmationMessage(
        "REPLICATION.REPLICATION_TITLE",
        "REPLICATION.REPLICATION_SUMMARY",
        rule.name,
        rule,
        ConfirmationTargets.TARGET,
        ConfirmationButtons.REPLICATE_CANCEL
      );
      this.replicationConfirmDialog.open(replicationMessage);
    }
  }

  confirmReplication(message: ConfirmationAcknowledgement) {
    if (
      message &&
      message.source === ConfirmationTargets.TARGET &&
      message.state === ConfirmationState.CONFIRMED
    ) {
      let rule: ReplicationRule = message.data;

      if (rule) {
        forkJoin(this.replicationOperate(rule)).subscribe(item => {
          this.selectOneRule(rule);
        });
      }
    }
  }

  replicationOperate(rule: ReplicationRule): Observable<any> {
    // init operation info
    let operMessage = new OperateInfo();
    operMessage.name = "OPERATION.REPLICATION";
    operMessage.data.id = rule.id;
    operMessage.state = OperationState.progressing;
    operMessage.data.name = rule.name;
    this.operationService.publishInfo(operMessage);

    return this.replicationService.replicateRule(+rule.id).pipe(
      map(response => {
        this.translateService
          .get("BATCH.REPLICATE_SUCCESS")
          .subscribe(res =>
            operateChanges(operMessage, OperationState.success)
          );
      }),
      catchError(error => {
        if (error && error.status === 412) {
          return forkJoin(
            this.translateService.get("BATCH.REPLICATE_FAILURE"),
            this.translateService.get("REPLICATION.REPLICATE_SUMMARY_FAILURE")
          ).pipe(
            map(function(res) {
              operateChanges(operMessage, OperationState.failure, res[1]);
            })
          );
        } else {
          return this.translateService.get("BATCH.REPLICATE_FAILURE").pipe(
            map(res => {
              operateChanges(operMessage, OperationState.failure, res);
            })
          );
        }
      })
    );
  }

  customRedirect(rule: ReplicationRule) {
    this.redirect.emit(rule);
  }

  doSearchRules(ruleName: string) {
    this.search.ruleName = ruleName;
    this.listReplicationRule.retrieveRules(ruleName);
  }

  doFilterJob($event: any): void {
    this.defaultFilter = $event["target"].value;
    this.doSearchJobs(this.currentTerm);
  }

  doSearchJobs(terms: string) {
    if (!terms) {
      return;
    }
    this.currentTerm = terms.trim();
    this.currentPage = 1;
    this.loadFirstPage();
  }

  hideJobs() {
    this.search.ruleId = 0;
    this.jobs = [];
    this.hiddenJobList = true;
  }

  openStopExecutionsDialog(targets: ReplicationJobItem[]) {
    let ExecutionId = targets.map(robot => robot.id).join(",");
    let StopExecutionsMessage = new ConfirmationMessage(
      "REPLICATION.STOP_TITLE",
      "REPLICATION.STOP_SUMMARY",
      ExecutionId,
      targets,
      ConfirmationTargets.STOP_EXECUTIONS,
      ConfirmationButtons.STOP_CANCEL
    );
    this.StopConfirmDialog.open(StopExecutionsMessage);
  }

  confirmStop(message: ConfirmationAcknowledgement) {
    if (
      message &&
      message.state === ConfirmationState.CONFIRMED &&
      message.source === ConfirmationTargets.STOP_EXECUTIONS
    ) {
      this.StopExecutions(message.data);
    }
  }

  StopExecutions(targets: ReplicationJobItem[]): void {
    if (targets && targets.length < 1) {
      return;
    }

    this.isStopOnGoing = true;
    if (this.jobs && this.jobs.length) {
      let ExecutionsStop$ = targets.map(target => this.StopOperate(target));
      forkJoin(ExecutionsStop$)
        .pipe(
          catchError(err => throwError(err)),
          finalize(() => {
            this.refreshJobs();
            this.isStopOnGoing = false;
            this.selectedRow = [];
          })
        )
        .subscribe(() => {});
    }
  }

  StopOperate(targets: ReplicationJobItem): any {
    let operMessage = new OperateInfo();
    operMessage.name = "OPERATION.STOP_EXECUTIONS";
    operMessage.data.id = targets.id;
    operMessage.state = OperationState.progressing;
    operMessage.data.name = targets.id;
    this.operationService.publishInfo(operMessage);

    return this.replicationService
      .stopJobs(targets.id)
      .pipe(
        map(
          () => operateChanges(operMessage, OperationState.success),
          err => operateChanges(operMessage, OperationState.failure, err)
        )
      );
  }

  reloadRules(isReady: boolean) {
    if (isReady) {
      this.search.ruleName = "";
      this.listReplicationRule.retrieveRules(this.search.ruleName);
    }
  }

  refreshRules() {
    this.listReplicationRule.retrieveRules();
  }

  refreshJobs() {
    this.currentTerm = "";
    this.currentPage = 1;

    let st: State = {
      page: {
        from: 0,
        to: this.pageSize - 1,
        size: this.pageSize
      }
    };
    this.clrLoadJobs(st);
  }

  openFilter(isOpen: boolean): void {
    if (isOpen) {
      this.isOpenFilterTag = true;
    } else {
      this.isOpenFilterTag = false;
    }
  }
  getDuration(j: ReplicationJobItem) {
    if (!j) {
      return;
    }

    let start_time = new Date(j.start_time).getTime();
    let end_time = new Date(j.end_time).getTime();
    let timesDiff = end_time - start_time;
    let timesDiffSeconds = timesDiff / 1000;
    let minutes = Math.floor(((timesDiffSeconds % ONE_DAY_SECONDS) % ONE_HOUR_SECONDS) / ONE_MINUTE_SECONDS);
    let seconds = Math.floor(timesDiffSeconds % ONE_MINUTE_SECONDS);
    if (minutes > 0) {
      return minutes + "m" + seconds + "s";
    }

    if (seconds > 0) {
      return seconds + "s";
    }

    if (seconds <= 0 && timesDiff > 0) {
      return timesDiff + 'ms';
    }
  }
}
