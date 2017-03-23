import { Component, Output, EventEmitter, ViewChild, AfterViewChecked } from '@angular/core';
import { NgForm } from '@angular/forms';

import { ReplicationService } from '../replication.service';
import { MessageService } from '../../global-message/message.service';
import { AlertType, ActionType } from '../../shared/shared.const';

import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';

import { Target } from '../target';

import { TranslateService } from '@ngx-translate/core';

@Component({
  selector: 'create-edit-destination',
  templateUrl: './create-edit-destination.component.html'
})
export class CreateEditDestinationComponent implements AfterViewChecked {

  modalTitle: string;
  createEditDestinationOpened: boolean;

  testOngoing: boolean;
  pingTestMessage: string;
  pingStatus: boolean;

  actionType: ActionType;

  target: Target = new Target();
  initVal: Target = new Target();

  targetForm: NgForm;

  staticBackdrop: boolean = true;
  closable: boolean = false;

  @ViewChild('targetForm')
  currentForm: NgForm;

  hasChanged: boolean;

  @ViewChild(InlineAlertComponent)
  inlineAlert: InlineAlertComponent;

  @Output() reload = new EventEmitter<boolean>();
  
  constructor(
    private replicationService: ReplicationService,
    private messageService: MessageService,
    private translateService: TranslateService) {}

  openCreateEditTarget(targetId?: number) {
    this.target = new Target();
    this.createEditDestinationOpened = true;
    
    this.hasChanged = false;
    
    this.pingTestMessage = '';
    this.pingStatus = true;
    this.testOngoing = false;  

    if(targetId) {
      this.actionType = ActionType.EDIT;
      this.translateService.get('DESTINATION.TITLE_EDIT').subscribe(res=>this.modalTitle=res);
      this.replicationService
          .getTarget(targetId)
          .subscribe(
            target=>{ 
              this.target = target;
              this.initVal.name = this.target.name;
              this.initVal.endpoint = this.target.endpoint;
              this.initVal.username = this.target.username;
              this.initVal.password = this.target.password;
            },
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
                    this.messageService.announceMessage(error.status, errorMessageKey, AlertType.DANGER);
                    this.inlineAlert.showInlineError(errorMessageKey);
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
                    this.inlineAlert.showInlineError(errorMessageKey);
                    this.messageService.announceMessage(error.status, errorMessageKey, AlertType.DANGER);
                  });
            }
          );
        break;
    }
  }

  onCancel() {
    if(this.hasChanged) {
      this.inlineAlert.showInlineConfirmation({message: 'ALERT.FORM_CHANGE_CONFIRMATION'});
    } else {
      this.createEditDestinationOpened = false;
    }
  }

  confirmCancel(confirmed: boolean) {
    this.createEditDestinationOpened = false;
    this.inlineAlert.close();
  }

  mappedName: {} = {
    'targetName': 'name',
    'endpointUrl': 'endpoint',
    'username': 'username',
    'password': 'password'
  };

  ngAfterViewChecked(): void {
    this.targetForm = this.currentForm;
    if(this.targetForm) {
      this.targetForm.valueChanges.subscribe(data=>{
        for(let i in data) {
          let current = data[i];
          let origin = this.initVal[this.mappedName[i]];
          if(current && current !== origin) {
            this.hasChanged = true;
            break;
          } else {
            this.hasChanged = false;
            this.inlineAlert.close();
          }
        }
      });
    }
  }

}