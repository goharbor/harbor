import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { CreateEditPolicyComponent } from './create-edit-policy/create-edit-policy.component';

import { MessageService } from '../global-message/message.service';
import { AlertType } from '../shared/shared.const';

import { ReplicationService } from './replication.service';

import { SessionUser } from '../shared/session-user';
import { Policy } from './policy';
import { Job } from './job';
import { Target } from './target';

@Component({
  selector: 'replicaton',
  templateUrl: 'replication.component.html'
})
export class ReplicationComponent implements OnInit {
   
   currentUser: SessionUser;
   projectId: number;

   policyName: string;
   
   policy: Policy;
  
   changedPolicies: Policy[];
   changedJobs: Job[];

   @ViewChild(CreateEditPolicyComponent) 
   createEditPolicyComponent: CreateEditPolicyComponent

   constructor(private route: ActivatedRoute, private messageService: MessageService, private replicationService: ReplicationService) {
     this.route.data.subscribe(data=>this.currentUser = <SessionUser>data);
   }

   ngOnInit(): void {
     this.projectId = +this.route.snapshot.parent.params['id'];
     console.log('Get projectId from route params snapshot:' + this.projectId);
     this.retrievePolicies();
   }

   retrievePolicies(): void {
     this.replicationService
         .listPolicies(this.projectId, this.policyName)
         .subscribe(
           response=>{
             this.changedPolicies = response;
             if(this.changedPolicies && this.changedPolicies.length > 0) {
               this.fetchPolicyJobs(this.changedPolicies[0].id);
             }
           },
           error=>this.messageService.announceMessage(error.status,'Failed to get policies with project ID:' + this.projectId, AlertType.DANGER)
         );
   }

   openModal(): void {
     this.createEditPolicyComponent.openCreateEditPolicy();
     console.log('Clicked open create-edit policy.');
   }

   fetchPolicyJobs(policyId: number) {
     console.log('Received policy ID ' + policyId + ' by clicked row.');
     this.replicationService
         .listJobs(policyId)
         .subscribe(
           response=>this.changedJobs = response,
           error=>this.messageService.announceMessage(error.status, 'Failed to fetch jobs with policy ID:' + policyId, AlertType.DANGER)
         );
   }
}