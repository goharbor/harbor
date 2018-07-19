import { ChangeDetectorRef, ChangeDetectionStrategy, ViewChild } from "@angular/core";
import { Component, OnInit, Input, Output, EventEmitter } from "@angular/core";
import { NgForm } from '@angular/forms';

import { forkJoin } from "rxjs/observable/forkJoin";
import { Observable } from "rxjs/Observable";
import "rxjs/observable/of";
import { TranslateService } from '@ngx-translate/core';

import "rxjs/observable/timer";
import {operateChanges, OperateInfo, OperationService, OperationState} from "harbor-ui";

import { UserGroup } from "./../../../group/group";
import { MemberService } from "./../member.service";
import { GroupService } from "../../../group/group.service";
import { ProjectRoles } from "../../../shared/shared.const";
import { MessageHandlerService } from '../../../shared/message-handler/message-handler.service';
import { Member } from "../member";

@Component({
  selector: "add-group",
  templateUrl: "./add-group.component.html",
  styleUrls: ["./add-group.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class AddGroupComponent implements OnInit {
  opened = false;
  createGroupMode = false;
  onLoading = false;
  roles = ProjectRoles;
  currentTerm = '';

  selectedRole = 1;
  group = new UserGroup();
  selectedGroups: UserGroup[] = [];
  groups: UserGroup[] = [];

  dnTooltip = 'TOOLTIP.ITEM_REQUIRED';

  @Input() projectId: number;
  @Input() memberList: Member[] = [];
  @Output() added = new EventEmitter<boolean>();

  @ViewChild('groupForm')
  groupForm: NgForm;

  constructor(
    private translateService: TranslateService,
    private msgHandler: MessageHandlerService,
    private operationService: OperationService,
    private ref: ChangeDetectorRef,
    private groupService: GroupService,
    private memberService: MemberService
  ) {}

  ngOnInit() { }

  public get isValid(): boolean {
    if (this.createGroupMode) {
      return this.groupForm && this.groupForm.valid;
    } else {
      return this.selectedGroups.length > 0;
    }
  }
  public get isDNInvalid(): boolean {
    if (!this.groupForm) {return false; };
    let dnControl = this.groupForm.controls['ldap_group_dn'];
    return  dnControl && dnControl.invalid && (dnControl.dirty || dnControl.touched);
  }

  loadGroups() {
    this.onLoading = true;
    this.groupService.getUserGroups().subscribe(groups => {
      this.groups = groups.filter(group => {
        if (!group.group_name) {group.group_name = ''; };
        return group.group_name.includes(this.currentTerm)
        && !this.memberList.some(member => member.entity_type === 'g' && member.entity_id === group.id);
      });
      this.onLoading = false;
      this.ref.detectChanges();
    });
  }

  doFilter(name: string) {
    this.currentTerm = name;
    this.loadGroups();
  }

  resetModaldata() {
    this.group = new UserGroup();
    this.selectedRole = 1;
    this.selectedGroups = [];
    this.groups = [];
  }

  public open() {
    this.resetModaldata();
    this.loadGroups();
    this.opened = true;
    this.ref.detectChanges();
  }

  public close() {
    this.resetModaldata();
    this.opened = false;
  }

  onSave() {
    if (!this.createGroupMode) {
      this.addGroups();
    } else {
      this.createGroupAsMember();
    }
  }

  onCancel() {
    this.opened = false;
  }

  addGroups() {
    let GroupAdders$ = this.selectedGroups.map(group => {
      let operMessage = new OperateInfo();
      operMessage.name = 'OPERATION.ADD_GROUP';
      operMessage.data.id = group.id;
      operMessage.state = OperationState.progressing;
      operMessage.data.name = group.group_name;
      this.operationService.publishInfo(operMessage);
      return this.memberService
        .addGroupMember(this.projectId, group, this.selectedRole)
        .flatMap(response => {
           return this.translateService.get("BATCH.DELETED_SUCCESS")
           .flatMap(res => {
            operateChanges(operMessage, OperationState.success);
            return Observable.of(res);
           }); })
           .catch(error => {
            return this.translateService.get("BATCH.DELETED_FAILURE")
            .flatMap(res => {
              operateChanges(operMessage, OperationState.failure, res);
              return Observable.of(res);
            }); })
        .catch(error => Observable.of(error.status));
      });
    forkJoin(GroupAdders$)
      .subscribe(results => {
        if (results.some(code => code < 200 || code > 299)) {
          this.added.emit(false);
        } else {
          this.added.emit(true);
        }
      });
    this.opened = false;
  }

  createGroupAsMember() {
    let groupCopy = Object.assign({}, this.group);
    this.memberService.addGroupMember(this.projectId, groupCopy, this.selectedRole)
    .subscribe(
      res => this.added.emit(true),
      err => {
        this.msgHandler.handleError(err);
        this.added.emit(false);
      }
    );
    this.opened = false;
  }
}
