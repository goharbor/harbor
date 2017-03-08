import { Component, Output, EventEmitter } from '@angular/core';

import { ReplicationService } from '../replication.service';
import { MessageService } from '../../global-message/message.service';
import { AlertType, ActionType } from '../../shared/shared.const';

import { Target } from '../target';


@Component({
  selector: 'create-edit-destination',
  templateUrl: './create-edit-destination.component.html'
})
export class CreateEditDestinationComponent {

  createEditDestinationOpened: boolean;

  errorMessageOpened: boolean;
  errorMessage: string;

  testOngoing: boolean;
  pingTestMessage: string;
  pingStatus: boolean;

  actionType: ActionType;

  target: Target = new Target();

  @Output() reload = new EventEmitter<boolean>();
  
  constructor(
    private replicationService: ReplicationService,
    private messageService: MessageService) {}

  openCreateEditTarget(targetId?: number) {
    this.target = new Target();

    this.createEditDestinationOpened = true;
    
    this.errorMessageOpened = false;
    this.errorMessage = '';
    
    this.pingTestMessage = '';
    this.pingStatus = true;
    this.testOngoing = false;  

    if(targetId) {
      this.actionType = ActionType.EDIT;
      this.replicationService
          .getTarget(targetId)
          .subscribe(
            target=>this.target=target,
            error=>this.messageService
                       .announceMessage(error.status, 'Failed to get target with ID:' + targetId, AlertType.DANGER)
          );
    } else {
      this.actionType = ActionType.ADD_NEW;
    }
  }

  testConnection() {
    this.pingTestMessage = 'Testing connection...';
    this.pingStatus = true;
    this.testOngoing = !this.testOngoing;
    this.replicationService
        .pingTarget(this.target)
        .subscribe(
          response=>{
            this.pingStatus = true;
            this.pingTestMessage = 'Connection tested successfully.';
            this.testOngoing = !this.testOngoing;
          },
          error=>{
            this.pingStatus = false;
            this.pingTestMessage = 'Failed to ping target.';
            this.testOngoing = !this.testOngoing;
          }
        )
  }

  onSubmit() {
    this.errorMessage = '';
    this.errorMessageOpened = false;

    switch(this.actionType) {
    case ActionType.ADD_NEW:
      this.replicationService
          .createTarget(this.target)
          .subscribe(
            response=>{
              console.log('Successful added target.');
              this.createEditDestinationOpened = false;
              this.reload.emit(true);
            },
            error=>{
              this.errorMessageOpened = true;
              this.errorMessage = 'Failed to add target:' + error;
              this.messageService
                      .announceMessage(error.status, this.errorMessage, AlertType.DANGER);
            }
          );
        break;
    case ActionType.EDIT:
      this.replicationService
          .updateTarget(this.target)
          .subscribe(
            response=>{ 
              console.log('Successful updated target.');
              this.createEditDestinationOpened = false;
              this.reload.emit(true);
            },
            error=>{
              this.errorMessageOpened = true;
              this.errorMessage = 'Failed to update target:' + error;
              this.messageService
                      .announceMessage(error.status, this.errorMessage, AlertType.DANGER);
            }
          );
        break;
    }
  }

  onErrorMessageClose(): void {
    this.errorMessageOpened = false;
    this.errorMessage = '';
  }

}