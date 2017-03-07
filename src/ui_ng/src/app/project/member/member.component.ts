import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute, Params, Router } from '@angular/router';
import { Response } from '@angular/http';

import { SessionUser } from '../../shared/session-user';
import { Member } from './member';
import { MemberService } from './member.service';

import { AddMemberComponent } from './add-member/add-member.component';

import { MessageService } from '../../global-message/message.service';
import { AlertType, DeletionTargets } from '../../shared/shared.const';

import { DeletionDialogService } from '../../shared/deletion-dialog/deletion-dialog.service';
import { DeletionMessage } from '../../shared/deletion-dialog/deletion-message';
import { SessionService } from '../../shared/session.service';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/switchMap';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';

export const roleInfo: {} = { 1: 'MEMBER.PROJECT_ADMIN', 2: 'MEMBER.DEVELOPER', 3: 'MEMBER.GUEST' };

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

  constructor(private route: ActivatedRoute, private router: Router,
    private memberService: MemberService, private messageService: MessageService,
    private deletionDialogService: DeletionDialogService,
    session:SessionService) {
    //Get current user from registered resolver.
    this.currentUser = session.getCurrentUser();
    deletionDialogService.deletionConfirm$.subscribe(message => {
      if (message && message.targetId === DeletionTargets.PROJECT_MEMBER) {
        this.memberService
          .deleteMember(this.projectId, message.data)
          .subscribe(
          response => {
            console.log('Successful change role with user ' + message.data);
            this.retrieve(this.projectId, '');
          },
          error => this.messageService.announceMessage(error.status, 'Failed to change role with user ' + message.data, AlertType.DANGER)
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
        this.messageService.announceMessage(error.status, 'Failed to get project member with project ID:' + projectId, AlertType.DANGER);
      }
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
      response => {
        console.log('Successful change role with user ' + userId + ' to roleId ' + roleId);
        this.retrieve(this.projectId, '');
      },
      error => this.messageService.announceMessage(error.status, 'Failed to change role with user ' + userId + ' to roleId ' + roleId, AlertType.DANGER)
      );
  }

  deleteMember(userId: number) {
    let deletionMessage: DeletionMessage = new DeletionMessage(
      'MEMBER.DELETION_TITLE',
      'MEMBER.DELETION_SUMMARY',
      userId+"",
      userId,
      DeletionTargets.PROJECT_MEMBER
    );
    this.deletionDialogService.openComfirmDialog(deletionMessage);
  }

  doSearch(searchMember) {
    this.retrieve(this.projectId, searchMember);
  }

  refresh() {
    this.retrieve(this.projectId, '');
  }
}