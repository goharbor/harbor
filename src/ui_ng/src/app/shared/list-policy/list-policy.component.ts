import { Component, Input, Output, EventEmitter, HostBinding, OnInit, ViewChild, OnDestroy } from '@angular/core';

import { ReplicationService } from '../../replication/replication.service';
import { Policy } from '../../replication/policy';

import { DeletionDialogService } from '../../shared/deletion-dialog/deletion-dialog.service';
import { DeletionMessage } from '../../shared/deletion-dialog/deletion-message';

import { DeletionTargets } from '../../shared/shared.const';

import { MessageService } from '../../global-message/message.service';
import { AlertType } from '../../shared/shared.const';

import { Subscription } from 'rxjs/Subscription';

@Component({
  selector: 'list-policy',
  templateUrl: 'list-policy.component.html',
})
export class ListPolicyComponent implements OnInit, OnDestroy {
  
  @Input() policies: Policy[];
  @Input() projectless: boolean;

  @Output() reload = new EventEmitter<boolean>();
  @Output() selectOne = new EventEmitter<Policy>();
  @Output() editOne = new EventEmitter<number>();
 
  selectedId: number;
  subscription: Subscription;

  constructor(
    private replicationService: ReplicationService,
    private deletionDialogService: DeletionDialogService,
    private messageService: MessageService) {
    
    this.subscription = this.subscription = this.deletionDialogService
         .deletionConfirm$
         .subscribe(
           message=>{
             if(message && message.targetId === DeletionTargets.POLICY) {
               this.replicationService
                   .deletePolicy(message.data)
                   .subscribe(
                     response=>{
                       console.log('Successful delete policy with ID:' + message.data);
                       this.reload.emit(true);
                     },
                     error=>this.messageService.announceMessage(error.status, 'Failed to delete policy with ID:' + message.data, AlertType.DANGER)
                   );
             }
           });

  }

  ngOnInit() {
    
  }

  ngOnDestroy() {
    if(this.subscription) {
      this.subscription.unsubscribe();
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
  
  enablePolicy(policy: Policy): void {
    console.log('Enable policy ID:' + policy.id + ' with activation status ' + policy.enabled);
    policy.enabled = policy.enabled === 0 ? 1 : 0; 
    this.replicationService.enablePolicy(policy.id, policy.enabled);
  }

  deletePolicy(policy: Policy) {
    let deletionMessage: DeletionMessage = new DeletionMessage('REPLICATION.DELETION_TITLE', 'REPLICATION.DELETION_SUMMARY', policy.name, policy.id, DeletionTargets.POLICY);
    this.deletionDialogService.openComfirmDialog(deletionMessage);
  }

}