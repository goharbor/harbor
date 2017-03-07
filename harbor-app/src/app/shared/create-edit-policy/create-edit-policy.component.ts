import { Component, Input, Output, EventEmitter, OnInit, HostBinding } from '@angular/core';

import { CreateEditPolicy } from './create-edit-policy';

import { ReplicationService } from '../../replication/replication.service';
import { MessageService } from '../../global-message/message.service';
import { AlertType, ActionType } from '../../shared/shared.const';

import { Policy } from '../../replication/policy';
import { Target } from '../../replication/target';

import { TranslateService } from '@ngx-translate/core';

@Component({
  selector: 'create-edit-policy',
  templateUrl: 'create-edit-policy.component.html'
})
export class CreateEditPolicyComponent implements OnInit {

  modalTitle: string;
  createEditPolicyOpened: boolean;
  createEditPolicy: CreateEditPolicy = new CreateEditPolicy();
  
  actionType: ActionType;

  errorMessageOpened: boolean;
  errorMessage: string;
  
  isCreateDestination: boolean;
  @Input() projectId: number;

  @Output() reload = new EventEmitter();

  targets: Target[];
  
  pingTestMessage: string;
  testOngoing: boolean;
  pingStatus: boolean;

  constructor(
    private replicationService: ReplicationService,
    private messageService: MessageService,
    private translateService: TranslateService) {}
  
  prepareTargets(targetId?: number) {
    this.replicationService
        .listTargets('')
        .subscribe(
          targets=>{
            this.targets = targets; 
            if(this.targets && this.targets.length > 0) {
              let initialTarget: Target;
              (targetId) ? initialTarget = this.targets.find(t=>t.id==targetId) : initialTarget = this.targets[0]; 
              this.createEditPolicy.targetId = initialTarget.id;
              this.createEditPolicy.targetName = initialTarget.name;
              this.createEditPolicy.endpointUrl = initialTarget.endpoint;
              this.createEditPolicy.username = initialTarget.username;
              this.createEditPolicy.password = initialTarget.password;
            }
          },
          error=>this.messageService.announceMessage(error.status, 'Error occurred while get targets.', AlertType.DANGER)
        );
  }

  ngOnInit(): void {}

  openCreateEditPolicy(policyId?: number): void {
    this.createEditPolicyOpened = true;
    this.createEditPolicy = new CreateEditPolicy();
    this.isCreateDestination = false;
    this.errorMessageOpened = false;
    this.errorMessage = '';
    
    this.pingTestMessage = '';
    this.pingStatus = true;
    this.testOngoing = false;  

    if(policyId) {
      this.actionType = ActionType.EDIT;
      this.translateService.get('REPLICATION.EDIT_POLICY').subscribe(res=>this.modalTitle=res);
      this.replicationService
          .getPolicy(policyId)
          .subscribe(
            policy=>{
              this.createEditPolicy.policyId = policyId;
              this.createEditPolicy.name = policy.name;
              this.createEditPolicy.description = policy.description;
              this.createEditPolicy.enable = policy.enabled === 1? true : false;
              this.prepareTargets(policy.target_id);
            }
          )
    } else {
      this.actionType = ActionType.ADD_NEW;
      this.translateService.get('REPLICATION.ADD_POLICY').subscribe(res=>this.modalTitle=res);
      this.prepareTargets(); 
    }
  } 

  newDestination(checkedAddNew: boolean): void {
    console.log('CheckedAddNew:' + checkedAddNew);
    this.isCreateDestination = checkedAddNew;
    this.createEditPolicy.targetName = '';
    this.createEditPolicy.endpointUrl = '';
    this.createEditPolicy.username = '';
    this.createEditPolicy.password = '';
  }

  selectTarget(): void {
    let result = this.targets.find(target=>target.id == this.createEditPolicy.targetId);
    if(result) {
      this.createEditPolicy.targetId = result.id;
      this.createEditPolicy.endpointUrl = result.endpoint;
      this.createEditPolicy.username = result.username;
      this.createEditPolicy.password = result.password;
    }
  }
  
  onErrorMessageClose(): void {
    this.errorMessageOpened = false;
    this.errorMessage = '';
  }
  
  getPolicyByForm(): Policy {
    let policy = new Policy();
    policy.project_id = this.projectId;
    policy.id = this.createEditPolicy.policyId;
    policy.name = this.createEditPolicy.name;
    policy.description = this.createEditPolicy.description;
    policy.enabled = this.createEditPolicy.enable ? 1 : 0;
    policy.target_id = this.createEditPolicy.targetId;
    return policy;
  }

  getTargetByForm(): Target {
    let target = new Target();
    target.id = this.createEditPolicy.targetId;
    target.name = this.createEditPolicy.targetName;
    target.endpoint = this.createEditPolicy.endpointUrl;
    target.username = this.createEditPolicy.username;
    target.password = this.createEditPolicy.password;
    return target;
  }

  createPolicy(): void {
    console.log('Create policy with existing target in component.');
    this.replicationService
        .createPolicy(this.getPolicyByForm())
        .subscribe(
          response=>{
            console.log('Successful created policy: ' + response);
            this.createEditPolicyOpened = false;
            this.reload.emit(true);
          },
          error=>{
            this.errorMessageOpened = true;
            this.errorMessage = error['_body'];
            console.log('Failed to create policy:' + error.status + ', error message:' + JSON.stringify(error['_body']));
          });
  }

  createOrUpdatePolicyAndCreateTarget(): void {
    console.log('Creating policy with new created target.');
    this.replicationService
        .createOrUpdatePolicyWithNewTarget(this.getPolicyByForm(), this.getTargetByForm())
        .subscribe(
          response=>{
            console.log('Successful created policy and target:' + response);
            this.createEditPolicyOpened = false;
            this.reload.emit(true);
          },
          error=>{
            this.errorMessageOpened = true;
            this.errorMessage = error['_body'];
            console.log('Failed to create policy and target:' + error.status + ', error message:' + JSON.stringify(error['_body']));
          }
        );
  }

  updatePolicy(): void {
    console.log('Creating policy with existing target.');
    this.replicationService
        .updatePolicy(this.getPolicyByForm())
        .subscribe(
          response=>{
            console.log('Successful created policy and target:' + response);
            this.createEditPolicyOpened = false;
            this.reload.emit(true);
          },
          error=>{
            this.errorMessageOpened = true;
            this.errorMessage = error['_body'];
            console.log('Failed to create policy and target:' + error.status + ', error message:' + JSON.stringify(error['_body']));
          }
        );
  }

  onSubmit() {
    if(this.isCreateDestination) {
      this.createOrUpdatePolicyAndCreateTarget();
    } else {
      if(this.actionType === ActionType.ADD_NEW) {
        this.createPolicy();
      } else if(this.actionType === ActionType.EDIT){
        this.updatePolicy();
      }
    }
    
    this.errorMessageOpened = false;
    this.errorMessage = '';
  }

  testConnection() {
    this.pingStatus = true;
    this.translateService.get('REPLICATION.TESTING_CONNECTION').subscribe(res=>this.pingTestMessage=res);
    this.testOngoing = !this.testOngoing;
    let pingTarget = new Target();
    pingTarget.endpoint = this.createEditPolicy.endpointUrl;
    pingTarget.username = this.createEditPolicy.username;
    pingTarget.password = this.createEditPolicy.password;
    this.replicationService
        .pingTarget(pingTarget)
        .subscribe(
          response=>{
            this.testOngoing = !this.testOngoing;
            this.translateService.get('REPLICATION.TEST_CONNECTION_SUCCESS').subscribe(res=>this.pingTestMessage=res);
            this.pingStatus = true;
          },
          error=>{
            this.testOngoing = !this.testOngoing;
            this.translateService.get('REPLICATION.TEST_CONNECTION_FAILURE').subscribe(res=>this.pingTestMessage=res);
            this.pingStatus = false;
          }
        );
  }
}