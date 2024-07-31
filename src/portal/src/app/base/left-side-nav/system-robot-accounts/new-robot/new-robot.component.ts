import {
    Component,
    EventEmitter,
    Input,
    OnDestroy,
    OnInit,
    Output,
    ViewChild,
} from '@angular/core';
import { ConfigurationService } from '../../../../services/config.service';
import { Robot } from '../../../../../../ng-swagger-gen/models/robot';
import { ListAllProjectsComponent } from '../list-all-projects/list-all-projects.component';
import { NgForm } from '@angular/forms';
import {
    debounceTime,
    distinctUntilChanged,
    filter,
    finalize,
    map,
    switchMap,
} from 'rxjs/operators';
import {
    ExpirationType,
    getSystemAccess,
    NAMESPACE_ALL_PROJECTS,
    NAMESPACE_SYSTEM,
    NEW_EMPTY_ROBOT,
    onlyHasPushPermission,
    PermissionsKinds,
} from '../system-robot-util';
import {
    clone,
    isSameArrayValue,
    isSameObject,
} from '../../../../shared/units/utils';
import { RobotService } from '../../../../../../ng-swagger-gen/services/robot.service';
import { ClrLoadingState, ClrWizard } from '@clr/angular';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { Subject, Subscription } from 'rxjs';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { errorHandler } from '../../../../shared/units/shared.utils';
import { RobotPermission } from '../../../../../../ng-swagger-gen/models/robot-permission';
import { PermissionSelectPanelModes } from '../../../../shared/components/robot-permissions-panel/robot-permissions-panel.component';
import { Permissions } from '../../../../../../ng-swagger-gen/models/permissions';

const MINI_SECONDS_ONE_DAY: number = 60 * 24 * 60 * 1000;

@Component({
    selector: 'new-robot',
    templateUrl: './new-robot.component.html',
    styleUrls: ['./new-robot.component.scss'],
})
export class NewRobotComponent implements OnInit, OnDestroy {
    isEditMode: boolean = false;
    originalRobotForEdit: Robot;
    @Output()
    addSuccess: EventEmitter<Robot> = new EventEmitter<Robot>();
    addRobotOpened: boolean = false;
    systemRobot: Robot = clone(NEW_EMPTY_ROBOT);
    expirationType: string = ExpirationType.DAYS;
    systemExpirationDays: number;
    coverAll: boolean = false;
    coverAllForEdit: boolean = false;

    isNameExisting: boolean = false;
    loading: boolean = false;
    checkNameOnGoing: boolean = false;
    loadingSystemConfig: boolean = false;
    @ViewChild(ListAllProjectsComponent)
    listAllProjectsComponent: ListAllProjectsComponent;
    @ViewChild(InlineAlertComponent)
    inlineAlertComponent: InlineAlertComponent;
    @ViewChild('robotForm', { static: true }) robotForm: NgForm;
    saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    private _nameSubject: Subject<string> = new Subject<string>();
    private _nameSubscription: Subscription;

    @Input()
    robotMetadata: Permissions;

    permissionForCoverAll: RobotPermission = {
        access: [],
        kind: PermissionsKinds.PROJECT,
        namespace: NAMESPACE_ALL_PROJECTS,
    };

    permissionForCoverAllForEdit: RobotPermission;

    permissionForSystem: RobotPermission = {
        access: [],
        kind: PermissionsKinds.SYSTEM,
        namespace: NAMESPACE_SYSTEM,
    };

    permissionForSystemForEdit: RobotPermission;
    showPage3: boolean = false;
    @ViewChild('wizard') wizard: ClrWizard;
    constructor(
        private configService: ConfigurationService,
        private robotService: RobotService,
        private msgHandler: MessageHandlerService,
        private operationService: OperationService
    ) {}
    ngOnInit(): void {
        this.subscribeName();
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
                            this.originalRobotForEdit &&
                            this.originalRobotForEdit.name === name
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
                        return this.robotService
                            .ListRobot({
                                q: encodeURIComponent(`name=${name}`),
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
    isExpirationInvalid(): boolean {
        return this.systemRobot.duration < -1;
    }
    inputExpiration() {
        if (+this.systemRobot.duration === -1) {
            this.expirationType = ExpirationType.NEVER;
        } else {
            this.expirationType = ExpirationType.DAYS;
        }
    }
    changeExpirationType() {
        if (this.expirationType === ExpirationType.DEFAULT) {
            this.systemRobot.duration = this.systemExpirationDays;
        }
        if (this.expirationType === ExpirationType.DAYS) {
            this.systemRobot.duration = this.systemExpirationDays;
        }
        if (this.expirationType === ExpirationType.NEVER) {
            this.systemRobot.duration = -1;
        }
    }
    getSystemRobotExpiration() {
        this.loadingSystemConfig = true;
        this.configService
            .getConfiguration()
            .pipe(finalize(() => (this.loadingSystemConfig = false)))
            .subscribe(res => {
                if (
                    res &&
                    res.robot_token_duration &&
                    res.robot_token_duration.value
                ) {
                    this.systemRobot.duration = res.robot_token_duration.value;
                    this.systemExpirationDays = this.systemRobot.duration;
                }
            });
    }
    inputName() {
        this._nameSubject.next(this.systemRobot.name);
    }
    cancel() {
        this.wizard.reset();
        this.reset();
        this.addRobotOpened = false;
    }

    reset() {
        this.open(false);
        this.systemRobot = clone(NEW_EMPTY_ROBOT);
        this.permissionForCoverAll = {
            access: [],
            kind: PermissionsKinds.PROJECT,
            namespace: NAMESPACE_ALL_PROJECTS,
        };
        this.permissionForSystem = {
            access: [],
            kind: PermissionsKinds.SYSTEM,
            namespace: NAMESPACE_SYSTEM,
        };
        this.coverAll = false;
        this.showPage3 = false;
        this.robotForm.reset();
        this.expirationType = ExpirationType.DAYS;
        this.getSystemRobotExpiration();
    }
    resetForEdit(robot: Robot) {
        this.open(true);
        this.originalRobotForEdit = clone(robot);
        this.systemRobot = clone(robot);
        this.permissionForSystem = {
            access: getSystemAccess(robot),
            kind: PermissionsKinds.SYSTEM,
            namespace: NAMESPACE_SYSTEM,
        };

        this.permissionForSystemForEdit = clone(this.permissionForSystem);

        this.expirationType =
            robot.duration === -1 ? ExpirationType.NEVER : ExpirationType.DAYS;
        if (robot && robot.permissions && robot.permissions.length) {
            this.coverAll = false;
            robot.permissions.forEach(item => {
                if (
                    item.kind === PermissionsKinds.PROJECT &&
                    item.namespace === NAMESPACE_ALL_PROJECTS
                ) {
                    this.coverAll = true;
                    this.permissionForCoverAll = clone(item);
                    this.permissionForCoverAllForEdit = clone(item);
                }
            });
        }
        this.robotForm.reset({
            name: this.systemRobot.name,
            expiration: this.systemRobot.duration,
            description: this.systemRobot.description,
        });
        this.coverAllForEdit = this.coverAll;
    }
    open(isEditMode: boolean) {
        this.isNameExisting = false;
        this.isEditMode = isEditMode;
        this.addRobotOpened = true;
        this.inlineAlertComponent.close();
        this._nameSubject.next('');
    }
    disabled(): boolean {
        if (!this.isEditMode) {
            return !this.canAdd();
        }
        return !this.canEdit();
    }
    canAdd(): boolean {
        if (this.robotForm.invalid) {
            return false;
        }
        if (this.coverAll) {
            if (!this.permissionForCoverAll.access?.length) {
                return false;
            }
        } else {
            if (
                !this.permissionForSystem?.access?.length &&
                !this.listAllProjectsComponent?.selectedRow?.length
            ) {
                return false;
            }
            if (this.listAllProjectsComponent?.selectedRow?.length) {
                for (
                    let i = 0;
                    i < this.listAllProjectsComponent?.selectedRow?.length;
                    i++
                ) {
                    if (
                        !this.listAllProjectsComponent
                            ?.selectedProjectPermissionMap[
                            this.listAllProjectsComponent?.selectedRow[i].name
                        ]?.length
                    ) {
                        return false;
                    }
                }
            }
        }
        return true;
    }
    canEdit() {
        if (!this.canAdd()) {
            return false;
        }
        // eslint-disable-next-line eqeqeq
        if (this.systemRobot.duration != this.originalRobotForEdit.duration) {
            return true;
        }
        // eslint-disable-next-line eqeqeq
        if (
            this.systemRobot.description !=
            this.originalRobotForEdit.description
        ) {
            return true;
        }
        if (
            !isSameObject(
                this.permissionForSystem,
                this.permissionForSystemForEdit
            )
        ) {
            return true;
        }
        if (this.coverAll !== this.coverAllForEdit) {
            return true;
        }
        if (this.coverAll) {
            if (
                !isSameObject(
                    this.permissionForCoverAll,
                    this.permissionForCoverAllForEdit
                )
            ) {
                return true;
            }
        }
        if (this.listAllProjectsComponent) {
            if (
                !isSameArrayValue(
                    this.listAllProjectsComponent.selectedRow,
                    this.listAllProjectsComponent.selectedRowForEdit
                )
            ) {
                return true;
            } else {
                for (
                    let i = 0;
                    i < this.listAllProjectsComponent.selectedRow.length;
                    i++
                ) {
                    if (
                        !isSameArrayValue(
                            this.listAllProjectsComponent
                                .selectedProjectPermissionMap[
                                this.listAllProjectsComponent.selectedRow[i]
                                    .name
                            ],
                            this.listAllProjectsComponent
                                .selectedProjectPermissionMapForEdit[
                                this.listAllProjectsComponent.selectedRow[i]
                                    .name
                            ]
                        )
                    ) {
                        return true;
                    }
                }
            }
        }
        return false;
    }
    save() {
        const robot: Robot = clone(this.systemRobot);
        robot.disable = false;
        robot.level = PermissionsKinds.SYSTEM;
        robot.duration = +this.systemRobot.duration;
        robot.permissions = [];
        if (this.permissionForSystem?.access?.length) {
            robot.permissions.push(this.permissionForSystem);
        }
        if (this.coverAll) {
            if (this.permissionForCoverAll?.access?.length) {
                robot.permissions.push(this.permissionForCoverAll);
            }
        } else {
            this.listAllProjectsComponent.selectedRow.forEach(item => {
                if (
                    this.listAllProjectsComponent.selectedProjectPermissionMap[
                        item.name
                    ]?.length
                ) {
                    robot.permissions.push({
                        kind: PermissionsKinds.PROJECT,
                        namespace: item.name,
                        access: this.listAllProjectsComponent
                            .selectedProjectPermissionMap[item.name],
                    });
                }
            });
        }
        // Push permission must work with pull permission
        if (robot.permissions && robot.permissions.length) {
            for (let i = 0; i < robot.permissions.length; i++) {
                if (onlyHasPushPermission(robot.permissions[i].access)) {
                    this.inlineAlertComponent.showInlineError(
                        'SYSTEM_ROBOT.PUSH_PERMISSION_TOOLTIP'
                    );
                    return;
                }
            }
        }
        this.saveBtnState = ClrLoadingState.LOADING;
        if (this.isEditMode) {
            robot.disable = this.systemRobot.disable;
            const opeMessage = new OperateInfo();
            opeMessage.name = 'SYSTEM_ROBOT.UPDATE_ROBOT';
            opeMessage.data.id = robot.id;
            opeMessage.state = OperationState.progressing;
            opeMessage.data.name = robot.name;
            this.operationService.publishInfo(opeMessage);
            this.robotService
                .UpdateRobot({
                    robotId: this.originalRobotForEdit.id,
                    robot,
                })
                .subscribe(
                    res => {
                        this.saveBtnState = ClrLoadingState.SUCCESS;
                        this.addSuccess.emit(null);
                        this.cancel();
                        operateChanges(opeMessage, OperationState.success);
                        this.msgHandler.showSuccess(
                            'SYSTEM_ROBOT.UPDATE_ROBOT_SUCCESSFULLY'
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
            opeMessage.name = 'SYSTEM_ROBOT.ADD_ROBOT';
            opeMessage.data.id = robot.id;
            opeMessage.state = OperationState.progressing;
            opeMessage.data.name = robot.name;
            this.operationService.publishInfo(opeMessage);
            this.robotService
                .CreateRobot({
                    robot: robot,
                })
                .subscribe(
                    res => {
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

    calculateExpiresAt(): Date {
        if (
            this.systemRobot &&
            this.systemRobot.creation_time &&
            this.systemRobot.duration > 0
        ) {
            return new Date(
                new Date(this.systemRobot.creation_time).getTime() +
                    this.systemRobot.duration * MINI_SECONDS_ONE_DAY
            );
        }
        return null;
    }
    shouldShowWarning(): boolean {
        return new Date() >= this.calculateExpiresAt();
    }

    clrWizardPageOnLoad() {
        this.inlineAlertComponent.close();
        this.showPage3 = true;
    }

    protected readonly PermissionSelectPanelModes = PermissionSelectPanelModes;
}
