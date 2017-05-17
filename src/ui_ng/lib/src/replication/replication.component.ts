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
import { Component, OnInit, ViewChild, Input } from '@angular/core';
// import { ActivatedRoute } from '@angular/router';
import { ResponseOptions } from '@angular/http';
import { NgModel } from '@angular/forms';

import { CreateEditRuleComponent } from '../create-edit-rule/create-edit-rule.component';

import { ErrorHandler } from '../error-handler/error-handler';

import { ReplicationService } from '../service/replication.service';

import { RequestQueryParams } from '../service/RequestQueryParams';
// import { SessionUser } from '../shared/session-user';
import { ReplicationRule, ReplicationJob, Endpoint } from '../service/interface';

import { State } from 'clarity-angular';

import { toPromise } from '../utils';

import { TranslateService } from '@ngx-translate/core';

import { REPLICATION_STYLE } from './replication.component.css';
import { REPLICATION_TEMPLATE } from './replication.component.html';

const ruleStatus: {[key: string]: any} = [
  { 'key': 'all', 'description': 'REPLICATION.ALL_STATUS'},
  { 'key': '1', 'description': 'REPLICATION.ENABLED'},
  { 'key': '0', 'description': 'REPLICATION.DISABLED'}
];

const jobStatus: {[key: string]: any} = [
  { 'key': 'all', 'description': 'REPLICATION.ALL' },
  { 'key': 'pending',  'description': 'REPLICATION.PENDING' },
  { 'key': 'running',  'description': 'REPLICATION.RUNNING' },
  { 'key': 'error',    'description': 'REPLICATION.ERROR' },
  { 'key': 'retrying', 'description': 'REPLICATION.RETRYING' },
  { 'key': 'stopped' , 'description': 'REPLICATION.STOPPED' },
  { 'key': 'finished', 'description': 'REPLICATION.FINISHED' },
  { 'key': 'canceled', 'description': 'REPLICATION.CANCELED' }  
];

const optionalSearch: {} = {0: 'REPLICATION.ADVANCED', 1: 'REPLICATION.SIMPLE'};

export class SearchOption {
  ruleId: number | string;
  ruleName: string = '';
  repoName: string = '';
  status: string = '';
  startTime: string = '';
  startTimestamp: string = '';
  endTime: string = '';
  endTimestamp: string = '';
  page: number = 1;
  pageSize: number = 5;
}

@Component({
  selector: 'hbr-replication',
  template: REPLICATION_TEMPLATE
})
export class ReplicationComponent implements OnInit {
   
   @Input() projectId: number | string;

   search: SearchOption = new SearchOption();

   ruleStatus = ruleStatus;
   currentRuleStatus: {key: string, description: string};

   jobStatus = jobStatus;
   currentJobStatus: {key: string, description: string};

   changedRules: ReplicationRule[];
   changedJobs: ReplicationJob[];
   initSelectedId: number | string;

   rules: ReplicationRule[];
   jobs: ReplicationJob[];

   jobsTotalRecordCount: number;
   jobsTotalPage: number;

   toggleJobSearchOption = optionalSearch;
   currentJobSearchOption: number;

   @ViewChild(CreateEditRuleComponent) 
   createEditPolicyComponent: CreateEditRuleComponent;

   constructor(
     private errorHandler: ErrorHandler,
     private replicationService: ReplicationService,
     private translateService: TranslateService) {
   }

   ngOnInit(): void {
     this.projectId = 1;
     this.currentRuleStatus = this.ruleStatus[0];
     this.currentJobStatus  = this.jobStatus[0];
     this.currentJobSearchOption = 0;
     this.retrievePolicies();
   }

   retrievePolicies(): void {
     toPromise<ReplicationRule[]>(this.replicationService
         .getReplicationRules(this.projectId, this.search.ruleName))
         .then(response=>{
             this.changedRules = response || [];
             if(this.changedRules && this.changedRules.length > 0) {
               this.initSelectedId = this.changedRules[0].id || '';
             }
             this.rules = this.changedRules;
             if(this.changedRules && this.changedRules.length > 0) {
               this.search.ruleId = this.changedRules[0].id || '';
               this.fetchReplicationJobs();
             }
           },
           error=>this.errorHandler.error(error)
         );
   } 

   openModal(): void {
     this.createEditPolicyComponent.openCreateEditRule(true);
   }

   openEditRule(rule: ReplicationRule) {
     if(rule) {
       let editable = true;
       if(rule.enabled === 1) {
         editable = false;
       }
       this.createEditPolicyComponent.openCreateEditRule(editable, rule.id);
     }
   }

   fetchReplicationJobs(state?: State) { 
     if(state && state.page && state.page.to) {
       this.search.page = state.page.to + 1;
     }
     let params: RequestQueryParams = new RequestQueryParams();
     params.set('status', this.search.status);
     params.set('repository', this.search.repoName);
     params.set('start_time', this.search.startTimestamp);
     params.set('end_time', this.search.endTimestamp);
     params.set('page', this.search.page + '');
     params.set('page_size', this.search.pageSize + '');

     toPromise<any>(this.replicationService
       .getJobs(this.search.ruleId, params))
       .then(
         response=>{
           this.jobsTotalRecordCount = response.headers.get('x-total-count');
           this.jobsTotalPage = Math.ceil(this.jobsTotalRecordCount / this.search.pageSize);
           this.changedJobs = response.json();
           this.jobs = this.changedJobs;
           this.jobs.forEach(j=>{
             if(j.status === 'retrying' || j.status === 'error') {
               this.translateService.get('REPLICATION.FOUND_ERROR_IN_JOBS')
                  .subscribe(res=>this.errorHandler.error(res));
             }
           })    
         }).catch(error=>this.errorHandler.error(error));
   }

   selectOneRule(rule: ReplicationRule) {
     if (rule) {
       this.search.ruleId = rule.id || '';
       this.search.repoName = '';
       this.search.status = '';
       this.currentJobSearchOption = 0;
       this.currentJobStatus = { 'key': 'all', 'description': 'REPLICATION.ALL' };
       this.fetchReplicationJobs();
     }
   }
   
   doSearchRules(ruleName: string) {
     this.search.ruleName = ruleName;
     this.retrievePolicies();
   }

   doFilterRuleStatus($event: any) {
     if ($event && $event.target && $event.target["value"]) {
       let status = $event.target["value"];
       this.currentRuleStatus = this.ruleStatus.find((r: any)=>r.key === status);
       if(this.currentRuleStatus.key === 'all') {
         this.changedRules = this.rules;
       } else {
         this.changedRules = this.rules.filter(policy=>policy.enabled === +this.currentRuleStatus.key);
       }
     }
   }

   doFilterJobStatus($event: any) {
     if ($event && $event.target && $event.target["value"]) {
       let status = $event.target["value"];
       this.currentJobStatus = this.jobStatus.find((r: any)=>r.key === status);
       if(this.currentJobStatus.key === 'all') {
         status = '';
       }
       this.search.status = status;
       this.doSearchJobs(this.search.repoName);
     }
   }

   doSearchJobs(repoName: string) {
     this.search.repoName = repoName;
     this.fetchReplicationJobs();
   }

   reloadRules(isReady: boolean) {
     if(isReady) {
       this.search.ruleName = '';
       this.retrievePolicies();
     }
   }

   refreshRules() {
     this.retrievePolicies();
   }

   refreshJobs() {
     this.fetchReplicationJobs();
   }

   toggleSearchJobOptionalName(option: number) {
     (option === 1) ? this.currentJobSearchOption = 0 : this.currentJobSearchOption = 1;
   }

   doJobSearchByStartTime(fromTimestamp: string) {
     this.search.startTimestamp = fromTimestamp;
     this.fetchReplicationJobs();
   }

   doJobSearchByEndTime(toTimestamp: string) {
     this.search.endTimestamp = toTimestamp;
     this.fetchReplicationJobs();
   }
}