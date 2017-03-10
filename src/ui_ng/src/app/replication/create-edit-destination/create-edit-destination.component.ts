import { Component, Output, EventEmitter } from '@angular/core';

import { ReplicationService } from '../replication.service';
import { MessageService } from '../../global-message/message.service';
import { AlertType, ActionType } from '../../shared/shared.const';

import { Target } from '../target';

import { TranslateService } from '@ngx-translate/core';

@Component({
  selector: 'create-edit-destination',
  templateUrl: './create-edit-destination.component.html'
})
export class CreateEditDestinationComponent {

  modalTitle: string;
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
    private messageService: MessageService,
    private translateService: TranslateService) {}

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
      this.translateService.get('DESTINATION.TITLE_EDIT').subscribe(res=>this.modalTitle=res);
      this.replicationService
          .getTarget(targetId)
          .subscribe(
            target=>this.target=target,
            error=>this.messageService
                       .announceMessage(error.status, 'DESTINATION.FAILED_TO_GET_TARGET', AlertType.DANGER)
          );
    } else {
      this.actionType = ActionType.ADD_NEW;
      this.translateService.get('DESTINATION.TITLE_ADD').subscribe(res=>this.modalTitle=res);
    }
  }

  testConnection() {
    this.translateService.get('DESTINATION.TESTING_CONNECTION').subscribe(res=>this.pingTestMessage=res);
    this.pingStatus = true;
    this.testOngoing = !this.testOngoing;
    this.replicationService
        .pingTarget(this.target)
        .subscribe(
          response=>{
            this.pingStatus = true;
            this.translateService.get('DESTINATION.TEST_CONNECTION_SUCCESS').subscribe(res=>this.pingTestMessage=res);
            this.testOngoing = !this.testOngoing;
          },
          error=>{
            this.pingStatus = false;
            this.translateService.get('DESTINATION.TEST_CONNECTION_FAILURE').subscribe(res=>this.pingTestMessage=res);
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
              let errorMessageKey = '';
              switch(error.status) {
              case 409:
                errorMessageKey = 'DESTINATION.CONFLICT_NAME';
                break;
              case 400:
                errorMessageKey = 'DESTINATION.INVALID_NAME';
                break;
              default:
                errorMessageKey = 'UNKNOWN_ERROR';
              }
              this.translateService
                  .get(errorMessageKey)
                  .subscribe(res=>{
                    this.errorMessage = res;
                    this.messageService.announceMessage(error.status, errorMessageKey, AlertType.DANGER);
                  });
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
              let errorMessageKey = '';
              switch(error.status) {
              case 409:
                errorMessageKey = 'DESTINATION.CONFLICT_NAME';
                break;
              case 400:
                errorMessageKey = 'DESTINATION.INVALID_NAME';
                break;
              default:
                errorMessageKey = 'UNKNOWN_ERROR';
              }
              this.translateService
                  .get(errorMessageKey)
                  .subscribe(res=>{
                    this.errorMessage = res;
                    this.messageService.announceMessage(error.status, errorMessageKey, AlertType.DANGER);
                  });
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