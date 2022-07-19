import { of, Subscription, forkJoin } from 'rxjs';
import {
    flatMap,
    catchError,
    finalize,
    debounceTime,
    switchMap,
    filter,
} from 'rxjs/operators';
import { SessionService } from '../../../shared/services/session.service';
import { TranslateService } from '@ngx-translate/core';
import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';
import { AddGroupModalComponent } from './add-group-modal/add-group-modal.component';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { throwError as observableThrowError } from 'rxjs';
import { AppConfigService } from '../../../services/app-config.service';
import { OperationService } from '../../../shared/components/operation/operation.service';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../shared/components/operation/operate';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
    GroupType,
} from '../../../shared/entities/shared.const';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import { errorHandler } from '../../../shared/units/shared.utils';
import { ConfirmationMessage } from '../../global-confirmation-dialog/confirmation-message';
import { ClrDatagridStateInterface } from '@clr/angular';
import {
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../shared/units/utils';
import { UsergroupService } from '../../../../../ng-swagger-gen/services/usergroup.service';
import { UserGroup } from '../../../../../ng-swagger-gen/models/user-group';
import { FilterComponent } from '../../../shared/components/filter/filter.component';

@Component({
    selector: 'app-group',
    templateUrl: './group.component.html',
    styleUrls: ['./group.component.scss'],
})
export class GroupComponent implements OnInit, OnDestroy {
    loading = true;
    groups: UserGroup[] = [];
    currentPage: number = 1;
    totalCount: number = 0;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.SYSTEM_GROUP_COMPONENT
    );
    selectedGroups: UserGroup[] = [];
    currentTerm = '';
    delSub: Subscription;
    batchOps = 'idle';
    isLdapMode: boolean;

    @ViewChild(AddGroupModalComponent) newGroupModal: AddGroupModalComponent;
    searchSub: Subscription;
    @ViewChild(FilterComponent, { static: true })
    filterComponent: FilterComponent;
    constructor(
        private operationService: OperationService,
        private translate: TranslateService,
        private operateDialogService: ConfirmationDialogService,
        private groupService: UsergroupService,
        private msgHandler: MessageHandlerService,
        private session: SessionService,
        private translateService: TranslateService,
        private appConfigService: AppConfigService
    ) {}

    ngOnInit() {
        if (this.appConfigService.isLdapMode()) {
            this.isLdapMode = true;
        }
        this.delSub = this.operateDialogService.confirmationConfirm$.subscribe(
            message => {
                if (
                    message &&
                    message.state === ConfirmationState.CONFIRMED &&
                    message.source === ConfirmationTargets.PROJECT_MEMBER
                ) {
                    if (this.batchOps === 'delete') {
                        this.deleteGroups();
                    }
                }
            }
        );
        if (!this.searchSub) {
            this.searchSub = this.filterComponent.filterTerms
                .pipe(
                    filter(groupName => !!groupName),
                    debounceTime(500),
                    switchMap(groupName => {
                        this.currentPage = 1;
                        this.selectedGroups = [];
                        this.loading = true;
                        return this.groupService
                            .listUserGroupsResponse({
                                groupName: groupName,
                                pageSize: this.pageSize,
                                page: this.currentPage,
                            })
                            .pipe(
                                finalize(() => {
                                    this.loading = false;
                                })
                            );
                    })
                )
                .subscribe(
                    response => {
                        this.totalCount = Number.parseInt(
                            response.headers.get('x-total-count'),
                            10
                        );
                        this.groups = response.body as UserGroup[];
                    },
                    error => {
                        this.msgHandler.handleError(error);
                    }
                );
        }
    }
    ngOnDestroy(): void {
        this.delSub.unsubscribe();
        if (this.searchSub) {
            this.searchSub.unsubscribe();
            this.searchSub = null;
        }
    }

    refresh(): void {
        this.currentPage = 1;
        this.selectedGroups = [];
        this.currentTerm = '';
        this.filterComponent.currentValue = '';
        this.loadData();
    }

    loadData(state?: ClrDatagridStateInterface): void {
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.SYSTEM_GROUP_COMPONENT,
                this.pageSize
            );
        }
        this.loading = true;
        if (this.currentTerm) {
            this.groupService
                .searchUserGroupsResponse({
                    groupname: encodeURIComponent(this.currentTerm),
                    page: this.currentPage,
                    pageSize: this.pageSize,
                })
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    response => {
                        this.totalCount = Number.parseInt(
                            response.headers.get('x-total-count'),
                            10
                        );
                        this.groups = response.body as UserGroup[];
                    },
                    err => {
                        this.msgHandler.error(err);
                    }
                );
        } else {
            this.groupService
                .listUserGroupsResponse({
                    page: this.currentPage,
                    pageSize: this.pageSize,
                })
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    response => {
                        this.totalCount = Number.parseInt(
                            response.headers.get('x-total-count'),
                            10
                        );
                        this.groups = response.body as UserGroup[];
                    },
                    err => {
                        this.msgHandler.error(err);
                    }
                );
        }
    }

    addGroup(): void {
        this.newGroupModal.open();
    }

    editGroup(): void {
        this.newGroupModal.open(this.selectedGroups[0], true);
    }

    openDeleteConfirmationDialog(): void {
        // open delete modal
        this.batchOps = 'delete';
        let nameArr: string[] = [];
        if (this.selectedGroups.length > 0) {
            this.selectedGroups.forEach(group => {
                nameArr.push(group.group_name);
            });
            // batchInfo.id = group.id;
            let deletionMessage = new ConfirmationMessage(
                'GROUP.DELETION_TITLE',
                'GROUP.DELETION_SUMMARY',
                nameArr.join(','),
                this.selectedGroups,
                ConfirmationTargets.PROJECT_MEMBER,
                ConfirmationButtons.DELETE_CANCEL
            );
            this.operateDialogService.openComfirmDialog(deletionMessage);
        }
    }

    deleteGroups() {
        let obs = this.selectedGroups.map(group => {
            let operMessage = new OperateInfo();
            operMessage.name = 'OPERATION.DELETE_GROUP';
            operMessage.data.id = group.id;
            operMessage.state = OperationState.progressing;
            operMessage.data.name = group.group_name;

            this.operationService.publishInfo(operMessage);
            return this.groupService
                .deleteUserGroup({
                    groupId: group.id,
                })
                .pipe(
                    flatMap(response => {
                        return this.translate.get('BATCH.DELETED_SUCCESS').pipe(
                            flatMap(res => {
                                operateChanges(
                                    operMessage,
                                    OperationState.success
                                );
                                return of(res);
                            })
                        );
                    })
                )
                .pipe(
                    catchError(error => {
                        const message = errorHandler(error);
                        this.translateService
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
        });

        forkJoin(obs).subscribe(
            res => {
                this.selectedGroups = [];
                this.batchOps = 'idle';
                this.loadData();
            },
            err => this.msgHandler.handleError(err)
        );
    }

    groupToSring(type: number) {
        if (type === GroupType.LDAP_TYPE) {
            return 'GROUP.LDAP_TYPE';
        } else if (type === GroupType.HTTP_TYPE) {
            return 'GROUP.HTTP_TYPE';
        } else if (type === GroupType.OIDC_TYPE) {
            return 'GROUP.OIDC_TYPE';
        } else {
            return 'UNKNOWN';
        }
    }

    doFilter(groupName: string): void {
        if (!groupName) {
            this.currentTerm = groupName;
            this.loadData();
        }
    }
    get canAddGroup(): boolean {
        return this.session.currentUser.has_admin_role;
    }

    get canEditGroup(): boolean {
        return (
            this.selectedGroups.length === 1 &&
            this.session.currentUser.has_admin_role &&
            this.isLdapMode
        );
    }
    get canDeleteGroup(): boolean {
        return (
            this.selectedGroups.length >= 1 &&
            this.session.currentUser.has_admin_role
        );
    }
}
