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
import { Component, Input, Output, EventEmitter, ViewChild, OnDestroy } from '@angular/core';

import { ReplicationService } from '../../replication/replication.service';
import { Policy } from '../../replication/policy';

import { ConfirmationDialogService } from '../../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';

import { ConfirmationState, ConfirmationTargets } from '../../shared/shared.const';

import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';

import { Subscription } from 'rxjs/Subscription';

@Component({
  selector: 'list-policy',
  templateUrl: 'list-policy.component.html',
})
export class ListPolicyComponent implements OnDestroy {

  nullTime: string = '0001-01-01T00:00:00Z';

  @Input() policies: Policy[];
  @Input() projectless: boolean;
  @Input() selectedId: number;

  @Output() reload = new EventEmitter<boolean>();
  @Output() selectOne = new EventEmitter<Policy>();
  @Output() editOne = new EventEmitter<Policy>();
  @Output() toggleOne = new EventEmitter<Policy>();

  toggleSubscription: Subscription;
  subscription: Subscription;

  constructor(
    private replicationService: ReplicationService,
    private toggleConfirmDialogService: ConfirmationDialogService,
    private deletionDialogService: ConfirmationDialogService,
    private messageHandlerService: MessageHandlerService) {

    this.toggleSubscription = this.toggleConfirmDialogService
        .confirmationConfirm$
        .subscribe(
          message=> {
            if(message &&
             message.source === ConfirmationTargets.TOGGLE_CONFIRM && 
             message.state === ConfirmationState.CONFIRMED) {
               let policy: Policy = message.data;
               policy.enabled = policy.enabled === 0 ? 1 : 0;
               console.log('Enable policy ID:' + policy.id + ' with activation status ' + policy.enabled);
               this.replicationService
                   .enablePolicy(policy.id, policy.enabled)
                   .subscribe(
                      response => {
                        this.messageHandlerService.showSuccess('REPLICATION.TOGGLED_SUCCESS');
                        console.log('Successful toggled policy status')
                      },
                      error => this.messageHandlerService.handleError(error)
                   );
             }
          }
        );
    this.subscription =  this.deletionDialogService
      .confirmationConfirm$
      .subscribe(
        message => {
          if (message &&
            message.source === ConfirmationTargets.POLICY &&
            message.state === ConfirmationState.CONFIRMED) {
            this.replicationService
                .deletePolicy(message.data)
                .subscribe(
                  response => {
                    this.messageHandlerService.showSuccess('REPLICATION.DELETED_SUCCESS');
                    console.log('Successful delete policy with ID:' + message.data);
                    this.reload.emit(true);
                  },
                  error => {
                    if(error && error.status === 412) {
                      this.messageHandlerService.handleError('REPLICATION.FAILED_TO_DELETE_POLICY_ENABLED');
                    } else {
                      this.messageHandlerService.handleError(error);
                    }
                  }
                );
        }
      }
    );

  }

  ngOnDestroy() {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
    if(this.toggleSubscription) {
      this.toggleSubscription.unsubscribe();
    }
  }

  selectPolicy(policy: Policy): void {
    this.selectedId = policy.id;
    console.log('Select policy ID:' + policy.id);
    this.selectOne.emit(policy);
  }

  editPolicy(policy: Policy) {
    console.log('Open modal to edit policy.');
    this.editOne.emit(policy);
  }

  togglePolicy(policy: Policy) {
    let toggleConfirmMessage: ConfirmationMessage = new ConfirmationMessage(
      policy.enabled === 1 ? 'REPLICATION.TOGGLE_DISABLE_TITLE' : 'REPLICATION.TOGGLE_ENABLE_TITLE',
      policy.enabled === 1 ? 'REPLICATION.CONFIRM_TOGGLE_DISABLE_POLICY': 'REPLICATION.CONFIRM_TOGGLE_ENABLE_POLICY',
      policy.name,
      policy,
      ConfirmationTargets.TOGGLE_CONFIRM
    );
    this.toggleConfirmDialogService.openComfirmDialog(toggleConfirmMessage);
  }

  deletePolicy(policy: Policy) {
    let deletionMessage: ConfirmationMessage = new ConfirmationMessage(
      'REPLICATION.DELETION_TITLE',
      'REPLICATION.DELETION_SUMMARY',
      policy.name,
      policy.id,
      ConfirmationTargets.POLICY);
    this.deletionDialogService.openComfirmDialog(deletionMessage);
  }

}