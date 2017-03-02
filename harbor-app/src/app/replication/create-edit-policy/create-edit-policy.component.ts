import { Component, Input, Output, EventEmitter, OnInit } from '@angular/core';

import { CreateEditPolicy } from './create-edit-policy';

import { ReplicationService } from '../replication.service';
import { MessageService } from '../../global-message/message.service';
import { AlertType } from '../../shared/shared.const';

import { Policy } from '../policy';
import { Target } from '../target';

@Component({
  selector: 'create-edit-policy',
  templateUrl: 'create-edit-policy.component.html'
})
export class CreateEditPolicyComponent implements OnInit {

  createEditPolicyOpened: boolean;
  createEditPolicy: CreateEditPolicy = new CreateEditPolicy();
  
  errorMessageOpened: boolean;
  errorMessage: string;
  
  isCreateDestination: boolean;
  @Input() projectId: number;

  @Output() reload = new EventEmitter();

  targets: Target[];
  
  constructor(private replicationService: ReplicationService,
              private messageService: MessageService) {}
  
  prepareTargets(targetId?: number) {
    this.replicationService
        .listTargets()
        .subscribe(
          targets=>{
            this.targets = targets; 
            if(this.targets && this.targets.length > 0) {
              let initialTarget: Target;
              (targetId) ? initialTarget = this.targets.find(t=>t.id===targetId) : initialTarget = this.targets[0]; 
              this.createEditPolicy.targetId = initialTarget.id;
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
    console.log('createEditPolicyOpened:' + this.createEditPolicyOpened);
    this.createEditPolicyOpened = true;
    this.createEditPolicy = new CreateEditPolicy();
    this.errorMessageOpened = false;
    this.errorMessage = '';
    this.prepareTargets();
    if(policyId) {
      this.replicationService
          .getPolicy(policyId)
          .subscribe(
            policy=>{
              this.createEditPolicy.name = policy.name;
              this.createEditPolicy.description = policy.description;
              this.createEditPolicy.enable = policy.enabled === 1? true : false;
              this.createEditPolicy.targetId = policy.target_id;
              this.prepareTargets(policy.target_id);
            }
          )
    }    
  } 

  newDestination(checkedAddNew: boolean): void {
    console.log('CheckedAddNew:' + checkedAddNew);
    this.isCreateDestination = checkedAddNew;
  }

  selectTarget(): void {
    let results = this.targets.filter(target=>target.id == this.createEditPolicy.targetId);
    if(results && results.length > 0) {
      this.createEditPolicy.targetId = results[0].id;
      this.createEditPolicy.endpointUrl = results[0].endpoint;
      this.createEditPolicy.username = results[0].username;
      this.createEditPolicy.password = results[0].password;
    }
  }
  
  onErrorMessageClose(): void {
    this.errorMessageOpened = false;
    this.errorMessage = '';
  }
  
  getPolicyByForm(): Policy {
    let policy = new Policy();
    policy.project_id = this.projectId;
    policy.name = this.createEditPolicy.name;
    policy.description = this.createEditPolicy.description;
    policy.enabled = this.createEditPolicy.enable ? 1 : 0;
    policy.target_id = this.createEditPolicy.targetId;
    return policy;
  }

  getTargetByForm(): Target {
    let target = new Target();
    target.name = this.createEditPolicy.targetName;
    target.endpoint = this.createEditPolicy.endpointUrl;
    target.username = this.createEditPolicy.username;
    target.password = this.createEditPolicy.password;
    return target;
  }

  createPolicy(): void {
    console.log('Create policy with existed target in component.');
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

  createPolicyWithTarget(): void {
    console.log('Create policy with new target in component.');
    this.replicationService
        .createPolcyWithTarget(this.getPolicyByForm(), this.getTargetByForm())
        .subscribe(
          response=>{
            console.log('Successful created policy with added target:' + response);
            this.createEditPolicyOpened = false;
            this.reload.emit(true);
          },
          error=>{
            this.errorMessageOpened = true;
            this.errorMessage = error['_body'];
            console.log('Failed to create policy with new added target:' + error.status + ', error message:' + JSON.stringify(error['_body']));
          }
        );
  }

  onSubmit() {
    if(this.isCreateDestination) {
      this.createPolicyWithTarget();
    } else {
      this.createPolicy();
    }
    this.errorMessageOpened = false;
    this.errorMessage = '';
  }
}