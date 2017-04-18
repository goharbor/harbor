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
import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';
import { ActivatedRoute, Params, Router } from '@angular/router';
import { Response } from '@angular/http';

import { SessionUser } from '../../shared/session-user';
import { Member } from './member';
import { MemberService } from './member.service';

import { AddMemberComponent } from './add-member/add-member.component';

import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { ConfirmationTargets, ConfirmationState } from '../../shared/shared.const';

import { ConfirmationDialogService } from '../../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';
import { SessionService } from '../../shared/session.service';

import { RoleInfo } from '../../shared/shared.const';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/switchMap';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';
import { Subscription } from 'rxjs/Subscription';

import { Project } from '../../project/project';

@Component({
  templateUrl: 'member.component.html',
  styleUrls: ['./member.component.css']
})
export class MemberComponent implements OnInit, OnDestroy {

  members: Member[];
  projectId: number;
  roleInfo = RoleInfo;
  private delSub: Subscription;

  @ViewChild(AddMemberComponent)
  addMemberComponent: AddMemberComponent;

  currentUser: SessionUser;
  hasProjectAdminRole: boolean;

  searchMember: string;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private memberService: MemberService, 
    private messageHandlerService: MessageHandlerService,
    private deletionDialogService: ConfirmationDialogService,
    private session: SessionService) {
    
    this.delSub = deletionDialogService.confirmationConfirm$.subscribe(message => {
      if (message &&
        message.state === ConfirmationState.CONFIRMED &&
        message.source === ConfirmationTargets.PROJECT_MEMBER) {
        this.memberService
          .deleteMember(this.projectId, message.data)
          .subscribe(
          response => {
            this.messageHandlerService.showSuccess('MEMBER.DELETED_SUCCESS');
            console.log('Successful delete member: ' + message.data);
            this.retrieve(this.projectId, '');
          },
          error => this.messageHandlerService.handleError(error)
          );
      }
    });
  }

  retrieve(projectId: number, username: string) {
    this.memberService
      .listMembers(projectId, username)
      .subscribe(
      response => this.members = response,
      error => {
        this.router.navigate(['/harbor', 'projects']);
        this.messageHandlerService.handleError(error);
      });
  }

  ngOnDestroy() {
    if (this.delSub) {
      this.delSub.unsubscribe();
    }
  }

  ngOnInit() {
    //Get projectId from route params snapshot.          
    this.projectId = +this.route.snapshot.parent.params['id'];
    console.log('Get projectId from route params snapshot:' + this.projectId);
    
    this.currentUser = this.session.getCurrentUser();
    //Get current user from registered resolver.
    let resolverData = this.route.snapshot.parent.data;
    if(resolverData) {
      this.hasProjectAdminRole = (<Project>resolverData['projectResolver']).has_project_admin_role;
    }

   

    this.retrieve(this.projectId, '');
  }

  openAddMemberModal() {
    this.addMemberComponent.openAddMemberModal();
  }

  addedMember() {
    this.searchMember = '';
    this.retrieve(this.projectId, '');
  }

  changeRole(m: Member, roleId: number) {
    if(m) {
      this.memberService
        .changeMemberRole(this.projectId, m.user_id, roleId)
        .subscribe(
        response => {
          this.messageHandlerService.showSuccess('MEMBER.SWITCHED_SUCCESS');
          console.log('Successful change role with user ' + m.user_id + ' to roleId ' + roleId);
          this.retrieve(this.projectId, '');
        },
        error => this.messageHandlerService.handleError(error)
        );
      }
  }

  deleteMember(m: Member) {
    let deletionMessage: ConfirmationMessage = new ConfirmationMessage(
      'MEMBER.DELETION_TITLE',
      'MEMBER.DELETION_SUMMARY',
      m.username,
      m.user_id,
      ConfirmationTargets.PROJECT_MEMBER
    );
    this.deletionDialogService.openComfirmDialog(deletionMessage);
  }

  doSearch(searchMember) {
    this.searchMember = searchMember;
    this.retrieve(this.projectId, this.searchMember);
  }

  refresh() {
    this.retrieve(this.projectId, '');
  }
}