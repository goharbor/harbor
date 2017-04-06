import { Component, OnInit, ViewChild } from '@angular/core';
import { ReplicationService } from '../../replication/replication.service';

import { CreateEditPolicyComponent } from '../../shared/create-edit-policy/create-edit-policy.component';

import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';

import { Policy } from '../../replication/policy';

@Component({
  selector: 'total-replication',
  templateUrl: 'total-replication.component.html',
  providers: [ ReplicationService ],
  styleUrls: ['./total-replication.component.css']
})
export class TotalReplicationComponent implements OnInit {

  changedPolicies: Policy[];
  policies: Policy[];
  policyName: string = '';
  projectId: number;

  @ViewChild(CreateEditPolicyComponent) 
  createEditPolicyComponent: CreateEditPolicyComponent;

  constructor(
    private replicationService: ReplicationService,
    private messageHandlerService: MessageHandlerService) {}

  ngOnInit() {
    this.retrievePolicies();
  }

  retrievePolicies(): void {
    this.replicationService
        .listPolicies(this.policyName)
        .subscribe(
          response=>{
            this.changedPolicies = response;
            this.policies = this.changedPolicies;
          },
          error=>this.messageHandlerService.handleError(error)
        );
  }

  doSearchPolicies(policyName: string) {
    this.policyName = policyName;
    this.retrievePolicies();
  }
  
  openEditPolicy(policy: Policy) {
    if(policy) {
      let editable = true;
      if(policy.enabled === 1) {
        editable = false;
      }
      this.createEditPolicyComponent.openCreateEditPolicy(editable, policy.id);
    }
  }

  selectPolicy(policy: Policy) {
    if(policy) {
      this.projectId = policy.project_id;
    }
  }

  refreshPolicies() {
    this.retrievePolicies();
  }

  reloadPolicies(isReady: boolean) {
    if(isReady) {
      this.policyName = '';
      this.retrievePolicies();
    }
  }
}