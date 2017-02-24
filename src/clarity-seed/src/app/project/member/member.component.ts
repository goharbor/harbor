import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute, Params } from '@angular/router';

import { SessionUser } from '../../shared/session-user';
import { Member } from './member';
import { MemberService } from './member.service';

import { AddMemberComponent } from './add-member/add-member.component';

import { MessageService } from '../../global-message/message.service';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/switchMap';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';

export const roleInfo: {} = { 1: 'ProjectAdmin', 2: 'Developer', 3: 'Guest'};

@Component({
  templateUrl: 'member.component.html'
})
export class MemberComponent implements OnInit {
 
  currentUser: SessionUser;
  members: Member[];
  projectId: number;
  roleInfo = roleInfo;

  @ViewChild(AddMemberComponent)
  addMemberComponent: AddMemberComponent;

  constructor(private route: ActivatedRoute, private memberService: MemberService, private messageService: MessageService) {
    //Get current user from registered resolver.
    this.route.data.subscribe(data=>this.currentUser = <SessionUser>data['memberResolver']);    
  }

  retrieve(projectId:number, username: string) {
    this.memberService
        .listMembers(projectId, username)
        .subscribe(
          response=>this.members = response,
          error=>console.log(error)
        );
  }

  ngOnInit() {
    //Get projectId from route params snapshot.          
    this.projectId = +this.route.snapshot.parent.params['id'];
    console.log('Get projectId from route params snapshot:' + this.projectId);

    this.retrieve(this.projectId, '');
  }

  openAddMemberModal() {
    this.addMemberComponent.openAddMemberModal();
  }
  
  addedMember() {
    this.retrieve(this.projectId, '');
  }

  changeRole(userId: number, roleId: number) {
    this.memberService
        .changeMemberRole(this.projectId, userId, roleId)
        .subscribe(
          response=>{
            console.log('Successful change role with user ' + userId + ' to roleId ' + roleId);
            this.retrieve(this.projectId, '');
          },
          error => this.messageService.announceMessage('Failed to change role with user ' + userId + ' to roleId ' + roleId)
        );
  }

  deleteMember(userId: number) {
    this.memberService
        .deleteMember(this.projectId, userId)
        .subscribe(
          response=>{
            console.log('Successful change role with user ' + userId);
            this.retrieve(this.projectId, '');
          },
          error => this.messageService.announceMessage('Failed to change role with user ' + userId)
        );
  }
  
  doSearch(searchMember) {
    this.retrieve(this.projectId, searchMember);
  }
}