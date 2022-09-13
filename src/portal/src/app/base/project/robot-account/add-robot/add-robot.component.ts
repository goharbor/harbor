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
    switchMap,
} from 'rxjs/operators';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import {
    ACTION_RESOURCE_I18N_MAP,
    ExpirationType,
    FrontAccess,
    INITIAL_ACCESSES,
    onlyHasPushPermission,
    PermissionsKinds,
} from '../../../left-side-nav/system-robot-accounts/system-robot-util';
import { Robot } from '../../../../../../ng-swagger-gen/models/robot';
import { NgForm } from '@angular/forms';
import { ClrLoadingState } from '@clr/angular';
import { Subject, Subscription } from 'rxjs';
import { RobotService } from '../../../../../../ng-swagger-gen/services/robot.service';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { clone } from '../../../../shared/units/utils';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { Access } from '../../../../../../ng-swagger-gen/models/access';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { errorHandler } from '../../../../shared/units/shared.utils';

const MINI_SECONDS_ONE_DAY: number = 60 * 24 * 60 * 1000;
@Component({
    selector: 'add-robot',
    templateUrl: './add-robot.component.html',
    styleUrls: ['./add-robot.component.scss'],
})
export class AddRobotComponent implements OnInit, OnDestroy {
    @Input() projectId: number;
    @Input() projectName: string;
    i18nMap = ACTION_RESOURCE_I18N_MAP;
    isEditMode: boolean = false;
    originalRobotForEdit: Robot;
    @Output()
    addSuccess: EventEmitter<Robot> = new EventEmitter<Robot>();
    addRobotOpened: boolean = false;
    systemRobot: Robot = {};
    expirationType: string = ExpirationType.DAYS;
    isNameExisting: boolean = false;
    loading: boolean = false;
    checkNameOnGoing: boolean = false;
    defaultAccesses: FrontAccess[] = [];
    defaultAccessesForEdit: FrontAccess[] = [];
    @ViewChild(InlineAlertComponent)
    inlineAlertComponent: InlineAlertComponent;
    @ViewChild('robotForm', { static: true }) robotForm: NgForm;
    saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    private _nameSubject: Subject<string> = new Subject<string>();
    private _nameSubscription: Subscription;
    constructor(
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
                    debounceTime(500),
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
                    switchMap(name => {
                        this.isNameExisting = false;
                        this.checkNameOnGoing = true;
                        return this.robotService
                            .ListRobot({
                                q: encodeURIComponent(
                                    `Level=${PermissionsKinds.PROJECT},ProjectID=${this.projectId},name=${this.projectName}+${name}`
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
        if (this.expirationType === ExpirationType.DAYS) {
            this.systemRobot.duration = null;
        }
        if (this.expirationType === ExpirationType.NEVER) {
            this.systemRobot.duration = -1;
        }
    }
    inputName() {
        this._nameSubject.next(this.systemRobot.name);
    }
    cancel() {
        this.addRobotOpened = false;
    }
    getPermissions(): number {
        let count: number = 0;
        this.defaultAccesses.forEach(item => {
            if (item.checked) {
                count++;
            }
        });
        return count;
    }
    chooseAccess(access: FrontAccess) {
        access.checked = !access.checked;
    }
    reset() {
        this.open(false);
        this.defaultAccesses = clone(INITIAL_ACCESSES);
        this.systemRobot = {};
        this.robotForm.reset();
        this.expirationType = ExpirationType.DAYS;
    }
    resetForEdit(robot: Robot) {
        this.open(true);
        this.defaultAccesses = clone(INITIAL_ACCESSES);
        this.defaultAccesses.forEach(item => (item.checked = false));
        this.originalRobotForEdit = clone(robot);
        this.systemRobot = robot;
        this.expirationType =
            robot.duration === -1 ? ExpirationType.NEVER : ExpirationType.DAYS;
        this.defaultAccesses.forEach(item => {
            this.systemRobot.permissions[0].access.forEach(item2 => {
                if (
                    item.resource === item2.resource &&
                    item.action === item2.action
                ) {
                    item.checked = true;
                }
            });
        });
        this.defaultAccessesForEdit = clone(this.defaultAccesses);
        this.robotForm.reset({
            name: this.systemRobot.name,
            expiration: this.systemRobot.duration,
            description: this.systemRobot.description,
        });
    }
    open(isEditMode: boolean) {
        this.isEditMode = isEditMode;
        this.addRobotOpened = true;
        this.inlineAlertComponent.close();
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
        let flag = false;
        this.defaultAccesses.forEach(item => {
            if (item.checked) {
                flag = true;
            }
        });
        if (!flag) {
            return false;
        }
        return !this.robotForm.invalid;
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
            this.getAccessNum(this.defaultAccesses) !==
            this.getAccessNum(this.defaultAccessesForEdit)
        ) {
            return true;
        }
        let flag = true;
        this.defaultAccessesForEdit.forEach(item => {
            this.defaultAccesses.forEach(item2 => {
                if (
                    item.resource === item2.resource &&
                    item.action === item2.action &&
                    item.checked !== item2.checked
                ) {
                    flag = false;
                }
            });
        });
        return !flag;
    }
    save() {
        const robot: Robot = clone(this.systemRobot);
        robot.disable = false;
        robot.level = PermissionsKinds.PROJECT;
        robot.duration = +this.systemRobot.duration;
        const access: Access[] = [];
        this.defaultAccesses.forEach(item => {
            if (item.checked) {
                access.push({
                    resource: item.resource,
                    action: item.action,
                });
            }
        });
        robot.permissions = [
            {
                namespace: this.projectName,
                kind: PermissionsKinds.PROJECT,
                access: access,
            },
        ];
        // Push permission must work with pull permission
        if (onlyHasPushPermission(access)) {
            this.inlineAlertComponent.showInlineError(
                'SYSTEM_ROBOT.PUSH_PERMISSION_TOOLTIP'
            );
            return;
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
                        this.addRobotOpened = false;
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
            opeMessage.data.name = `${this.projectName}+${robot.name}`;
            this.operationService.publishInfo(opeMessage);
            this.robotService
                .CreateRobot({
                    robot: robot,
                })
                .subscribe(
                    res => {
                        this.saveBtnState = ClrLoadingState.SUCCESS;
                        this.saveBtnState = ClrLoadingState.SUCCESS;
                        this.addSuccess.emit(res);
                        this.addRobotOpened = false;
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
    getAccessNum(access: FrontAccess[]): number {
        let count: number = 0;
        access.forEach(item => {
            if (item.checked) {
                count++;
            }
        });
        return count;
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
    isSelectAll(permissions: FrontAccess[]): boolean {
        if (permissions?.length) {
            return (
                permissions.filter(item => item.checked).length <
                permissions.length / 2
            );
        }
        return false;
    }
    selectAllOrUnselectAll(permissions: FrontAccess[]) {
        if (permissions?.length) {
            if (this.isSelectAll(permissions)) {
                permissions.forEach(item => {
                    item.checked = true;
                });
            } else {
                permissions.forEach(item => {
                    item.checked = false;
                });
            }
        }
    }
}
