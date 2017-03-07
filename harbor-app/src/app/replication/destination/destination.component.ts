import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';
import { Target } from '../target';
import { ReplicationService } from '../replication.service';
import { MessageService } from '../../global-message/message.service';
import { AlertType } from '../../shared/shared.const';

import { DeletionDialogService } from '../../shared/deletion-dialog/deletion-dialog.service';
import { DeletionMessage } from '../../shared/deletion-dialog/deletion-message';

import { DeletionTargets } from '../../shared/shared.const';

import { Subscription } from 'rxjs/Subscription';

import { CreateEditDestinationComponent } from '../create-edit-destination/create-edit-destination.component';

@Component({
  selector: 'destination',
  templateUrl: 'destination.component.html'
})
export class DestinationComponent implements OnInit {

  @ViewChild(CreateEditDestinationComponent) 
  createEditDestinationComponent: CreateEditDestinationComponent;  

  targets: Target[];
  target: Target;

  targetName: string;
  subscription : Subscription;

  constructor(
    private replicationService: ReplicationService,
    private messageService: MessageService,
    private deletionDialogService: DeletionDialogService) {
      this.subscription = this.deletionDialogService.deletionConfirm$.subscribe(message=>{
        let targetId = message.data;
        this.replicationService
            .deleteTarget(targetId)
            .subscribe(
              response=>{
                console.log('Successful deleted target with ID:' + targetId);
                this.reload();
              },
              error=>this.messageService
                         .announceMessage(error.status, 
                           'Failed to delete target with ID:' + targetId + ', error:' + error, 
                           AlertType.DANGER)
              );
      });
    }

  ngOnInit(): void {
    this.targetName = '';
    this.retrieve('');
  }

  ngOnDestroy(): void {
    if(this.subscription) {
      this.subscription.unsubscribe();
    }
  }

  retrieve(targetName: string): void {
    this.replicationService
        .listTargets(targetName)
        .subscribe(
          targets=>this.targets = targets,
          error=>this.messageService.announceMessage(error.status,'Failed to get targets:' + error, AlertType.DANGER)
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
    this.retrieve(this.targetName);
  }

  openModal() {
    this.createEditDestinationComponent.openCreateEditTarget();
    this.target = new Target();
  }

  editTarget(target: Target) {
    if(target) {
      this.createEditDestinationComponent.openCreateEditTarget(target.id);
    }
  }

  deleteTarget(target: Target) {
    if(target) {
      let targetId = target.id;
      let deletionMessage = new DeletionMessage('REPLICATION.DELETION_TITLE_TARGET', 'REPLICATION.DELETION_SUMMARY_TARGET', target.name, target.id, DeletionTargets.TARGET);
      this.deletionDialogService.openComfirmDialog(deletionMessage);
    }
  }
}