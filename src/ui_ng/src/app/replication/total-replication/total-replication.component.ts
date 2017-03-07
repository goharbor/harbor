import { Component, OnInit, ViewChild } from '@angular/core';
import { ReplicationService } from '../../replication/replication.service';

import { CreateEditPolicyComponent } from '../../shared/create-edit-policy/create-edit-policy.component';

import { MessageService } from '../../global-message/message.service';
import { AlertType } from '../../shared/shared.const';

import { Policy } from '../../replication/policy';

@Component({
  selector: 'total-replication',
  templateUrl: 'total-replication.component.html',
  providers: [ ReplicationService ]
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
    private messageService: MessageService) {}

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
          error=>this.messageService.announceMessage(error.status,'Failed to get policies.', AlertType.DANGER)
        );
  }

  doSearchPolicies(policyName: string) {
    this.policyName = policyName;
    this.retrievePolicies();
  }
  
  openEditPolicy(policyId: number) {
    console.log('Open modal to edit policy ID:' + policyId);
    this.createEditPolicyComponent.openCreateEditPolicy(policyId);
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
      this.retrievePolicies();
    }
  }
}