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
import { AppConfigService } from '../../../../services/app-config.service';
import { ProjectRootInterface } from '../../../../shared/services';
import {
    GroupType,
    PROJECT_ROOTS,
} from '../../../../shared/entities/shared.const';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { UsergroupService } from '../../../../../../ng-swagger-gen/services/usergroup.service';
import { of, Subject, Subscription } from 'rxjs';
import { UserGroup } from 'ng-swagger-gen/models/user-group';
import { ClrLoadingState } from '@clr/angular';
import { MemberService } from 'ng-swagger-gen/services/member.service';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';

@Component({
    selector: 'add-group',
    templateUrl: './add-group.component.html',
    styleUrls: ['./add-group.component.scss'],
})
export class AddGroupComponent implements OnInit, OnDestroy {
    projectRoots: ProjectRootInterface[] = PROJECT_ROOTS;
    memberGroup: UserGroup = {
        group_name: '',
    };
    roleId: number = 1; // default value is 1(project admin);
    addGroupOpened: boolean = false;
    staticBackdrop: boolean = true;
    closable: boolean = false;

    @ViewChild('groupForm', { static: true })
    currentForm: NgForm;

    @ViewChild(InlineAlertComponent)
    inlineAlert: InlineAlertComponent;

    @Input() projectId: number;
    @Output() added = new EventEmitter<boolean>();

    checkOnGoing: boolean = false;
    searchedGroups: UserGroup[] = [];
    groupChecker: Subject<string> = new Subject<string>();
    groupSearcher: Subject<string> = new Subject<string>();
    groupCheckerSub: Subscription;
    groupSearcherSub: Subscription;
    btnStatus: ClrLoadingState = ClrLoadingState.DEFAULT;
    isGroupNameValid: boolean = true;
    groupTooltip: string = 'MEMBER.GROUP_NAME_REQUIRED';
    isNameChecked: boolean = false; // this is only for LDAP mode
    constructor(
        private memberService: MemberService,
        private appConfigService: AppConfigService,
        private messageHandlerService: MessageHandlerService,
        private userGroupService: UsergroupService
    ) {}

    ngOnInit(): void {
        if (!this.groupCheckerSub) {
            this.groupCheckerSub = this.groupChecker
                .pipe(
                    debounceTime(500),
                    switchMap(name => {
                        if (name) {
                            this.checkOnGoing = true;
                            const params: MemberService.ListProjectMembersParams =
                                {
                                    projectNameOrId: this.projectId.toString(),
                                    page: 1,
                                    pageSize: 10,
                                    entityname: name,
                                };
                            return this.memberService
                                .listProjectMembers(params)
                                .pipe(
                                    finalize(() => (this.checkOnGoing = false))
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
                                g =>
                                    g.entity_name ===
                                    this.memberGroup.group_name
                            ).length > 0
                        ) {
                            this.isGroupNameValid = false;
                            this.groupTooltip = 'MEMBER.GROUP_ALREADY_ADDED';
                        }
                    }
                });
        }
        if (!this.groupSearcherSub) {
            this.groupSearcherSub = this.groupSearcher
                .pipe(
                    debounceTime(500),
                    switchMap(name => {
                        if (name) {
                            return this.userGroupService.searchUserGroups({
                                page: 1,
                                pageSize: 10,
                                groupname: name,
                            });
                        } else {
                            return of([]);
                        }
                    })
                )
                .subscribe(res => {
                    if (res) {
                        this.searchedGroups = res;
                    }
                    // for LDAP mode, if input group name is not found from search result, then show "Group name does not exists" error
                    if (
                        this.appConfigService.isLdapMode() &&
                        this.memberGroup.group_name
                    ) {
                        let flag = false;
                        this.searchedGroups.forEach(item => {
                            if (
                                item.group_name === this.memberGroup.group_name
                            ) {
                                flag = true;
                            }
                        });
                        if (!flag) {
                            this.isGroupNameValid = false;
                            this.groupTooltip = 'MEMBER.NON_EXISTENT_GROUP';
                        } else {
                            // it means input group name is valid
                            this.isNameChecked = true;
                        }
                    }
                });
        }
    }
    ngOnDestroy() {
        if (this.groupCheckerSub) {
            this.groupCheckerSub.unsubscribe();
            this.groupCheckerSub = null;
        }
        if (this.groupSearcherSub) {
            this.groupSearcherSub.unsubscribe();
            this.groupSearcherSub = null;
        }
    }

    createGroupAsMember() {
        this.btnStatus = ClrLoadingState.LOADING;
        if (this.appConfigService.isHttpAuthMode()) {
            this.memberGroup.group_type = GroupType.HTTP_TYPE;
        }
        if (this.appConfigService.isLdapMode()) {
            this.memberGroup.group_type = GroupType.LDAP_TYPE;
        }
        if (this.appConfigService.isOidcMode()) {
            this.memberGroup.group_type = GroupType.OIDC_TYPE;
        }
        this.memberService
            .createProjectMember({
                projectNameOrId: this.projectId.toString(),
                projectMember: {
                    role_id: this.roleId,
                    member_group: this.memberGroup,
                },
            })
            .subscribe(
                res => {
                    this.messageHandlerService.showSuccess(
                        'MEMBER.ADDED_SUCCESS'
                    );
                    this.btnStatus = ClrLoadingState.SUCCESS;
                    this.addGroupOpened = false;
                    this.added.emit(true);
                },
                err => {
                    this.btnStatus = ClrLoadingState.ERROR;
                    this.inlineAlert.showInlineError(err);
                    this.added.emit(false);
                }
            );
    }
    onSubmit(): void {
        this.createGroupAsMember();
    }

    onCancel() {
        this.addGroupOpened = false;
    }

    openAddGroupModal(): void {
        this.currentForm.reset({
            member_role: 1,
        });
        this.addGroupOpened = true;
        this.inlineAlert.close();
        this.memberGroup = {
            group_name: '',
        };
        this.isGroupNameValid = true;
        this.groupTooltip = 'MEMBER.USERNAME_IS_REQUIRED';
        this.searchedGroups = [];
    }
    isValid(): boolean {
        if (this.appConfigService.isLdapMode()) {
            if (!this.isNameChecked) {
                return false;
            }
        }
        return (
            this.isGroupNameValid &&
            this.currentForm &&
            this.currentForm.valid &&
            !this.checkOnGoing
        );
    }

    selectGroup(groupName) {
        if (this.appConfigService.isLdapMode()) {
            this.isNameChecked = true;
            this.isGroupNameValid = true;
        }
        this.memberGroup.group_name = groupName;
        this.groupChecker.next(groupName);
        this.searchedGroups = [];
    }

    leaveInput() {
        this.searchedGroups = [];
    }

    input() {
        if (this.appConfigService.isLdapMode()) {
            this.isNameChecked = false;
        }
        this.groupChecker.next(this.memberGroup.group_name);
        this.groupSearcher.next(this.memberGroup.group_name);
        if (!this.memberGroup.group_name) {
            this.searchedGroups = [];
            this.isGroupNameValid = false;
            this.groupTooltip = 'MEMBER.GROUP_NAME_REQUIRED';
        } else {
            this.isGroupNameValid = true;
        }
    }
}
