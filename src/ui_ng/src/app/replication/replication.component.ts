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
import { ActivatedRoute } from '@angular/router';

import { CreateEditPolicyComponent } from '../shared/create-edit-policy/create-edit-policy.component';

import { MessageHandlerService } from '../shared/message-handler/message-handler.service';

import { SessionService } from '../shared/session.service';

import { ReplicationService } from './replication.service';

import { SessionUser } from '../shared/session-user';
import { Policy } from './policy';
import { Job } from './job';
import { Target } from './target';

import { State } from 'clarity-angular';

const ruleStatus = [
  { 'key':  '', 'description': 'REPLICATION.ALL_STATUS'},
  { 'key': '1', 'description': 'REPLICATION.ENABLED'},
  { 'key': '0', 'description': 'REPLICATION.DISABLED'}
];

const jobStatus = [
  { 'key': '', 'description': 'REPLICATION.ALL' },
  { 'key': 'pending',  'description': 'REPLICATION.PENDING' },
  { 'key': 'running',  'description': 'REPLICATION.RUNNING' },
  { 'key': 'error',    'description': 'REPLICATION.ERROR' },
  { 'key': 'retrying', 'description': 'REPLICATION.RETRYING' },
  { 'key': 'stopped' , 'description': 'REPLICATION.STOPPED' },
  { 'key': 'finished', 'description': 'REPLICATION.FINISHED' },
  { 'key': 'canceled', 'description': 'REPLICATION.CANCELED' }  
];

const optionalSearch: {} = {0: 'REPLICATION.ADVANCED', 1: 'REPLICATION.SIMPLE'};

class SearchOption {
  policyId: number;
  policyName: string = '';
  repoName: string = '';
  status: string = '';
  startTime: string = '';
  endTime: string = '';
  page: number = 1;
  pageSize: number = 5;
}

@Component({
  selector: 'replicaton',
  templateUrl: 'replication.component.html',
  styleUrls: ['./replication.component.css']
})
export class ReplicationComponent implements OnInit {
   
   currentUser: SessionUser;
   projectId: number;

   search: SearchOption;

   ruleStatus = ruleStatus;
   currentRuleStatus: {key: string, description: string};

   jobStatus = jobStatus;
   currentJobStatus: {key: string, description: string};

   changedPolicies: Policy[];
   changedJobs: Job[];
   initSelectedId: number;

   policies: Policy[];
   jobs: Job[];

   jobsTotalRecordCount: number;
   jobsTotalPage: number;

   toggleJobSearchOption = optionalSearch;
   currentJobSearchOption: number;

   @ViewChild(CreateEditPolicyComponent) 
   createEditPolicyComponent: CreateEditPolicyComponent;

   constructor(
     private sessionService: SessionService, 
     private messageHandlerService: MessageHandlerService,
     private replicationService: ReplicationService,
     private route: ActivatedRoute) {
     this.currentUser = this.sessionService.getCurrentUser();
   }

   ngOnInit(): void {
     this.projectId = +this.route.snapshot.parent.params['id'];
     this.search = new SearchOption();
     this.currentRuleStatus = this.ruleStatus[0];
     this.currentJobStatus  = this.jobStatus[0];
     this.currentJobSearchOption = 0;
     this.retrievePolicies();

     let isCreate = this.route.snapshot.parent.queryParams['is_create'];
     if (isCreate && <boolean>isCreate) {
       this.openModal();
     }
   }

   retrievePolicies(): void {
     this.replicationService
         .listPolicies(this.search.policyName, this.projectId)
         .subscribe(
           response=>{
             this.changedPolicies = response;
             if(this.changedPolicies && this.changedPolicies.length > 0) {
               this.initSelectedId = this.changedPolicies[0].id;
             }
             this.policies = this.changedPolicies;
             if(this.changedPolicies && this.changedPolicies.length > 0) {
               this.search.policyId = this.changedPolicies[0].id;
               this.fetchPolicyJobs();
             } else {
               this.changedJobs = [];
             }
           },
           error=>this.messageHandlerService.handleError(error)
         );
   }

   openModal(): void {
     this.createEditPolicyComponent.openCreateEditPolicy(true);
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

   fetchPolicyJobs(state?: State) { 
     if(state) {
       this.search.page = state.page.to + 1;
     }
     this.replicationService
         .listJobs(this.search.policyId, this.search.status, this.search.repoName, 
           this.search.startTime, this.search.endTime, this.search.page, this.search.pageSize)
         .subscribe(
           response=>{
             this.jobsTotalRecordCount = response.headers.get('x-total-count');
             this.jobsTotalPage = Math.ceil(this.jobsTotalRecordCount / this.search.pageSize);
             this.changedJobs = response.json();
             this.jobs = this.changedJobs;
             for(let i = 0; i < this.jobs.length; i++) {
               let j = this.jobs[i];
               if(j.status == 'retrying' || j.status == 'error') {
                 this.messageHandlerService.showError('REPLICATION.FOUND_ERROR_IN_JOBS', '');
                 break;
               }
             }
           },
           error=>this.messageHandlerService.handleError(error)
         );
   }

   selectOnePolicy(policy: Policy) {
     if(policy) {
      this.search.policyId = policy.id;
      this.search.repoName = '';
      this.search.status = ''
      this.currentJobSearchOption = 0;
      this.currentJobStatus = { 'key': '', 'description': 'REPLICATION.ALL'};
      this.fetchPolicyJobs();
     }
   }
   
   doSearchPolicies(policyName: string) {
     this.search.policyName = policyName;
     this.retrievePolicies();
   }

   doFilterPolicyStatus(status: string) {
     this.currentRuleStatus = this.ruleStatus.find(r=>r.key === status);
     if(status.trim() === '') {
       this.changedPolicies = this.policies;
     } else {
       this.changedPolicies = this.policies.filter(policy=>policy.enabled === +this.currentRuleStatus.key);
     }
   }

   doFilterJobStatus(status: string) {
     this.currentJobStatus = this.jobStatus.find(r=>r.key === status);
     this.search.status = status;
     this.doSearchJobs(this.search.repoName);
   }

   doSearchJobs(repoName: string) {
     this.search.repoName = repoName;
     this.fetchPolicyJobs();
   }

   reloadPolicies(isReady: boolean) {
     if(isReady) {
       this.search.policyName = '';
       this.retrievePolicies();
     }
   }

   refreshPolicies() {
     this.retrievePolicies();
   }

   refreshJobs() {
     this.fetchPolicyJobs();
   }

   toggleSearchJobOptionalName(option: number) {
     (option === 1) ? this.currentJobSearchOption = 0 : this.currentJobSearchOption = 1;
   }

   doJobSearchByStartTime(strDate: string) {
     if(!strDate || strDate.trim() === '') {
        strDate = 0 + '';
     }     
     (strDate === '0') ? this.search.startTime = '' : this.search.startTime = (new Date(strDate).getTime() / 1000) + '';
     this.fetchPolicyJobs();
   }

   doJobSearchByEndTime(strDate: string) {
     if(!strDate || strDate.trim() === '') {
        strDate = 0 + '';
     }
     let oneDayOffset = 3600 * 24;
     (strDate === '0') ? this.search.endTime = '' : this.search.endTime = (new Date(strDate).getTime() / 1000 + oneDayOffset) + '';
     this.fetchPolicyJobs();
   }
}