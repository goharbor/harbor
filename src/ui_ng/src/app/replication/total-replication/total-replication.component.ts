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