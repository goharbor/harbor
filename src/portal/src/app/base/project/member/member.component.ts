// Copyright Project Harbor Authors
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
import { catchError, finalize, map } from 'rxjs/operators';
import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import {
    forkJoin,
    Observable,
    Subscription,
    throwError as observableThrowError,
} from 'rxjs';
import { TranslateService } from '@ngx-translate/core';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { SessionService } from '../../../shared/services/session.service';
import { SessionUser } from '../../../shared/entities/session-user';
import { AddMemberComponent } from './add-member/add-member.component';
import { AppConfigService } from '../../../services/app-config.service';
import { OperationService } from '../../../shared/components/operation/operation.service';
import {
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../shared/services';
import { ErrorHandler } from '../../../shared/units/error-handler';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../shared/components/operation/operate';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
    PAGE_SIZE_OPTIONS,
    RoleInfo,
} from '../../../shared/entities/shared.const';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import { errorHandler } from '../../../shared/units/shared.utils';
import { ConfirmationMessage } from '../../global-confirmation-dialog/confirmation-message';
import {
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../shared/units/utils';
import { MemberService } from '../../../../../ng-swagger-gen/services/member.service';
import { ClrDatagridStateInterface } from '@clr/angular';
import { ProjectMemberEntity } from '../../../../../ng-swagger-gen/models/project-member-entity';
import { AddGroupComponent } from './add-group/add-group.component';
import { RoleService } from '../../../../../ng-swagger-gen/services/role.service';
import { Role } from '../../../../../ng-swagger-gen/models/role';

@Component({
    templateUrl: 'member.component.html',
    styleUrls: ['./member.component.scss'],
})
export class MemberComponent implements OnInit, OnDestroy {
    clrPageSizeOptions: number[] = PAGE_SIZE_OPTIONS;
    members: ProjectMemberEntity[];
    projectId: number;
    roleInfo = RoleInfo;
    delSub: Subscription;

    currentUser: SessionUser;

    batchOps = 'delete';
    searchMember: string;
    selectedRow: ProjectMemberEntity[] = [];
    roleNum: number;
    isDelete = false;
    isChangeRole = false;
    loading = true;

    isChangingRole = false;
    batchChangeRoleInfos = {};
    isLdapMode: boolean;
    isHttpAuthMode: boolean;
    isOidcMode: boolean;
    roles: Role[] = [];
    currentUserRoleId: number | null = null;
    assignableRoleIds: Set<number> | null = null;
    @ViewChild(AddMemberComponent)
    addMemberComponent: AddMemberComponent;
    @ViewChild(AddGroupComponent)
    addGroupComponent: AddGroupComponent;
    hasCreateMemberPermission: boolean;
    hasUpdateMemberPermission: boolean;
    hasDeleteMemberPermission: boolean;
    page: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.MEMBER_COMPONENT
    );
    total: number = 0;
    constructor(
        private route: ActivatedRoute,
        private router: Router,
        private memberService: MemberService,
        private translate: TranslateService,
        private messageHandlerService: MessageHandlerService,
        private OperateDialogService: ConfirmationDialogService,
        private session: SessionService,
        private operationService: OperationService,
        private appConfigService: AppConfigService,
        private userPermissionService: UserPermissionService,
        private errorHandlerEntity: ErrorHandler,
        private roleService: RoleService
    ) {
        this.delSub = OperateDialogService.confirmationConfirm$.subscribe(
            message => {
                if (
                    message &&
                    message.state === ConfirmationState.CONFIRMED &&
                    message.source === ConfirmationTargets.PROJECT_MEMBER
                ) {
                    if (this.batchOps === 'delete') {
                        this.deleteMembers(message.data);
                    }
                }
            }
        );
    }

    ngOnDestroy() {
        if (this.delSub) {
            this.delSub.unsubscribe();
        }
    }

    ngOnInit() {
        // Get projectId from router-guard params snapshot.
        this.projectId = +this.route.snapshot.parent.parent.params['id'];
        // Get current user from registered resolver.
        this.currentUser = this.session.getCurrentUser();
        // get member permission rule
        this.getMemberPermissionRule(this.projectId);
        forkJoin({
            roles: this.roleService.ListRole({ page: 1, pageSize: 100 }),
            membership: this.memberService.listProjectMembers({
                projectNameOrId: this.projectId.toString(),
                entityname: this.currentUser.username,
                page: 1,
                pageSize: 5,
            }),
        }).subscribe(({ roles, membership }) => {
            this.roles = roles ?? [];
            const myEntry = membership?.find(
                m => m.entity_type === 'u' && m.entity_id === this.currentUser.user_id
            );
            this.currentUserRoleId = myEntry?.role_id ?? null;
            this.computeAssignableRoles();
        });
        if (this.appConfigService.isLdapMode()) {
            this.isLdapMode = true;
        }
        if (this.appConfigService.isHttpAuthMode()) {
            this.isHttpAuthMode = true;
        }
        if (this.appConfigService.isOidcMode()) {
            this.isOidcMode = true;
        }
    }
    doSearch(searchMember: string) {
        this.searchMember = searchMember;
        this.retrieve(this.searchMember);
    }

    refresh() {
        this.page = 1;
        this.total = 0;
        this.selectedRow = [];
        this.searchMember = null;
        this.retrieve('');
    }
    clrDgRefresh(state: ClrDatagridStateInterface) {
        this.retrieve('', state);
    }
    retrieve(username: string, state?: ClrDatagridStateInterface) {
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.MEMBER_COMPONENT,
                this.pageSize
            );
        }
        this.loading = true;
        this.selectedRow = [];
        this.memberService
            .listProjectMembersResponse({
                entityname: username,
                page: this.page,
                pageSize: this.pageSize,
                projectNameOrId: this.projectId.toString(),
            })
            .pipe(
                finalize(() => {
                    this.loading = false;
                })
            )
            .subscribe(
                response => {
                    // Get total count
                    if (response.headers) {
                        let xHeader: string =
                            response.headers.get('X-Total-Count');
                        if (xHeader) {
                            this.total = parseInt(xHeader, 0);
                        }
                    }
                    this.members = response.body || [];
                },
                error => {
                    this.messageHandlerService.handleError(error);
                }
            );
    }

    get onlySelf(): boolean {
        return (
            this.selectedRow.length === 1 &&
            this.selectedRow[0].entity_type === 'u' &&
            this.selectedRow[0].entity_id === this.currentUser.user_id
        );
    }

    member_type_toString(user_type: string) {
        if (user_type === 'u') {
            return 'MEMBER.USER_TYPE';
        } else {
            return 'MEMBER.GROUP_TYPE';
        }
    }

    // Add member
    openAddMemberModal() {
        this.addMemberComponent.openAddMemberModal();
    }

    addedMember(result: boolean) {
        this.refresh();
    }

    // Add group
    openAddGroupModal() {
        this.addGroupComponent.openAddGroupModal();
    }
    addedGroup(result: boolean) {
        this.searchMember = '';
        this.retrieve('');
    }

    changeMembersRole(members: ProjectMemberEntity[], roleId: number) {
        if (!members) {
            return;
        }

        let changeOperate = (
            projectId: number,
            member: ProjectMemberEntity
        ) => {
            return this.memberService
                .updateProjectMember({
                    projectNameOrId: this.projectId.toString(),
                    mid: member.id,
                    role: {
                        role_id: roleId,
                    },
                })
                .pipe(
                    map(() => (this.batchChangeRoleInfos[member.id] = 'done')),
                    catchError(error => {
                        this.messageHandlerService.handleError(error);
                        return observableThrowError(error);
                    })
                );
        };

        // Preparation for members role change
        this.batchChangeRoleInfos = {};
        let RoleChangeObservables: Observable<any>[] = [];
        members.forEach(member => {
            if (
                member.entity_type === 'u' &&
                member.entity_id === this.currentUser.user_id
            ) {
                return;
            }
            this.batchChangeRoleInfos[member.id] = 'pending';
            RoleChangeObservables.push(changeOperate(this.projectId, member));
        });

        forkJoin(...RoleChangeObservables).subscribe(() => {
            this.refresh();
        });
    }

    ChangeRoleOngoing(entity_id: number) {
        return this.batchChangeRoleInfos[entity_id] === 'pending';
    }

    // Delete members
    openDeleteMembersDialog(members: ProjectMemberEntity[]) {
        this.batchOps = 'delete';
        let nameArr: string[] = [];
        if (members && members.length) {
            members.forEach(data => {
                nameArr.push(data.entity_name);
            });
            let deletionMessage = new ConfirmationMessage(
                'MEMBER.DELETION_TITLE',
                'MEMBER.DELETION_SUMMARY',
                nameArr.join(','),
                members,
                ConfirmationTargets.PROJECT_MEMBER,
                ConfirmationButtons.DELETE_CANCEL
            );
            this.OperateDialogService.openComfirmDialog(deletionMessage);
        }
    }

    deleteMembers(members: ProjectMemberEntity[]) {
        if (!members) {
            return;
        }
        let memberDeletingObservables: Observable<any>[] = [];

        // Function to delete specific member
        let deleteMember = (member: ProjectMemberEntity) => {
            let operMessage = new OperateInfo();
            operMessage.name =
                member.entity_type === 'u'
                    ? 'OPERATION.DELETE_MEMBER'
                    : 'OPERATION.DELETE_GROUP';
            operMessage.data.id = member.id;
            operMessage.state = OperationState.progressing;
            operMessage.data.name = member.entity_name;

            this.operationService.publishInfo(operMessage);
            if (
                member.entity_type === 'u' &&
                member.entity_id === this.currentUser.user_id
            ) {
                this.translate.get('BATCH.DELETED_FAILURE').subscribe(res => {
                    operateChanges(operMessage, OperationState.failure, res);
                });
                return null;
            }

            return this.memberService
                .deleteProjectMember({
                    projectNameOrId: this.projectId.toString(),
                    mid: member.id,
                })
                .pipe(
                    map(response => {
                        this.translate
                            .get('BATCH.DELETED_SUCCESS')
                            .subscribe(res => {
                                operateChanges(
                                    operMessage,
                                    OperationState.success
                                );
                            });
                    }),
                    catchError(error => {
                        const message = errorHandler(error);
                        this.translate
                            .get(message)
                            .subscribe(res =>
                                operateChanges(
                                    operMessage,
                                    OperationState.failure,
                                    res
                                )
                            );
                        return observableThrowError(error);
                    })
                );
        };

        // Deleting member then wating for results
        members.forEach(member =>
            memberDeletingObservables.push(deleteMember(member))
        );

        forkJoin(...memberDeletingObservables).subscribe(
            () => {
                this.batchOps = 'idle';
                this.refresh();
            },
            error => {
                this.errorHandlerEntity.error(error);
            }
        );
    }
    getMemberRoleDisplayName(member: ProjectMemberEntity): string {
        const role = this.roles.find(r => r.id === member.role_id);
        if (role) {
            return this.getRoleDisplayName(role);
        }
        // roles not loaded yet or custom role; fall back to static map then raw name
        return this.roleInfo[member.role_id] ?? member.role_name;
    }

    get builtinRoles(): Role[] {
        return this.roles.filter(r => r.is_builtin);
    }

    get customRoles(): Role[] {
        return this.roles.filter(r => !r.is_builtin);
    }

    getRoleDisplayName(role: Role): string {
        if (!role.is_builtin) {
            return role.name;
        }
        const keys: Record<string, string> = {
            projectAdmin: 'MEMBER.PROJECT_ADMIN',
            maintainer: 'MEMBER.PROJECT_MAINTAINER',
            developer: 'MEMBER.DEVELOPER',
            guest: 'MEMBER.GUEST',
            limitedGuest: 'MEMBER.LIMITED_GUEST',
        };
        return keys[role.name] ?? role.name;
    }

    isRoleAssignable(role: Role): boolean {
        return this.assignableRoleIds === null || this.assignableRoleIds.has(role.id);
    }

    private computeAssignableRoles(): void {
        if (this.currentUser?.has_admin_role) {
            this.assignableRoleIds = null;
            return;
        }
        const myRole = this.roles.find(r => r.id === this.currentUserRoleId);
        if (!myRole) {
            this.assignableRoleIds = null;
            return;
        }
        if (myRole.is_builtin && myRole.name === 'projectAdmin') {
            this.assignableRoleIds = null;
            return;
        }
        const myPerms = this.permSet(myRole);
        const assignable = new Set<number>();
        for (const r of this.roles) {
            if (this.isSubset(r, myPerms)) {
                assignable.add(r.id);
            }
        }
        this.assignableRoleIds = assignable;
    }

    private permSet(role: Role): Set<string> {
        const s = new Set<string>();
        for (const p of role.permissions ?? []) {
            for (const a of p.access ?? []) {
                s.add(`${a.resource}:${a.action}`);
            }
        }
        return s;
    }

    private isSubset(role: Role, callerPerms: Set<string>): boolean {
        for (const p of role.permissions ?? []) {
            for (const a of p.access ?? []) {
                if (!callerPerms.has(`${a.resource}:${a.action}`)) {
                    return false;
                }
            }
        }
        return true;
    }

    getMemberPermissionRule(projectId: number): void {
        let hasCreateMemberPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.MEMBER.KEY,
                USERSTATICPERMISSION.MEMBER.VALUE.CREATE
            );
        let hasUpdateMemberPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.MEMBER.KEY,
                USERSTATICPERMISSION.MEMBER.VALUE.UPDATE
            );
        let hasDeleteMemberPermission =
            this.userPermissionService.getPermission(
                projectId,
                USERSTATICPERMISSION.MEMBER.KEY,
                USERSTATICPERMISSION.MEMBER.VALUE.DELETE
            );
        forkJoin(
            hasCreateMemberPermission,
            hasUpdateMemberPermission,
            hasDeleteMemberPermission
        ).subscribe(
            MemberRule => {
                this.hasCreateMemberPermission = MemberRule[0] as boolean;
                this.hasUpdateMemberPermission = MemberRule[1] as boolean;
                this.hasDeleteMemberPermission = MemberRule[2] as boolean;
            },
            error => this.errorHandlerEntity.error(error)
        );
    }
}
