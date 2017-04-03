import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';
import { Target } from '../target';
import { ReplicationService } from '../replication.service';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';

import { ConfirmationDialogService } from '../../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';

import { ConfirmationTargets, ConfirmationState } from '../../shared/shared.const';

import { Subscription } from 'rxjs/Subscription';

import { CreateEditDestinationComponent } from '../create-edit-destination/create-edit-destination.component';

@Component({
  selector: 'destination',
  templateUrl: 'destination.component.html',
  styleUrls: ['./destination.component.css']
})
export class DestinationComponent implements OnInit {

  @ViewChild(CreateEditDestinationComponent)
  createEditDestinationComponent: CreateEditDestinationComponent;

  targets: Target[];
  target: Target;

  targetName: string;
  subscription: Subscription;

  constructor(
    private replicationService: ReplicationService,
    private messageHandlerService: MessageHandlerService,
    private deletionDialogService: ConfirmationDialogService) {
    this.subscription = this.deletionDialogService.confirmationConfirm$.subscribe(message => {
      if (message &&
        message.source === ConfirmationTargets.TARGET &&
        message.state === ConfirmationState.CONFIRMED) {
        let targetId = message.data;
        this.replicationService
          .deleteTarget(targetId)
          .subscribe(
          response => {
            this.messageHandlerService.showSuccess('DESTINATION.DELETED_SUCCESS');
            this.reload();
          },
          error => { 
            if(error && error.status === 412) {
              this.messageHandlerService.showError('DESTINATION.FAILED_TO_DELETE_TARGET_IN_USED', '');
            } else {
              this.messageHandlerService.handleError(error);
            }
          });
      }
    });
  }

  ngOnInit(): void {
    this.targetName = '';
    this.retrieve('');
  }

  ngOnDestroy(): void {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
  }

  retrieve(targetName: string): void {
    this.replicationService
      .listTargets(targetName)
      .subscribe(
      targets => this.targets = targets,
      error => this.messageHandlerService.handleError(error)
      );
  }

  doSearchTargets(targetName: string) {
    this.targetName = targetName;
    this.retrieve(targetName);
  }

  refreshTargets() {
    this.retrieve('');
  }

  reload() {
    this.targetName = '';
    this.retrieve('');
  }

  openModal() {
    this.createEditDestinationComponent.openCreateEditTarget(true);
    this.target = new Target();
  }

  editTarget(target: Target) {
    if (target) {
      let editable = true;
      this.replicationService
          .listTargetPolicies(target.id)
          .subscribe(
            policies=>{
              if(policies && policies.length > 0) {
                for(let i = 0; i < policies.length; i++){
                  let p = policies[i];
                  if(p.enabled === 1) {
                    editable = false;
                    break;
                  }
                }
              }
              this.createEditDestinationComponent.openCreateEditTarget(editable, target.id);
            },
            error=>this.messageHandlerService.handleError(error)
          );
      
    }
  }

  deleteTarget(target: Target) {
    if (target) {
      let targetId = target.id;
      let deletionMessage = new ConfirmationMessage(
        'REPLICATION.DELETION_TITLE_TARGET',
        'REPLICATION.DELETION_SUMMARY_TARGET',
        target.name,
        target.id,
        ConfirmationTargets.TARGET);
      this.deletionDialogService.openComfirmDialog(deletionMessage);
    }
  }
}