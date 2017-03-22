import { Component, Input, Output, EventEmitter, ViewChild, OnDestroy } from '@angular/core';

import { ReplicationService } from '../../replication/replication.service';
import { Policy } from '../../replication/policy';

import { ConfirmationDialogService } from '../../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';

import { ConfirmationState, ConfirmationTargets } from '../../shared/shared.const';

import { MessageService } from '../../global-message/message.service';
import { AlertType } from '../../shared/shared.const';

import { Subscription } from 'rxjs/Subscription';

@Component({
  selector: 'list-policy',
  templateUrl: 'list-policy.component.html',
})
export class ListPolicyComponent implements OnDestroy {

  @Input() policies: Policy[];
  @Input() projectless: boolean;
  @Input() selectedId: number;

  @Output() reload = new EventEmitter<boolean>();
  @Output() selectOne = new EventEmitter<Policy>();
  @Output() editOne = new EventEmitter<number>();
  @Output() toggleOne = new EventEmitter<Policy>();

  toggleSubscription: Subscription;
  subscription: Subscription;

  constructor(
    private replicationService: ReplicationService,
    private toggleConfirmDialogService: ConfirmationDialogService,
    private deletionDialogService: ConfirmationDialogService,
    private messageService: MessageService) {

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
                      res => console.log('Successful toggled policy status'),
                      error => this.messageService.announceMessage(error.status, "Failed to toggle policy status.", AlertType.DANGER)
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
                    console.log('Successful delete policy with ID:' + message.data);
                    this.reload.emit(true);
                  },
                  error => this.messageService.announceMessage(error.status, 'Failed to delete policy with ID:' + message.data, AlertType.DANGER)
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
    this.editOne.emit(policy.id);
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