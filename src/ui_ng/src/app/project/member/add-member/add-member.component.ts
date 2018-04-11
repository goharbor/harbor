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
import {
  Component,
  Input,
  EventEmitter,
  Output,
  ViewChild,
  AfterViewChecked,
  OnInit,
  OnDestroy, ChangeDetectionStrategy, ChangeDetectorRef
} from '@angular/core';
import { Response } from '@angular/http';
import { NgForm } from '@angular/forms';

import { MemberService } from '../member.service';
import { UserService } from '../../../user/user.service';

import { MessageHandlerService } from '../../../shared/message-handler/message-handler.service';
import { InlineAlertComponent } from '../../../shared/inline-alert/inline-alert.component';

import { TranslateService } from '@ngx-translate/core';

import { Member } from '../member';

import { Subject } from 'rxjs/Subject';
import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/distinctUntilChanged';
import {User} from "../../../user/user";
import {ActivatedRoute, Router} from "@angular/router";
import {Project} from "../../project";

@Component({
  selector: 'add-member',
  templateUrl: 'add-member.component.html',
  styleUrls: ['add-member.component.css'],
  providers: [UserService],
  changeDetection: ChangeDetectionStrategy.Default
})
export class AddMemberComponent implements AfterViewChecked, OnInit, OnDestroy {

  member: Member = new Member();

  addMemberOpened: boolean;

  memberForm: NgForm;

  staticBackdrop: boolean = true;
  closable: boolean = false;

  @ViewChild('memberForm')
  currentForm: NgForm;

  hasChanged: boolean;

  @ViewChild(InlineAlertComponent)
  inlineAlert: InlineAlertComponent;

  @Input() projectId: number;
  @Output() added = new EventEmitter<boolean>();

  isMemberNameValid: boolean = true;
  memberTooltip: string = 'MEMBER.USERNAME_IS_REQUIRED';
  nameChecker: Subject<string> = new Subject<string>();
  checkOnGoing: boolean = false;
  selectUserName: string[] = [];
  userLists: User[];

  constructor(private memberService: MemberService,
    private userService: UserService,
    private messageHandlerService: MessageHandlerService,
    private translateService: TranslateService,
    private route: ActivatedRoute,
    private ref: ChangeDetectorRef) { }

  ngOnInit(): void {
  let resolverData = this.route.snapshot.parent.data;
  let hasProjectAdminRole: boolean;
  if (resolverData) {
    hasProjectAdminRole = (<Project>resolverData['projectResolver']).has_project_admin_role;
  }
  if (hasProjectAdminRole) {
    this.userService.getUsers()
        .then(users => {
          this.userLists = users;
        });

    this.nameChecker
        .debounceTime(500)
        .distinctUntilChanged()
        .subscribe((name: string) => {
          let cont = this.currentForm.controls['member_name'];
          if (cont) {
            this.isMemberNameValid = cont.valid;
            if (cont.valid) {
              this.checkOnGoing = true;
              this.memberService
                  .listMembers(this.projectId, cont.value).toPromise()
                  .then((members: Member[]) => {
                    if (members.filter(m => { return m.entity_name === cont.value }).length > 0) {
                      this.isMemberNameValid = false;
                      this.memberTooltip = 'MEMBER.USERNAME_ALREADY_EXISTS';
                    }
                    this.checkOnGoing = false;
                  })
                  .catch(error => {
                    this.checkOnGoing = false;
                  });
              //username autocomplete
              if (this.userLists && this.userLists.length) {
                this.selectUserName = [];
                this.userLists.filter(data => {
                  if (data.username.startsWith(cont.value)) {
                    if (this.selectUserName.length < 10) {
                      this.selectUserName.push(data.username);
                    }
                  }
                });
                setTimeout(() => {
                  setInterval(() => this.ref.markForCheck(), 100);
                }, 1000);
              }
            } else {
              this.memberTooltip = 'MEMBER.USERNAME_IS_REQUIRED';
            }
          }
        });
  }

  }

  ngOnDestroy(): void {
    this.nameChecker.unsubscribe();
  }

  onSubmit(): void {
    if (!this.member.entity_name || this.member.entity_name.length === 0) { return; }
    this.memberService
      .addMember(this.projectId, this.member.entity_name, +this.member.role_id)
      .subscribe(
      response => {
        this.messageHandlerService.showSuccess('MEMBER.ADDED_SUCCESS');
        this.added.emit(true);
        this.addMemberOpened = false;
      },
      error => {
        if (error instanceof Response) {
          let errorMessageKey: string;
          switch (error.status) {
            case 404:
              errorMessageKey = 'MEMBER.USERNAME_DOES_NOT_EXISTS';
              break;
            case 409:
              errorMessageKey = 'MEMBER.USERNAME_ALREADY_EXISTS';
              break;
            default:
              errorMessageKey = 'MEMBER.UNKNOWN_ERROR';
          }
          if (this.messageHandlerService.isAppLevel(error)) {
            this.messageHandlerService.handleError(error);
            this.addMemberOpened = false;
          } else {
            this.translateService
              .get(errorMessageKey)
              .subscribe(errorMessage => this.inlineAlert.showInlineError(errorMessage));
          }
        }
      }
      );

    setTimeout(() => {
      setInterval(() => this.ref.markForCheck(), 100);
    }, 1000);
  }

  selectedName(username: string) {
    this.member.entity_name = username;
    this.selectUserName = [];
  }

  onCancel() {
    if (this.hasChanged) {
      this.inlineAlert.showInlineConfirmation({ message: 'ALERT.FORM_CHANGE_CONFIRMATION' });
    } else {
      this.addMemberOpened = false;
      this.memberForm.reset();
    }
  }

  leaveInput() {
    this.selectUserName = [];
  }
  ngAfterViewChecked(): void {
    if (this.memberForm !== this.currentForm) {
      this.memberForm = this.currentForm;
    }
    if (this.memberForm) {
      this.memberForm.valueChanges.subscribe(data => {
        let memberName = data['member_name'];
        if (memberName && memberName !== '') {
          this.hasChanged = true;
          this.inlineAlert.close();
        } else {
          this.hasChanged = false;
        }
      });
    }
  }

  confirmCancel(confirmed: boolean) {
    this.addMemberOpened = false;
    this.inlineAlert.close();
    this.memberForm.reset();
  }

  openAddMemberModal(): void {
    this.currentForm.reset();
    this.member = new Member();
    this.addMemberOpened = true;
    this.hasChanged = false;
    this.member.role_id = 1;
    this.member.entity_name = '';
    this.isMemberNameValid = true;
    this.memberTooltip = 'MEMBER.USERNAME_IS_REQUIRED';
    this.selectUserName = [];
  }

  handleValidation(): void {
    let cont = this.currentForm.controls['member_name'];
    if (cont) {
      this.nameChecker.next(cont.value);
    }
  }

  public get isValid(): boolean {
    return this.currentForm && 
    this.currentForm.valid && 
    this.isMemberNameValid &&
    !this.checkOnGoing;
  }
}