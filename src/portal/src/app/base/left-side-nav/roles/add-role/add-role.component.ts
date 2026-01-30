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
import {
    Component,
    OnInit,
    Input,
    OnDestroy,
    Output,
    EventEmitter,
    ViewChild,
} from '@angular/core';
import {
    debounceTime,
    distinctUntilChanged,
    filter,
    finalize,
    map,
    switchMap,
} from 'rxjs/operators';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import {
    ExpirationType,
    NEW_EMPTY_ROLE,
    onlyHasPushPermission,
    PermissionsKinds,
} from '../roles-util';
import { Role } from '../../../../../../ng-swagger-gen/models/role';
import { NgForm } from '@angular/forms';
import { ClrLoadingState, ClrWizard } from '@clr/angular';
import { Subject, Subscription } from 'rxjs';
import { RoleService } from '../../../../../../ng-swagger-gen/services/role.service';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { clone, isSameArrayValue } from '../../../../shared/units/utils';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { errorHandler } from '../../../../shared/units/shared.utils';
import { PermissionSelectPanelModes } from '../../../../shared/components/role-permissions-panel/role-permissions-panel.component';
import { Permissions } from '../../../../../../ng-swagger-gen/models/permissions';

const MINI_SECONDS_ONE_DAY: number = 60 * 24 * 60 * 1000;

@Component({
    selector: 'add-role',
    templateUrl: './add-role.component.html',
    styleUrls: ['./add-role.component.scss'],
})
export class AddRoleComponent implements OnInit, OnDestroy {
    @Input() projectId: number;
    @Input() projectName: string;
    isEditMode: boolean = false;
    originalRoleForEdit: Role;
    @Output()
    addSuccess: EventEmitter<Role> = new EventEmitter<Role>();
    addRoleOpened: boolean = false;
    role: Role = clone(NEW_EMPTY_ROLE);
    expirationType: string = ExpirationType.DAYS;
    isNameExisting: boolean = false;
    loading: boolean = false;
    checkNameOnGoing: boolean = false;
    @ViewChild(InlineAlertComponent)
    inlineAlertComponent: InlineAlertComponent;
    @ViewChild('roleBasicForm', { static: true }) roleBasicForm: NgForm;
    saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    private _nameSubject: Subject<string> = new Subject<string>();
    private _nameSubscription: Subscription;

    @Input()
    roleMetadata: Permissions;

    @ViewChild('wizard') wizard: ClrWizard;
    constructor(
        private roleService: RoleService,
        private msgHandler: MessageHandlerService,
        private operationService: OperationService
    ) {}
    ngOnInit(): void {
        this.subscribeName();
        console.log("init new role");

    }
    ngOnDestroy() {
        if (this._nameSubscription) {
            this._nameSubscription.unsubscribe();
            this._nameSubscription = null;
        }
    }
    subscribeName() {
        if (!this._nameSubscription) {
            this._nameSubscription = this._nameSubject
                .pipe(
                    distinctUntilChanged(),
                    filter(name => {
                        if (
                            this.isEditMode &&
                            this.originalRoleForEdit &&
                            this.originalRoleForEdit.name === name
                        ) {
                            return false;
                        }
                        return name?.length > 0;
                    }),
                    map(name => {
                        this.checkNameOnGoing = !!name;
                        return name;
                    }),
                    debounceTime(500),
                    switchMap(name => {
                        this.isNameExisting = false;
                        this.checkNameOnGoing = true;
                        return this.roleService
                            .ListRole({
                                q: encodeURIComponent(
                                    `Level=${PermissionsKinds.ROLE},ProjectID=${this.projectId},name=${name}`
                                ),
                            })
                            .pipe(
                                finalize(() => (this.checkNameOnGoing = false))
                            );
                    })
                )
                .subscribe(res => {
                    if (res && res.length > 0) {
                        this.isNameExisting = true;
                    }
                });
        }
    }
    inputName() {
        this._nameSubject.next(this.role.name);
    }

    cancel() {
        this.wizard.reset();
        this.reset();
        this.addRoleOpened = false;
    }

    reset() {
        this.open(false);
        this.role = clone(NEW_EMPTY_ROLE);
        //this.roleBasicForm.reset();
        this.expirationType = ExpirationType.DAYS;
    }
    resetForEdit(role: Role) {
        this.open(true);
        this.originalRoleForEdit = clone(role);
        this.role = clone(role);
        this.roleBasicForm.reset({
            name: this.role.name,
        });
    }
    open(isEditMode: boolean) {
        this.isEditMode = isEditMode;
        this.addRoleOpened = true;
        //this.inlineAlertComponent.close();
        this.isNameExisting = false;
        this._nameSubject.next('');
    }
    disabled(): boolean {
        if (!this.isEditMode) {
            return !this.canAdd();
        }
        return !this.canEdit();
    }
    canAdd(): boolean {
        return (
            this.role?.permissions[0]?.access?.length > 0 &&
            !this.roleBasicForm.invalid
        );
    }
    canEdit() {
        if (!this.canAdd()) {
            return false;
        }
        // eslint-disable-next-line eqeqeq
        return !isSameArrayValue(
            this.role.permissions[0].access,
            this.originalRoleForEdit.permissions[0].access
        );
    }
    save() {
        const role: Role = clone(this.role);
        role.permissions[0].kind = PermissionsKinds.ROLE;
        role.permissions[0].namespace = this.projectName;
        // Push permission must work with pull permission
        if (onlyHasPushPermission(role.permissions[0].access)) {
            this.inlineAlertComponent.showInlineError(
                'ROLE.PUSH_PERMISSION_TOOLTIP'
            );
            return;
        }
        this.saveBtnState = ClrLoadingState.LOADING;
        if (this.isEditMode) {
            const opeMessage = new OperateInfo();
            opeMessage.name = 'ROLE.UPDATE_ROLE';
            opeMessage.data.id = role.id;
            opeMessage.state = OperationState.progressing;
            opeMessage.data.name = role.name;
            this.operationService.publishInfo(opeMessage);
            this.roleService
                .UpdateRole({
                    roleId: this.originalRoleForEdit.id,
                    role,
                })
                .subscribe(
                    res => {
                        this.saveBtnState = ClrLoadingState.SUCCESS;
                        this.addSuccess.emit(null);
                        this.cancel();
                        operateChanges(opeMessage, OperationState.success);
                        this.msgHandler.showSuccess(
                            'ROLE.UPDATE_ROLE_SUCCESSFULLY'
                        );
                    },
                    error => {
                        this.saveBtnState = ClrLoadingState.ERROR;
                        operateChanges(
                            opeMessage,
                            OperationState.failure,
                            errorHandler(error)
                        );
                        this.inlineAlertComponent.showInlineError(error);
                    }
                );
        } else {
            const opeMessage = new OperateInfo();
            opeMessage.name = 'ROLE.ADD_ROLE';
            opeMessage.data.id = role.id;
            opeMessage.state = OperationState.progressing;
            opeMessage.data.name = `${this.projectName}+${role.name}`;
            this.operationService.publishInfo(opeMessage);
            this.roleService
                .CreateRole({
                    role: role,
                })
                .subscribe(
                    res => {
                        this.saveBtnState = ClrLoadingState.SUCCESS;
                        this.saveBtnState = ClrLoadingState.SUCCESS;
                        this.addSuccess.emit(res);
                        this.cancel();
                        operateChanges(opeMessage, OperationState.success);
                    },
                    error => {
                        this.saveBtnState = ClrLoadingState.ERROR;
                        this.inlineAlertComponent.showInlineError(error);
                        operateChanges(
                            opeMessage,
                            OperationState.failure,
                            errorHandler(error)
                        );
                    }
                );
        }
    }


    clrWizardPageOnLoad() {
        this.inlineAlertComponent.close();
    }

    protected readonly PermissionSelectPanelModes = PermissionSelectPanelModes;
}
