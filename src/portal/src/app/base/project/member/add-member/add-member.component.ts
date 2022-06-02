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
import { debounceTime, finalize, switchMap } from 'rxjs/operators';
import {
    Component,
    EventEmitter,
    Input,
    OnDestroy,
    OnInit,
    Output,
    ViewChild,
} from '@angular/core';
import { NgForm } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { of, Subject, Subscription } from 'rxjs';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { Project } from '../../project';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { ClrLoadingState } from '@clr/angular';
import { MemberService } from 'ng-swagger-gen/services/member.service';
import { UserService } from 'ng-swagger-gen/services/user.service';
import { UserResp } from '../../../../../../ng-swagger-gen/models/user-resp';
import { UserEntity } from '../../../../../../ng-swagger-gen/models/user-entity';

@Component({
    selector: 'add-member',
    templateUrl: 'add-member.component.html',
    styleUrls: ['add-member.component.scss'],
})
export class AddMemberComponent implements OnInit, OnDestroy {
    member: UserEntity = {};
    addMemberOpened: boolean = false;
    staticBackdrop: boolean = true;
    closable: boolean = false;
    @ViewChild('memberForm', { static: true })
    currentForm: NgForm;
    @ViewChild(InlineAlertComponent)
    inlineAlert: InlineAlertComponent;
    @Input() projectId: number;
    @Output() added = new EventEmitter<boolean>();
    isMemberNameValid: boolean = true;
    memberTooltip: string = 'MEMBER.USERNAME_IS_REQUIRED';
    nameChecker: Subject<string> = new Subject<string>();
    searcher: Subject<string> = new Subject<string>();
    nameCheckerSub: Subscription;
    searcherSub: Subscription;
    checkOnGoing: boolean = false;
    searchedUserLists: UserResp[] = [];
    btnStatus: ClrLoadingState = ClrLoadingState.DEFAULT;
    roleId: number = 1; // default value is 1(project admin)

    constructor(
        private memberService: MemberService,
        private userService: UserService,
        private messageHandlerService: MessageHandlerService,
        private route: ActivatedRoute
    ) {}

    ngOnInit(): void {
        let resolverData = this.route.snapshot.parent.parent.data;
        let hasProjectAdminRole: boolean;
        if (resolverData) {
            hasProjectAdminRole = (<Project>resolverData['projectResolver'])
                .has_project_admin_role;
        }
        if (hasProjectAdminRole) {
            if (!this.searcherSub) {
                this.searcherSub = this.searcher
                    .pipe(
                        debounceTime(500),
                        switchMap(name => {
                            if (name) {
                                return this.userService.listUsers({
                                    page: 1,
                                    pageSize: 10,
                                    q: encodeURIComponent(`username=~${name}`),
                                });
                            } else {
                                return of([]);
                            }
                        })
                    )
                    .subscribe(res => {
                        if (res) {
                            this.searchedUserLists = res;
                        }
                    });
            }
            if (!this.nameCheckerSub) {
                this.nameCheckerSub = this.nameChecker
                    .pipe(
                        debounceTime(500),
                        switchMap(name => {
                            if (name) {
                                this.checkOnGoing = true;
                                return this.memberService
                                    .listProjectMembers({
                                        page: 1,
                                        pageSize: 10,
                                        projectNameOrId:
                                            this.projectId.toString(),
                                        entityname: name,
                                    })
                                    .pipe(
                                        finalize(
                                            () => (this.checkOnGoing = false)
                                        )
                                    );
                            } else {
                                return of([]);
                            }
                        })
                    )
                    .subscribe(res => {
                        if (res && res.length) {
                            if (
                                res.filter(
                                    m => m.entity_name === this.member.username
                                ).length > 0
                            ) {
                                this.isMemberNameValid = false;
                                this.memberTooltip =
                                    'MEMBER.USERNAME_ALREADY_EXISTS';
                            }
                        }
                    });
            }
        }
    }

    ngOnDestroy(): void {
        if (this.nameCheckerSub) {
            this.nameCheckerSub.unsubscribe();
            this.nameCheckerSub = null;
        }
        if (this.searcherSub) {
            this.searcherSub.unsubscribe();
            this.searcherSub = null;
        }
    }

    onSubmit(): void {
        if (!this.member.username || this.member.username.length === 0) {
            return;
        }
        this.btnStatus = ClrLoadingState.LOADING;
        this.memberService
            .createProjectMember({
                projectNameOrId: this.projectId.toString(),
                projectMember: {
                    role_id: this.roleId,
                    member_user: this.member,
                },
            })
            .subscribe(
                () => {
                    this.addMemberOpened = false;
                    this.btnStatus = ClrLoadingState.SUCCESS;
                    this.messageHandlerService.showSuccess(
                        'MEMBER.ADDED_SUCCESS'
                    );
                    this.added.emit(true);
                },
                error => {
                    this.btnStatus = ClrLoadingState.ERROR;
                    this.inlineAlert.showInlineError(error);
                }
            );
    }

    selectedName(username: string) {
        this.member.username = username;
        this.nameChecker.next(username);
        this.searchedUserLists = [];
    }

    onCancel() {
        this.addMemberOpened = false;
    }

    leaveInput() {
        this.searchedUserLists = [];
    }
    openAddMemberModal(): void {
        this.currentForm.reset({
            member_role: 1,
        });
        this.inlineAlert.close();
        this.member = {};
        this.addMemberOpened = true;
        this.member.username = '';
        this.isMemberNameValid = true;
        this.memberTooltip = 'MEMBER.USERNAME_IS_REQUIRED';
        this.searchedUserLists = [];
    }

    handleValidation(): void {
        this.nameChecker.next(this.member.username);
        this.searcher.next(this.member.username);
        if (!this.member.username) {
            this.searchedUserLists = [];
            this.isMemberNameValid = false;
            this.memberTooltip = 'MEMBER.USERNAME_IS_REQUIRED';
        } else {
            this.isMemberNameValid = true;
        }
    }

    isValid(): boolean {
        return (
            this.currentForm &&
            this.currentForm.valid &&
            this.isMemberNameValid &&
            !this.checkOnGoing
        );
    }
}
