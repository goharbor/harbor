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
import { Component, Input, Output, EventEmitter, OnInit, ViewChild, AfterViewChecked } from '@angular/core';

import { NgForm } from '@angular/forms';

import { CreateEditPolicy } from './create-edit-policy';

import { ReplicationService } from '../../replication/replication.service';
import { MessageHandlerService } from '../message-handler/message-handler.service';
import { ActionType } from '../../shared/shared.const';

import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';

import { Policy } from '../../replication/policy';
import { Target } from '../../replication/target';

import { TranslateService } from '@ngx-translate/core';

const FAKE_PASSWORD: string = 'ywJZnDTM';

@Component({
  selector: 'create-edit-policy',
  templateUrl: 'create-edit-policy.component.html',
  styleUrls: [ 'create-edit-policy.component.css' ]
})
export class CreateEditPolicyComponent implements OnInit, AfterViewChecked {

  modalTitle: string;
  createEditPolicyOpened: boolean;
  createEditPolicy: CreateEditPolicy = new CreateEditPolicy();
  initVal: CreateEditPolicy = new CreateEditPolicy();
  
  actionType: ActionType;
  
  isCreateDestination: boolean;
  @Input() projectId: number;

  @Output() reload = new EventEmitter();

  targets: Target[];
  
  pingTestMessage: string;
  testOngoing: boolean;
  pingStatus: boolean;

  policyForm: NgForm;

  staticBackdrop: boolean = true;
  closable: boolean = false;

  @ViewChild('policyForm')
  currentForm: NgForm;

  hasChanged: boolean;

  editable: boolean;

  @ViewChild(InlineAlertComponent)
  inlineAlert: InlineAlertComponent;

  get readonly(): boolean {
    return this.actionType === ActionType.EDIT && this.createEditPolicy.enable;
  }

  get untoggleable(): boolean {
    return this.actionType === ActionType.EDIT && this.initVal.enable;
  }

  get showNewDestination(): boolean {
    return this.actionType === ActionType.ADD_NEW || !this.createEditPolicy.enable;
  }

  constructor(
    private replicationService: ReplicationService,
    private messageHandlerService: MessageHandlerService,
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
              this.createEditPolicy.password = FAKE_PASSWORD;

              this.initVal.targetId = this.createEditPolicy.targetId;
              this.initVal.endpointUrl = this.createEditPolicy.endpointUrl;
              this.initVal.username = this.createEditPolicy.username;
              this.initVal.password = this.createEditPolicy.password;
            }
          },
          error=>{ 
            this.messageHandlerService.handleError(error);
            this.createEditPolicyOpened = false;
          }
        );
  }

  ngOnInit(): void {}

  openCreateEditPolicy(editable: boolean, policyId?: number): void {
    this.createEditPolicyOpened = true;
    this.createEditPolicy = new CreateEditPolicy();
    
    this.editable = editable;

    this.isCreateDestination = false;
    
    this.hasChanged = false;

    this.pingTestMessage = '';
    this.pingStatus = true;
    this.testOngoing = false;  

    if(policyId) {
      this.actionType = ActionType.EDIT;
      this.translateService.get('REPLICATION.EDIT_POLICY_TITLE').subscribe(res=>this.modalTitle=res);
      this.replicationService
          .getPolicy(policyId)
          .subscribe(
            policy=>{
              this.createEditPolicy.policyId = policyId;
              this.createEditPolicy.name = policy.name;
              this.createEditPolicy.description = policy.description;
              this.createEditPolicy.enable = policy.enabled === 1? true : false;
              this.prepareTargets(policy.target_id);         

              this.initVal.name = this.createEditPolicy.name;
              this.initVal.description = this.createEditPolicy.description;
              this.initVal.enable = this.createEditPolicy.enable;
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
    if(this.isCreateDestination) {
      this.createEditPolicy.targetName = '';
      this.createEditPolicy.endpointUrl = '';
      this.createEditPolicy.username = '';
      this.createEditPolicy.password = '';
    } else {
      this.prepareTargets();
    }
  }

  selectTarget(): void {
    let result = this.targets.find(target=>target.id == this.createEditPolicy.targetId);
    if(result) {
      this.createEditPolicy.targetId = result.id;
      this.createEditPolicy.endpointUrl = result.endpoint;
      this.createEditPolicy.username = result.username;
      this.createEditPolicy.password = FAKE_PASSWORD;
    }
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
            this.messageHandlerService.showSuccess('REPLICATION.CREATED_SUCCESS');
            console.log('Successful created policy: ' + response);
            this.createEditPolicyOpened = false;
            this.reload.emit(true);
          },
          error=>{
            if(this.messageHandlerService.isAppLevel(error)) {
              this.messageHandlerService.handleError(error);
              this.createEditPolicyOpened = false;
            } else if (error.status === 409) {
              this.inlineAlert.showInlineError('REPLICATION.POLICY_ALREADY_EXISTS');
            } else {
              this.inlineAlert.showInlineError(error);
            }            
            console.log('Failed to create policy:' + error.status + ', error message:' + JSON.stringify(error['_body']));
          });
  }

  createOrUpdatePolicyAndCreateTarget(): void {
    console.log('Creating policy with new created target.');
    this.replicationService
        .createOrUpdatePolicyWithNewTarget(this.getPolicyByForm(), this.getTargetByForm())
        .subscribe(
          response=>{
            this.messageHandlerService.showSuccess('REPLICATION.CREATED_SUCCESS');
            console.log('Successful created policy and target:' + response);
            this.createEditPolicyOpened = false;
            this.reload.emit(true);
          },
          error=>{
            if(this.messageHandlerService.isAppLevel(error)) {
              this.messageHandlerService.handleError(error);
              this.createEditPolicyOpened = false;  
            } else if (error.status === 409) {
              this.inlineAlert.showInlineError('REPLICATION.POLICY_ALREADY_EXISTS');            
            } else {
              this.inlineAlert.showInlineError(error);
            }
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
            this.messageHandlerService.showSuccess('REPLICATION.UPDATED_SUCCESS')
            this.createEditPolicyOpened = false;
            this.reload.emit(true);
          },
          error=>{
            if(this.messageHandlerService.isAppLevel(error)) {
              this.messageHandlerService.handleError(error);
              this.createEditPolicyOpened = false;
            } else if (error.status === 409) {
              this.inlineAlert.showInlineError('REPLICATION.POLICY_ALREADY_EXISTS');
            } else {
              this.inlineAlert.showInlineError(error);
            }
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
  }

  onCancel() {
    if(this.hasChanged) {
      this.inlineAlert.showInlineConfirmation({message: 'ALERT.FORM_CHANGE_CONFIRMATION'});
    } else {
      this.createEditPolicyOpened = false;
      this.policyForm.reset();
    }
  }

  confirmCancel(confirmed: boolean) {
    this.createEditPolicyOpened = false;
    this.inlineAlert.close();
    this.policyForm.reset();
  }

  ngAfterViewChecked(): void {
    this.policyForm = this.currentForm;
    if(this.policyForm) {
      this.policyForm.valueChanges.subscribe(data=>{
        for(let i in data) {
          let origin = this.initVal[i];          
          let current = data[i];
          if(((this.actionType === ActionType.EDIT && !this.readonly && !current ) || current) && current !== origin) {
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

  testConnection() {
    this.pingStatus = true;
    this.translateService.get('REPLICATION.TESTING_CONNECTION').subscribe(res=>this.pingTestMessage=res);
    this.testOngoing = !this.testOngoing;
    let pingTarget: Target | any = {};
    if(this.isCreateDestination) {
      pingTarget.endpoint = this.createEditPolicy.endpointUrl;
      pingTarget.username = this.createEditPolicy.username;
      pingTarget.password = this.createEditPolicy.password;
    } else {
      pingTarget.id = this.createEditPolicy.targetId;
    }
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