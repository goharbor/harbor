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
  Input,
  Output,
  OnInit,
  EventEmitter,
  ViewChild,
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  OnChanges,
  SimpleChange,
  SimpleChanges
} from '@angular/core';

import { ReplicationService } from '../service/replication.service';
import {ReplicationJob, ReplicationJobItem, ReplicationRule} from '../service/interface';

import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ConfirmationMessage } from '../confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../confirmation-dialog/confirmation-state-message';

import { ConfirmationState, ConfirmationTargets, ConfirmationButtons } from '../shared/shared.const';

import { TranslateService } from '@ngx-translate/core';

import { ErrorHandler } from '../error-handler/error-handler';
import { toPromise, CustomComparator } from '../utils';

import { State, Comparator } from 'clarity-angular';

import { LIST_REPLICATION_RULE_TEMPLATE } from './list-replication-rule.component.html';
import {BatchInfo, BathInfoChanges} from "../confirmation-dialog/confirmation-batch-message";
import {Observable} from "rxjs/Observable";

@Component({
  selector: 'hbr-list-replication-rule',
  template: LIST_REPLICATION_RULE_TEMPLATE,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListReplicationRuleComponent implements OnInit, OnChanges {

  nullTime: string = '0001-01-01T00:00:00Z';

  @Input() projectId: number;
  @Input() selectedId: number | string;
  @Input() withReplicationJob: boolean;
  @Input() readonly: boolean;

  @Input() loading: boolean = false;

  @Output() reload = new EventEmitter<boolean>();
  @Output() selectOne = new EventEmitter<ReplicationRule>();
  @Output() editOne = new EventEmitter<ReplicationRule>();
  @Output() toggleOne = new EventEmitter<ReplicationRule>();
  @Output() redirect = new EventEmitter<ReplicationRule>();
  @Output() openNewRule = new EventEmitter<any>();
  @Output() replicateManual = new EventEmitter<ReplicationRule>();

  projectScope: boolean = false;

  rules: ReplicationRule[];
  changedRules: ReplicationRule[];
  ruleName: string;
  canDeleteRule: boolean;

  selectedRow: ReplicationRule;
  batchDelectionInfos: BatchInfo[] = [];

  @ViewChild('toggleConfirmDialog')
  toggleConfirmDialog: ConfirmationDialogComponent;

  @ViewChild('deletionConfirmDialog')
  deletionConfirmDialog: ConfirmationDialogComponent;

  startTimeComparator: Comparator<ReplicationRule> = new CustomComparator<ReplicationRule>('start_time', 'date');
  enabledComparator: Comparator<ReplicationRule> = new CustomComparator<ReplicationRule>('enabled', 'number');

  constructor(
    private replicationService: ReplicationService,
    private translateService: TranslateService,
    private errorHandler: ErrorHandler,
    private ref: ChangeDetectorRef) {
    setInterval(() => ref.markForCheck(), 500);
  }

  public get creationAvailable(): boolean {
    return !this.readonly && !this.projectId ? true : false;
  }


  ngOnInit(): void {
    //Global scope
    if (!this.projectScope) {
      this.retrieveRules();
    }
  }

  ngOnChanges(changes: SimpleChanges): void {
    let proIdChange: SimpleChange = changes["projectId"];
    if (proIdChange) {
      if (proIdChange.currentValue !== proIdChange.previousValue) {
        if (proIdChange.currentValue) {
          this.projectId = proIdChange.currentValue;
          this.projectScope = true; //Scope is project, not global list
          //Initially load the replication rule data
          this.retrieveRules();
        }
      }
    }
  }

  selectedChange(): void {
    let hnd = setInterval(() => this.ref.markForCheck(), 200);
    setTimeout(() => clearInterval(hnd), 2000);
  }

  retrieveRules(ruleName: string = ''): void {
    this.loading = true;
    toPromise<ReplicationRule[]>(this.replicationService
      .getReplicationRules(this.projectId, ruleName))
      .then(rules => {
        this.rules = rules || [];
        if (this.rules && this.rules.length > 0) {
          this.selectedId = this.rules[0].id || '';
          this.selectOne.emit(this.rules[0]);
        }
        this.changedRules = this.rules;
        this.loading = false;
      }
      ).catch(error => {
        this.errorHandler.error(error);
        this.loading = false;
      });
  }

  filterRuleStatus(status: string) {
    if (status === 'all') {
      this.changedRules = this.rules;
    } else {
      this.changedRules = this.rules.filter(policy => policy.enabled === +status);
    }
  }

  toggleConfirm(message: ConfirmationAcknowledgement) {
    if (message &&
        message.source === ConfirmationTargets.TOGGLE_CONFIRM &&
        message.state === ConfirmationState.CONFIRMED) {
      this.batchDelectionInfos = [];
      let rule: ReplicationRule = message.data;
      let initBatchMessage = new BatchInfo ();
      initBatchMessage.name = rule.name;
      this.batchDelectionInfos.push(initBatchMessage);

      if (rule) {
        rule.enabled = rule.enabled === 0 ? 1 : 0;
        toPromise<any>(this.replicationService
            .enableReplicationRule(rule.id || '', rule.enabled))
            .then(() =>
                this.translateService.get('REPLICATION.TOGGLED_SUCCESS')
                    .subscribe(res => this.batchDelectionInfos[0].status = res))
            .catch(error => this.batchDelectionInfos[0].status = error);
      }
    }
  }

  deletionConfirm(message: ConfirmationAcknowledgement) {
    if (message &&
      message.source === ConfirmationTargets.POLICY &&
      message.state === ConfirmationState.CONFIRMED) {
      let rule: ReplicationRule = message.data;
      toPromise<any>(this.replicationService
        .deleteReplicationRule(rule.id))
        .then(() => {
          this.translateService.get('BATCH.DELETED_SUCCESS')
            .subscribe(res => {
              this.batchDelectionInfos[0] = BathInfoChanges(this.batchDelectionInfos[0], res);
            });
          this.reload.emit(true);
        })
        .catch(error => {
          if (error && error.status === 412) {
            Observable.forkJoin(this.translateService.get('BATCH.DELETED_FAILURE'),
                this.translateService.get('REPLICATION.FAILED_TO_DELETE_POLICY_ENABLED')).subscribe(res => {
              this.batchDelectionInfos[0] = BathInfoChanges(this.batchDelectionInfos[0], res[0], false, true, res[1]);
            });
          } else {
            this.translateService.get('BATCH.DELETED_FAILURE').subscribe(res => {
              this.batchDelectionInfos[0] = BathInfoChanges(this.batchDelectionInfos[0], res, false, true);
            });
          }
        });
    }
  }

  selectRule(rule: ReplicationRule): void {
    this.selectedId = rule.id || '';
    this.selectOne.emit(rule);
  }

  redirectTo(rule: ReplicationRule): void {
    this.redirect.emit(rule);
  }

  openModal(): void {
    this.openNewRule.emit();
  }

  editRule(rules: ReplicationRule) {
    this.editOne.emit(rules);
  }

  replicateRule(rule: ReplicationRule) {
    this.replicateManual.emit(rule);
  }

  toggleRule(rule: ReplicationRule) {
    let toggleConfirmMessage: ConfirmationMessage = new ConfirmationMessage(
      rule.enabled === 1 ? 'REPLICATION.TOGGLE_DISABLE_TITLE' : 'REPLICATION.TOGGLE_ENABLE_TITLE',
      rule.enabled === 1 ? 'REPLICATION.CONFIRM_TOGGLE_DISABLE_POLICY' : 'REPLICATION.CONFIRM_TOGGLE_ENABLE_POLICY',
      rule.name || '',
      rule,
      ConfirmationTargets.TOGGLE_CONFIRM
    );
    this.toggleConfirmDialog.open(toggleConfirmMessage);
  }

  jobList(): Promise<void> {
    let ruleData: ReplicationJobItem[];
    this.canDeleteRule = true;
    let count: number = 0;
    return toPromise<ReplicationJob>(this.replicationService
        .getJobs(this.selectedId))
        .then(response => {
          ruleData = response.data;
          if (ruleData.length) {
            ruleData.forEach(job => {
              if ((job.status === 'pending') || (job.status === 'running') || (job.status === 'retrying')) {
                count ++;
              }
            });
          }
          this.canDeleteRule = count > 0 ? false : true;
        })
        .catch(error => this.errorHandler.error(error));
  }

  deleteRule(rule: ReplicationRule) {
    this.jobList().then(() => {
      let deletionMessage: ConfirmationMessage;
      if (!this.canDeleteRule) {
        deletionMessage = new ConfirmationMessage(
            'REPLICATION.DELETION_TITLE_FAILURE',
            'REPLICATION.DELETION_SUMMARY_FAILURE',
            rule.name || '',
            rule,
            ConfirmationTargets.POLICY,
            ConfirmationButtons.CLOSE);
      } else {
        this.batchDelectionInfos = [];
        let initBatchMessage = new BatchInfo ();
        initBatchMessage.name = rule.name;
        this.batchDelectionInfos.push(initBatchMessage);
        deletionMessage = new ConfirmationMessage(
            'REPLICATION.DELETION_TITLE',
            'REPLICATION.DELETION_SUMMARY',
            rule.name || '',
            rule,
            ConfirmationTargets.POLICY,
            ConfirmationButtons.DELETE_CANCEL);
      }
      this.deletionConfirmDialog.open(deletionMessage);
    });
  }

}