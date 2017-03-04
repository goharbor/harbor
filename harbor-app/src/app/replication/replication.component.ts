import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { CreateEditPolicyComponent } from '../shared/create-edit-policy/create-edit-policy.component';

import { MessageService } from '../global-message/message.service';
import { AlertType } from '../shared/shared.const';

import { SessionService } from '../shared/session.service';

import { ReplicationService } from './replication.service';

import { SessionUser } from '../shared/session-user';
import { Policy } from './policy';
import { Job } from './job';
import { Target } from './target';

const ruleStatus = [
  { 'key':  '', 'description': 'All Status'},
  { 'key': '1', 'description': 'Enabled'},
  { 'key': '0', 'description': 'Disabled'}
];

const jobStatus = [
  { 'key': '', 'description': 'All' },
  { 'key': 'pending',  'description': 'Pending' },
  { 'key': 'running',  'description': 'Running' },
  { 'key': 'error',    'description': 'Error' },
  { 'key': 'retrying', 'description': 'Retrying' },
  { 'key': 'stopped' , 'description': 'Stopped' },
  { 'key': 'finished', 'description': 'Finished' },
  { 'key': 'canceled', 'description': 'Canceled' }  
];

const optionalSearch: {} = {0: 'Advanced', 1: 'Simple'};

class SearchOption {
  policyId: number;
  policyName: string = '';
  repoName: string = '';
  status: string = '';
  startTime: string = '';
  endTime: string = '';
}

@Component({
  selector: 'replicaton',
  templateUrl: 'replication.component.html'
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

   policies: Policy[];
   jobs: Job[];

   toggleJobSearchOption = optionalSearch;
   currentJobSearchOption: number;

   @ViewChild(CreateEditPolicyComponent) 
   createEditPolicyComponent: CreateEditPolicyComponent;

   constructor(
     private sessionService: SessionService, 
     private messageService: MessageService,
     private replicationService: ReplicationService,
     private route: ActivatedRoute) {
     this.currentUser = this.sessionService.getCurrentUser();
   }

   ngOnInit(): void {
     this.projectId = +this.route.snapshot.parent.params['id'];
     console.log('Get projectId from route params snapshot:' + this.projectId);
     this.search = new SearchOption();
     this.currentRuleStatus = this.ruleStatus[0];
     this.currentJobStatus  = this.jobStatus[0];
     this.currentJobSearchOption = 0;
     this.retrievePolicies();
   }

   retrievePolicies(): void {
     this.replicationService
         .listPolicies(this.search.policyName, this.projectId)
         .subscribe(
           response=>{
             this.changedPolicies = response;
             this.policies = this.changedPolicies;
             if(this.changedPolicies && this.changedPolicies.length > 0) {
               this.fetchPolicyJobs(this.changedPolicies[0].id);
             } else {
               this.changedJobs = [];
             }
           },
           error=>this.messageService.announceMessage(error.status,'Failed to get policies with project ID:' + this.projectId, AlertType.DANGER)
         );
   }

   openModal(): void {
     console.log('Open modal to create policy.');
     this.createEditPolicyComponent.openCreateEditPolicy();
   }

   openEditPolicy(policyId: number) {
     console.log('Open modal to edit policy ID:' + policyId);
     this.createEditPolicyComponent.openCreateEditPolicy(policyId);
   }

   fetchPolicyJobs(policyId: number) { 
     this.search.policyId = policyId;
     console.log('Received policy ID ' + this.search.policyId + ' by clicked row.');
     this.replicationService
         .listJobs(this.search.policyId, this.search.status, this.search.repoName, this.search.startTime, this.search.endTime)
         .subscribe(
           response=>{
             this.changedJobs = response;
             this.jobs = this.changedJobs;
           },
           error=>this.messageService.announceMessage(error.status, 'Failed to fetch jobs with policy ID:' + this.search.policyId, AlertType.DANGER)
         );
   }

   selectOne(policy: Policy) {
     if(policy) {
      this.fetchPolicyJobs(policy.id);
     }
   }

   doSearchPolicies(policyName: string) {
     this.search.policyName = policyName;
     this.retrievePolicies();
   }

   doFilterPolicyStatus(status: string) {
     console.log('Do filter policies with status:' + status);
     this.currentRuleStatus = this.ruleStatus.find(r=>r.key === status);
     if(status.trim() === '') {
       this.changedPolicies = this.policies;
     } else {
       this.changedPolicies = this.policies.filter(policy=>policy.enabled === +this.currentRuleStatus.key);
     }
   }

   doFilterJobStatus(status: string) {
     console.log('Do filter jobs with status:' + status);
     this.currentJobStatus = this.jobStatus.find(r=>r.key === status);
     if(status.trim() === '') {
       this.changedJobs = this.jobs;
     } else {
       this.changedJobs = this.jobs.filter(job=>job.status === status);
     }
   }

   doSearchJobs(repoName: string) {
     this.search.repoName = repoName;
     this.fetchPolicyJobs(this.search.policyId);
   }

   reloadPolicies(isReady: boolean) {
     if(isReady) {
       this.retrievePolicies();
     }
   }

   refreshPolicies() {
     this.retrievePolicies();
   }

   refreshJobs() {
     this.fetchPolicyJobs(this.search.policyId);
   }

   toggleSearchJobOptionalName(option: number) {
     (option === 1) ? this.currentJobSearchOption = 0 : this.currentJobSearchOption = 1;
   }

   doJobSearchByTimeRange(strDate: string, target: string) {
     if(!strDate || strDate.trim() === '') {
        strDate = 0 + '';
     }
     let oneDayOffset = 3600 * 24;
     switch(target) {
     case 'begin':
       this.search.startTime = (new Date(strDate).getTime() / 1000) + '';
       break;
     case 'end':
       this.search.endTime = (new Date(strDate).getTime() / 1000 + oneDayOffset) + '';
       break;
     }
     console.log('Search jobs filtered by time range, begin: ' + this.search.startTime + ', end:' + this.search.endTime);
     this.fetchPolicyJobs(this.search.policyId);
   }

}