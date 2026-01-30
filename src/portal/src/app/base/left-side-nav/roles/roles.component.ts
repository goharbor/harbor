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
import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core';

import { ViewTokenComponent } from '../../../shared/components/view-token/view-token.component';
import { RoleService } from '../../../../../ng-swagger-gen/services/role.service';
import { Role } from '../../../../../ng-swagger-gen/models/role';
import {
    clone,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../shared/units/utils';
import { ClrDatagridStateInterface, ClrLoadingState } from '@clr/angular';
import {
    catchError,
    debounceTime,
    distinctUntilChanged,
    finalize,
    map,
    switchMap,
} from 'rxjs/operators';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import {
    FrontRole,
    getRoleAccess,
    NAMESPACE_ALL_PROJECTS,
    NEW_EMPTY_ROLE,
    PermissionsKinds,
} from './roles-util';
import { forkJoin, Observable, of, Subscription } from 'rxjs';
import { FilterComponent } from '../../../shared/components/filter/filter.component';
import { HttpErrorResponse } from '@angular/common/http';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../shared/components/operation/operate';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { DomSanitizer } from '@angular/platform-browser';
import { TranslateService } from '@ngx-translate/core';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
    PAGE_SIZE_OPTIONS,
} from '../../../shared/entities/shared.const';
import { errorHandler } from '../../../shared/units/shared.utils';
import { ConfirmationMessage } from '../../global-confirmation-dialog/confirmation-message';
import { RolePermission } from '../../../../../ng-swagger-gen/models/role-permission';
import { SysteminfoService } from '../../../../../ng-swagger-gen/services/systeminfo.service';
import { Access } from '../../../../../ng-swagger-gen/models/access';
import { PermissionSelectPanelModes } from '../../../shared/components/role-permissions-panel/role-permissions-panel.component';
import { PermissionsService } from '../../../../../ng-swagger-gen/services/permissions.service';
import { Permissions } from '../../../../../ng-swagger-gen/models/permissions';
import {AddRoleComponent } from './add-role/add-role.component'

@Component({
    selector: 'roles',
    templateUrl: './roles.component.html',
    styleUrls: ['./roles.component.scss'],
})
export class RolesComponent implements OnInit, OnDestroy {
    clrPageSizeOptions: number[] = PAGE_SIZE_OPTIONS;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.SYSTEM_ROBOT_COMPONENT
        //TODO create page size for ROBOT
    );
    currentPage: number = 1;
    total: number = 0;
    roles: FrontRole[] = [];
    selectedRows: FrontRole[] = [];
    loading: boolean = true;
    loadingData: boolean = false;
    addBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    @ViewChild(AddRoleComponent)
    newRoleComponent: AddRoleComponent;
    @ViewChild(FilterComponent, { static: true })
    filterComponent: FilterComponent;
    searchSub: Subscription;
    searchKey: string;
    subscription: Subscription;
    deltaTime: number; // the different between server time and local time

    roleMetadata: Permissions;
    loadingMetadata: boolean = false;
    constructor(
        private roleService: RoleService,
        private msgHandler: MessageHandlerService,
        private operateDialogService: ConfirmationDialogService,
        private operationService: OperationService,
        private sanitizer: DomSanitizer,
        private translate: TranslateService,
        private systemInfoService: SysteminfoService,
        private permissionService: PermissionsService
    ) {}
    ngOnInit() {
        this.getRolePermissions();
        this.getCurrenTime();
        if (!this.searchSub) {
            this.searchSub = this.filterComponent.filterTerms
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged(),
                    switchMap(roleSearchName => {
                        this.currentPage = 1;
                        this.selectedRows = [];
                        const queryParam: RoleService.ListRoleParams = {
                            page: this.currentPage,
                            pageSize: this.pageSize,
                        };
                        this.searchKey = roleSearchName;
                        if (this.searchKey) {
                            queryParam.q = encodeURIComponent(
                                `name=~${this.searchKey}`
                            );
                        }
                        this.loading = true;
                        return this.roleService
                            .ListRoleResponse(queryParam)
                            .pipe(
                                finalize(() => {
                                    this.loading = false;
                                })
                            );
                    })
                )
                .subscribe(
                    response => {
                        this.total = Number.parseInt(
                            response.headers.get('x-total-count'),
                            10
                        );
                        this.roles = response.body as Role[];
                    },
                    error => {
                        this.msgHandler.handleError(error);
                    }
                );
        }
        if (!this.subscription) {
            this.subscription =
                this.operateDialogService.confirmationConfirm$.subscribe(
                    message => {
                        if (
                            message &&
                            message.state === ConfirmationState.CONFIRMED &&
                            message.source === ConfirmationTargets.ROLE
                        ) {
                            this.deleteRoles(message.data);
                        }
                    }
                );
        }
    }

    ngAfterViewInit() {
        console.log("new role component: " + this.newRoleComponent);
    }


    ngOnDestroy() {
        if (this.searchSub) {
            this.searchSub.unsubscribe();
            this.searchSub = null;
        }
        if (this.subscription) {
            this.subscription.unsubscribe();
            this.subscription = null;
        }
    }

    getRolePermissions() {
        this.loadingData = true;
        this.permissionService
            .getPermissions()
            .pipe(finalize(() => (this.loadingData = false)))
            .subscribe(res => {
                this.roleMetadata = res;
            });
    }

    getCurrenTime() {
        this.systemInfoService.getSystemInfo().subscribe(res => {
            if (res?.current_time) {
                this.deltaTime =
                    new Date().getTime() -
                    new Date(res?.current_time).getTime();
            }
        });
    }

    clrLoad(state?: ClrDatagridStateInterface) {
        if (state && state.page && state.page.size) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.SYSTEM_ROBOT_COMPONENT,
                this.pageSize
            );
        }
        this.selectedRows = [];
        const queryParam: RoleService.ListRoleParams = {
            page: this.currentPage,
            pageSize: this.pageSize,
            sort: getSortingString(state),
        };
        if (this.searchKey) {
            queryParam.q = encodeURIComponent(`name=~${this.searchKey}`);
        }
        this.loading = true;
        this.roleService
            .ListRoleResponse(queryParam)
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                response => {
                    this.total = Number.parseInt(
                        response.headers.get('x-total-count'),
                        10
                    );
                    this.roles = response.body as Role[];
                },
                err => {
                    this.msgHandler.error(err);
                }
            );
    }
    openNewRoleModal(isEditMode: boolean) {
        if (isEditMode) {
            this.newRoleComponent.resetForEdit(clone(this.selectedRows[0]));
        } else {
            this.newRoleComponent.reset();
        }
    }

    getProjects(r: Role): RolePermission[] {
        const arr = [];
        if (r && r.permissions && r.permissions.length) {
            for (let i = 0; i < r.permissions.length; i++) {
                if (r.permissions[i].kind === PermissionsKinds.ROLE) {
                    arr.push(r.permissions[i]);
                }
            }
        }
        return arr;
    }

    refresh() {
        this.currentPage = 1;
        this.selectedRows = [];
        this.clrLoad();
    }
    deleteRoles(roles: Role[]) {
        let observableLists: Observable<any>[] = [];
        if (roles && roles.length) {
            roles.forEach(item => {
                observableLists.push(this.deleteRole(item));
            });
            forkJoin(...observableLists).subscribe(resArr => {
                let error;
                if (resArr && resArr.length) {
                    resArr.forEach(item => {
                        if (item instanceof HttpErrorResponse) {
                            error = errorHandler(item);
                        }
                    });
                }
                if (error) {
                    this.msgHandler.handleError(error);
                } else {
                    this.msgHandler.showSuccess(
                        'ROLE.DELETE_ROLE_SUCCESS'
                    );
                }
                this.refresh();
            });
        }
    }
    deleteRole(role: Role): Observable<any> {
        let operMessage = new OperateInfo();
        operMessage.name = 'ROLE.DELETE_ROLE';
        operMessage.data.id = role.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = role.name;
        this.operationService.publishInfo(operMessage);
        return this.roleService.DeleteRole({ roleId: role.id }).pipe(
            map(() => {
                operateChanges(operMessage, OperationState.success);
            }),
            catchError(error => {
                const message = errorHandler(error);
                operateChanges(operMessage, OperationState.failure, message);
                return of(error);
            })
        );
    }
    openDeleteRolesDialog() {
        const roleNames = this.selectedRows.map(role => role.name).join(',');
        const deletionMessage = new ConfirmationMessage(
            'ROLE.DELETION_TITLE',
            'ROLE.DELETION_SUMMARY',
            roleNames,
            this.selectedRows,
            ConfirmationTargets.ROLE,
            ConfirmationButtons.DELETE_CANCEL
        );
        this.operateDialogService.openComfirmDialog(deletionMessage);
    }

    addSuccess(role: Role) {
        if (role) {
            this.translate
                .get('ROLE.CREATED_SUCCESS', { param: role.name });
            // export to token file
            const downLoadUrl = `data:text/json;charset=utf-8, ${encodeURIComponent(
                JSON.stringify(role)
            )}`;
        }
        this.refresh();
    }




    getRoleAccess(r: Role): Access[] {
        return getRoleAccess(r);
    }

    protected readonly NEW_EMPTY_ROLE = NEW_EMPTY_ROLE;
    protected readonly PermissionSelectPanelModes = PermissionSelectPanelModes;
}
