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
import { Component, OnInit, ViewChild, OnDestroy, ChangeDetectionStrategy, ChangeDetectorRef } from '@angular/core';
import { ActivatedRoute, Params, Router } from '@angular/router';
import { Response } from '@angular/http';

import { SessionUser } from '../../shared/session-user';
import { Member } from './member';
import { MemberService } from './member.service';

import { AddMemberComponent } from './add-member/add-member.component';

import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { ConfirmationTargets, ConfirmationState, ConfirmationButtons } from '../../shared/shared.const';

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
import {TranslateService} from "@ngx-translate/core";
import {BatchInfo, BathInfoChanges} from "../../shared/confirmation-dialog/confirmation-batch-message";

@Component({
  templateUrl: 'member.component.html',
  styleUrls: ['./member.component.css'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class MemberComponent implements OnInit, OnDestroy {

  members: Member[];
  projectId: number;
  roleInfo = RoleInfo;
  delSub: Subscription;

  @ViewChild(AddMemberComponent)
  addMemberComponent: AddMemberComponent;

  currentUser: SessionUser;
  hasProjectAdminRole: boolean;

  searchMember: string;
  selectedRow: Member[] = [];
  batchDelectionInfos: BatchInfo[] = [];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private memberService: MemberService, 
    private translate: TranslateService,
    private messageHandlerService: MessageHandlerService,
    private deletionDialogService: ConfirmationDialogService,
    private session: SessionService,
    private ref: ChangeDetectorRef) {
    
    this.delSub = deletionDialogService.confirmationConfirm$.subscribe(message => {
      if (message &&
        message.state === ConfirmationState.CONFIRMED &&
        message.source === ConfirmationTargets.PROJECT_MEMBER) {
        this.deleteMem(message.data);
      }
    });
    let hnd = setInterval(()=>ref.markForCheck(), 100);
    setTimeout(()=>clearInterval(hnd), 1000);
  }

  retrieve(projectId: number, username: string) {
    this.memberService
      .listMembers(projectId, username)
      .subscribe(
      response => {
        this.members = response;
        let hnd = setInterval(()=>this.ref.markForCheck(), 100);
        setTimeout(()=>clearInterval(hnd), 1000);
      },
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
    //Get current user from registered resolver.
    this.currentUser = this.session.getCurrentUser();
    let resolverData = this.route.snapshot.parent.data;
    if(resolverData) {
      this.hasProjectAdminRole = (<Project>resolverData['projectResolver']).has_project_admin_role;
    }
    this.retrieve(this.projectId, '');
  }

  openAddMemberModal() {
    this.addMemberComponent.openAddMemberModal();
  }

  addedMember($event: any) {
    this.searchMember = '';
    this.retrieve(this.projectId, '');
  }

  changeRole(m: Member[], roleId: number) {
    if (m) {
      let promiseList: any[] = [];
      m.forEach(data => {
        if (!(data.user_id === this.currentUser.user_id  || !this.hasProjectAdminRole)) {
          promiseList.push(this.memberService.changeMemberRole(this.projectId, data.user_id, roleId));
        }
      })
      Promise.all(promiseList).then(num => {
            if (num.length === promiseList.length) {
              this.messageHandlerService.showSuccess('MEMBER.SWITCHED_SUCCESS');
              this.retrieve(this.projectId, '');
            }
          },
          error => {
            this.messageHandlerService.handleError(error);
          }
      );
      }
  }

  deleteMembers(m: Member[]) {
    let nameArr: string[] = [];
    this.batchDelectionInfos = [];
    if (m && m.length) {
      m.forEach(data => {
        nameArr.push(data.username);
        let initBatchMessage = new BatchInfo ();
        initBatchMessage.name = data.username;
        this.batchDelectionInfos.push(initBatchMessage);
      });
      this.deletionDialogService.addBatchInfoList(this.batchDelectionInfos);

      let deletionMessage = new ConfirmationMessage(
          'PROJECT.DELETION_TITLE',
          'PROJECT.DELETION_SUMMARY',
          nameArr.join(','),
          m,
          ConfirmationTargets.PROJECT_MEMBER,
          ConfirmationButtons.DELETE_CANCEL
      );
      this.deletionDialogService.openComfirmDialog(deletionMessage);
    }
  }

  deleteMem(members: Member[]) {
    if (members && members.length) {
      let promiseLists: any[] = [];
      members.forEach(member => {
        if (member.user_id === this.currentUser.user_id) {
          let findedList = this.batchDelectionInfos.find(data => data.name === member.username);
          this.translate.get('BATCH.DELETED_FAILURE').subscribe(res => {
            findedList = BathInfoChanges(findedList, res, false, true);
          });
        }else {
          promiseLists.push(this.delOperate(this.projectId, member.user_id, member.username));
        }

      });

      Promise.all(promiseLists).then(item => {
        this.selectedRow = [];
        this.retrieve(this.projectId, '');
      });
    }
  }

  delOperate(projectId: number, memberId: number, username: string) {
    let findedList = this.batchDelectionInfos.find(data => data.name === username);
    return this.memberService
        .deleteMember(projectId, memberId)
        .then(
            response => {
              this.translate.get('BATCH.DELETED_SUCCESS').subscribe(res => {
                findedList = BathInfoChanges(findedList, res);
              });
            },
            error => {
              this.translate.get('BATCH.DELETED_FAILURE').subscribe(res => {
                findedList = BathInfoChanges(findedList, res, false, true);
              });
            }
        );
  }

  SelectedChange(): void {
    //this.forceRefreshView(5000);
  }

  doSearch(searchMember: string) {
    this.searchMember = searchMember;
    this.retrieve(this.projectId, this.searchMember);
  }

  refresh() {
    this.retrieve(this.projectId, '');
  }
}