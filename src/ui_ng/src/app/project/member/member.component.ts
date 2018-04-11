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
import { Component, OnInit, ViewChild, OnDestroy, ChangeDetectionStrategy, ChangeDetectorRef } from "@angular/core";
import { ActivatedRoute, Router } from "@angular/router";

import { SessionUser } from "../../shared/session-user";
import { Member } from "./member";
import { MemberService } from "./member.service";

import { AddMemberComponent } from "./add-member/add-member.component";

import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { ConfirmationTargets, ConfirmationState, ConfirmationButtons } from "../../shared/shared.const";

import { ConfirmationDialogService } from "../../shared/confirmation-dialog/confirmation-dialog.service";
import { ConfirmationMessage } from "../../shared/confirmation-dialog/confirmation-message";
import { SessionService } from "../../shared/session.service";

import { RoleInfo } from "../../shared/shared.const";

import "rxjs/add/operator/switchMap";
import "rxjs/add/operator/catch";
import "rxjs/add/operator/map";
import "rxjs/add/observable/throw";
import { Subscription } from "rxjs/Subscription";

import { Project } from "../../project/project";
import {TranslateService} from "@ngx-translate/core";
import {BatchInfo, BathInfoChanges} from "../../shared/confirmation-dialog/confirmation-batch-message";

@Component({
  templateUrl: "member.component.html",
  styleUrls: ["./member.component.css"],
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
  roleNum: number;
  isDelete = false;
  isChangeRole = false;
  batchActionInfos: BatchInfo[] = [];
  batchDeletionInfos: BatchInfo[] = [];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private memberService: MemberService,
    private translate: TranslateService,
    private messageHandlerService: MessageHandlerService,
    private OperateDialogService: ConfirmationDialogService,
    private session: SessionService,
    private ref: ChangeDetectorRef) {

    this.delSub = OperateDialogService.confirmationConfirm$.subscribe(message => {
      if (message &&
        message.state === ConfirmationState.CONFIRMED &&
        message.source === ConfirmationTargets.PROJECT_MEMBER) {
        if (this.isDelete) {
          this.deleteMem(message.data);
        }
        if (this.isChangeRole) {
          this.changeOpe(message.data);
        }
      }
    });
    let hnd = setInterval(() => ref.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 1000);
  }

  retrieve(projectId: number, username: string) {
    this.selectedRow = [];
    this.memberService
      .listMembers(projectId, username)
      .subscribe(
      response => {
        this.members = response;
        let hnd = setInterval(() => this.ref.markForCheck(), 100);
        setTimeout(() => clearInterval(hnd), 1000);
      },
      error => {
        this.router.navigate(["/harbor", "projects"]);
        this.messageHandlerService.handleError(error);
      });
  }

  ngOnDestroy() {
    if (this.delSub) {
      this.delSub.unsubscribe();
    }
  }

  ngOnInit() {
    // Get projectId from route params snapshot.
    this.projectId = +this.route.snapshot.parent.params["id"];
    // Get current user from registered resolver.
    this.currentUser = this.session.getCurrentUser();
    let resolverData = this.route.snapshot.parent.data;
    if (resolverData) {
      this.hasProjectAdminRole = (<Project>resolverData["projectResolver"]).has_project_admin_role;
    }
    this.retrieve(this.projectId, "");
  }

  openAddMemberModal() {
    this.addMemberComponent.openAddMemberModal();
  }

  addedMember($event: any) {
    this.searchMember = "";
    this.retrieve(this.projectId, "");
  }

  get onlySelf(): boolean {
    if (this.selectedRow.length === 1 && this.selectedRow[0].id === this.currentUser.user_id) {
      return true;
    }
    return false;
  }

  changeRole(m: Member[], roleId: number) {
    if (m && m.length) {
      this.isDelete = false;
      this.isChangeRole = true;
      this.roleNum = roleId;
      let nameArr: string[] = [];
      this.batchActionInfos = [];
      m.forEach(data => {
        nameArr.push(data.entity_name);
        let initBatchMessage = new BatchInfo();
        initBatchMessage.name = data.entity_name;
        this.batchActionInfos.push(initBatchMessage);
      });

      this.changeOpe(m);
    }
  }

  changeOpe(members: Member[]) {
    if (members && members.length) {
      let promiseList: any[] = [];
      members.forEach(member => {
        if (member.id === this.currentUser.user_id) {
          let foundMember = this.batchActionInfos.find(batchInfo => batchInfo.name === member.entity_name);
          this.translate.get("BATCH.SWITCH_FAILURE").subscribe(res => {
            this.messageHandlerService.handleError(res + ": " + foundMember.name);
            foundMember = BathInfoChanges(foundMember, res, false, true);
          });
        } else {
          promiseList.push(this.changeOperate(this.projectId, member.id, this.roleNum, member.entity_name));
        }
      });

      Promise.all(promiseList).then(num => {
            this.retrieve(this.projectId, "");
          },
      );
    }
  }

  changeOperate(projectId: number, memberId: number, roleId: number, username: string) {
    let foundMember = this.batchActionInfos.find(batchInfo => batchInfo.name === username);
    return this.memberService
        .changeMemberRole(projectId, memberId, roleId)
        .then(
            response => {
              this.translate.get("BATCH.SWITCH_SUCCESS").subscribe(res => {
                foundMember = BathInfoChanges(foundMember, res);
              });
            },
            error => {
              this.translate.get("BATCH.SWITCH_FAILURE").subscribe(res => {
                this.messageHandlerService.handleError(res + ": " + username);
                foundMember = BathInfoChanges(foundMember, res, false, true);
              });
            }
        );
  }

  ChangeRoleOngoing(username: string) {
    if (this.batchActionInfos) {
      let memberActionInfo = this.batchActionInfos.find(batchInfo => batchInfo.name === username);
      return memberActionInfo && memberActionInfo.status === "pending";
    } else {
      return false;
    }
  }

  deleteMembers(m: Member[]) {
    this.isDelete = true;
    this.isChangeRole = false;
    let nameArr: string[] = [];
    this.batchDeletionInfos = [];
    if (m && m.length) {
      m.forEach(data => {
        nameArr.push(data.entity_name);
        let initBatchMessage = new BatchInfo ();
        initBatchMessage.name = data.entity_name;
        this.batchDeletionInfos.push(initBatchMessage);
      });
      this.OperateDialogService.addBatchInfoList(this.batchDeletionInfos);

      let deletionMessage = new ConfirmationMessage(
        "MEMBER.DELETION_TITLE",
        "MEMBER.DELETION_SUMMARY",
        nameArr.join(","),
        m,
        ConfirmationTargets.PROJECT_MEMBER,
        ConfirmationButtons.DELETE_CANCEL
      );
       this.OperateDialogService.openComfirmDialog(deletionMessage);
    }
  }

  deleteMem(members: Member[]) {
    if (members && members.length) {
      let promiseLists: any[] = [];
      members.forEach(member => {
        if (member.id === this.currentUser.user_id) {
          let findedList = this.batchDeletionInfos.find(data => data.name === member.entity_name);
          this.translate.get("BATCH.DELETED_FAILURE").subscribe(res => {
            findedList = BathInfoChanges(findedList, res, false, true);
          });
        }else {
          promiseLists.push(this.delOperate(this.projectId, member.id, member.entity_name));
        }

      });

      Promise.all(promiseLists).then(item => {
        this.selectedRow = [];
        this.retrieve(this.projectId, "");
      });
    }
  }

  delOperate(projectId: number, memberId: number, username: string) {
    let findedList = this.batchDeletionInfos.find(data => data.name === username);
    return this.memberService
        .deleteMember(projectId, memberId)
        .then(
            response => {
              this.translate.get("BATCH.DELETED_SUCCESS").subscribe(res => {
                findedList = BathInfoChanges(findedList, res);
              });
            },
            error => {
              this.translate.get("BATCH.DELETED_FAILURE").subscribe(res => {
                findedList = BathInfoChanges(findedList, res, false, true);
              });
            }
        );
  }

  SelectedChange(): void {
    // this.forceRefreshView(5000);
  }

  doSearch(searchMember: string) {
    this.searchMember = searchMember;
    this.retrieve(this.projectId, this.searchMember);
  }

  refresh() {
    this.retrieve(this.projectId, "");
  }
}