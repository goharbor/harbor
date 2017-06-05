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
import { Component, Input, Output, EventEmitter, ViewChild, ChangeDetectionStrategy, ChangeDetectorRef } from '@angular/core';

import { ReplicationService } from '../service/replication.service';
import { ReplicationRule } from '../service/interface';

import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ConfirmationMessage } from '../confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../confirmation-dialog/confirmation-state-message';

import { ConfirmationState, ConfirmationTargets, ConfirmationButtons } from '../shared/shared.const';

import { TranslateService } from '@ngx-translate/core';

import { ErrorHandler } from '../error-handler/error-handler';
import { toPromise, CustomComparator } from '../utils';

import { State, Comparator } from 'clarity-angular';

import { LIST_REPLICATION_RULE_TEMPLATE } from './list-replication-rule.component.html';

@Component({
  selector: 'list-replication-rule',
  template: LIST_REPLICATION_RULE_TEMPLATE,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListReplicationRuleComponent {

  nullTime: string = '0001-01-01T00:00:00Z';

  @Input() rules: ReplicationRule[];
  @Input() projectless: boolean;
  @Input() selectedId: number | string;

  @Input() loading: boolean = false;

  @Output() reload = new EventEmitter<boolean>();
  @Output() selectOne = new EventEmitter<ReplicationRule>();
  @Output() editOne = new EventEmitter<ReplicationRule>();
  @Output() toggleOne = new EventEmitter<ReplicationRule>();

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
    setInterval(()=>ref.markForCheck(), 500);
  }

  toggleConfirm(message: ConfirmationAcknowledgement) {
    if(message &&
      message.source === ConfirmationTargets.TOGGLE_CONFIRM && 
      message.state === ConfirmationState.CONFIRMED) {
      let rule: ReplicationRule = message.data;
      if(rule) {
        rule.enabled = rule.enabled === 0 ? 1 : 0;
        toPromise<any>(this.replicationService
          .enableReplicationRule(rule.id || '', rule.enabled))
          .then(() => 
            this.translateService.get('REPLICATION.TOGGLED_SUCCESS')
                .subscribe(res=>this.errorHandler.info(res)))
          .catch(error => this.errorHandler.error(error));
        }
    }
  }

  deletionConfirm(message: ConfirmationAcknowledgement) {
    if (message &&
      message.source === ConfirmationTargets.POLICY &&
      message.state === ConfirmationState.CONFIRMED) {
      toPromise<any>(this.replicationService
        .deleteReplicationRule(message.data))
        .then(() => {
          this.translateService.get('REPLICATION.DELETED_SUCCESS')
              .subscribe(res=>this.errorHandler.info(res));
          this.reload.emit(true);
        })
        .catch(error => {
          if(error && error.status === 412) {
            this.translateService.get('REPLICATION.FAILED_TO_DELETE_POLICY_ENABLED')
                .subscribe(res=>this.errorHandler.error(res));
          } else {
            this.errorHandler.error(error);
          }
        });
    }
  }

  selectRule(rule: ReplicationRule): void {
    this.selectedId = rule.id || '';
    this.selectOne.emit(rule);
  }

  editRule(rule: ReplicationRule) {
    this.editOne.emit(rule);
  }

  toggleRule(rule: ReplicationRule) {
    let toggleConfirmMessage: ConfirmationMessage = new ConfirmationMessage(
      rule.enabled === 1 ? 'REPLICATION.TOGGLE_DISABLE_TITLE' : 'REPLICATION.TOGGLE_ENABLE_TITLE',
      rule.enabled === 1 ? 'REPLICATION.CONFIRM_TOGGLE_DISABLE_POLICY': 'REPLICATION.CONFIRM_TOGGLE_ENABLE_POLICY',
      rule.name || '',
      rule,
      ConfirmationTargets.TOGGLE_CONFIRM
    );
    this.toggleConfirmDialog.open(toggleConfirmMessage);
  }

  deleteRule(rule: ReplicationRule) {
    let deletionMessage: ConfirmationMessage = new ConfirmationMessage(
      'REPLICATION.DELETION_TITLE',
      'REPLICATION.DELETION_SUMMARY',
      rule.name || '',
      rule.id,
      ConfirmationTargets.POLICY,
      ConfirmationButtons.DELETE_CANCEL);
    this.deletionConfirmDialog.open(deletionMessage);
  }

}