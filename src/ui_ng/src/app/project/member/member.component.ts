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
import {operateChanges, OperateInfo, OperationService, OperationState} from "harbor-ui";

@Component({
  templateUrl: "member.component.html",
  styleUrls: ["./member.component.scss"],
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

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private memberService: MemberService,
    private translate: TranslateService,
    private messageHandlerService: MessageHandlerService,
    private OperateDialogService: ConfirmationDialogService,
    private session: SessionService,
    private operationService: OperationService,
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
    if (this.selectedRow.length === 1 && this.selectedRow[0].entity_id === this.currentUser.user_id) {
      return true;
    }
    return false;
  }

  changeRole(m: Member[], roleId: number) {
    if (m && m.length) {
      this.isDelete = false;
      this.isChangeRole = true;
      this.roleNum = roleId;
      this.changeOpe(m);
    }
  }

  changeOpe(members: Member[]) {
    if (members && members.length) {
      let promiseList: any[] = [];
      members.forEach(member => {
        promiseList.push(this.changeOperate(this.projectId, this.roleNum, member));
      });

      Promise.all(promiseList).then(num => {
            this.retrieve(this.projectId, "");
          },
      );
    }
  }

  changeOperate(projectId: number, roleId: number, member: Member) {
    // init operation info
    let operMessage = new OperateInfo();
    operMessage.name = 'OPERATION.SWITCH_ROLE';
    operMessage.data.id = member.id;
    operMessage.state = OperationState.progressing;
    operMessage.data.name = member.entity_name;
    this.operationService.publishInfo(operMessage);

    if (member.entity_id === this.currentUser.user_id) {
      this.translate.get("BATCH.SWITCH_FAILURE").subscribe(res => {
        operateChanges(operMessage, OperationState.failure, res);
      });
      return null;
    }
    return this.memberService
        .changeMemberRole(projectId, member.id, roleId)
        .then(
            response => {
              this.translate.get("BATCH.SWITCH_SUCCESS").subscribe(res => {
                operateChanges(operMessage, OperationState.success);
              });
            },
            error => {
              this.translate.get("BATCH.SWITCH_FAILURE").subscribe(res => {
                operateChanges(operMessage, OperationState.failure, res);
              });
            }
        );
  }

  deleteMembers(m: Member[]) {
    this.isDelete = true;
    this.isChangeRole = false;
    let nameArr: string[] = [];
    if (m && m.length) {
      m.forEach(data => {
        nameArr.push(data.entity_name);
      });

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
        promiseLists.push(this.delOperate(this.projectId, member));
      });

      Promise.all(promiseLists).then(item => {
        this.selectedRow = [];
        this.retrieve(this.projectId, "");
      });
    }
  }

  delOperate(projectId: number, member: Member) {
    // init operation info
    let operMessage = new OperateInfo();
    operMessage.name = 'OPERATION.DELETE_MEMBER';
    operMessage.data.id = member.id;
    operMessage.state = OperationState.progressing;
    operMessage.data.name = member.entity_name;
    this.operationService.publishInfo(operMessage);

    if (member.entity_id === this.currentUser.user_id) {
      this.translate.get("BATCH.DELETED_FAILURE").subscribe(res => {
        operateChanges(operMessage, OperationState.failure, res);
      });
      return null;
    }

    return this.memberService
        .deleteMember(projectId, member.id)
        .then(
            response => {
              this.translate.get("BATCH.DELETED_SUCCESS").subscribe(res => {
                operateChanges(operMessage, OperationState.success);
              });
            },
            error => {
              this.translate.get("BATCH.DELETED_FAILURE").subscribe(res => {
                operateChanges(operMessage, OperationState.failure, res);
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
