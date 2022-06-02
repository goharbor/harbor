import {
    Component,
    EventEmitter,
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
    switchMap,
} from 'rxjs/operators';
import { Access } from '../../../../../../ng-swagger-gen/models/access';
import {
    ACTION_RESOURCE_I18N_MAP,
    ExpirationType,
    FrontAccess,
    INITIAL_ACCESSES,
    NAMESPACE_ALL_PROJECTS,
    onlyHasPushPermission,
    PermissionsKinds,
} from '../system-robot-util';
import { clone } from '../../../../shared/units/utils';
import { RobotService } from '../../../../../../ng-swagger-gen/services/robot.service';
import { ClrLoadingState } from '@clr/angular';
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

const MINI_SECONDS_ONE_DAY: number = 60 * 24 * 60 * 1000;

@Component({
    selector: 'new-robot',
    templateUrl: './new-robot.component.html',
    styleUrls: ['./new-robot.component.scss'],
})
export class NewRobotComponent implements OnInit, OnDestroy {
    i18nMap = ACTION_RESOURCE_I18N_MAP;
    isEditMode: boolean = false;
    originalRobotForEdit: Robot;
    @Output()
    addSuccess: EventEmitter<Robot> = new EventEmitter<Robot>();
    addRobotOpened: boolean = false;
    systemRobot: Robot = {};
    expirationType: string = ExpirationType.DAYS;
    systemExpirationDays: number;
    coverAll: boolean = false;
    coverAllForEdit: boolean = false;

    isNameExisting: boolean = false;
    loading: boolean = false;
    checkNameOnGoing: boolean = false;
    loadingSystemConfig: boolean = false;
    defaultAccesses: FrontAccess[] = [];
    defaultAccessesForEdit: FrontAccess[] = [];
    @ViewChild(ListAllProjectsComponent)
    listAllProjectsComponent: ListAllProjectsComponent;
    @ViewChild(InlineAlertComponent)
    inlineAlertComponent: InlineAlertComponent;
    @ViewChild('robotForm', { static: true }) robotForm: NgForm;
    saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    private _nameSubject: Subject<string> = new Subject<string>();
    private _nameSubscription: Subscription;
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
        this.listAllProjectsComponent.init(false);
        this.listAllProjectsComponent.selectedRow = [];
        this.systemRobot = {};
        this.robotForm.reset();
        this.expirationType = ExpirationType.DAYS;
        this.getSystemRobotExpiration();
    }
    resetForEdit(robot: Robot) {
        this.open(true);
        this.defaultAccesses = clone(INITIAL_ACCESSES);
        this.defaultAccesses.forEach(item => (item.checked = false));
        this.originalRobotForEdit = clone(robot);
        this.systemRobot = robot;
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
                    if (item && item.access) {
                        item.access.forEach(item2 => {
                            this.defaultAccesses.forEach(item3 => {
                                if (
                                    item3.resource === item2.resource &&
                                    item3.action === item2.action
                                ) {
                                    item3.checked = true;
                                }
                            });
                        });
                        this.defaultAccessesForEdit = clone(
                            this.defaultAccesses
                        );
                    }
                }
            });
        }
        if (!this.coverAll) {
            this.defaultAccesses.forEach(item => (item.checked = true));
        }
        this.robotForm.reset({
            name: this.systemRobot.name,
            expiration: this.systemRobot.duration,
            description: this.systemRobot.description,
            coverAll: this.coverAll,
        });
        this.coverAllForEdit = this.coverAll;
        this.listAllProjectsComponent.init(true);
        this.listAllProjectsComponent.selectedRow = [];
        const map = {};
        this.listAllProjectsComponent.projects.forEach((pro, index) => {
            if (this.systemRobot && this.systemRobot.permissions) {
                this.systemRobot.permissions.forEach(item => {
                    if (pro.name === item.namespace) {
                        item.access.forEach(acc => {
                            pro.permissions[0].access.forEach(item3 => {
                                if (
                                    item3.resource === acc.resource &&
                                    item3.action === acc.action
                                ) {
                                    item3.checked = true;
                                }
                            });
                        });
                        map[index] = true;
                        this.listAllProjectsComponent.selectedRow.push(pro);
                    }
                });
            }
        });
        this.listAllProjectsComponent.defaultAccesses.forEach(
            item => (item.checked = true)
        );
        this.listAllProjectsComponent.projects.forEach((pro, index) => {
            if (!map[index]) {
                pro.permissions[0].access.forEach(item => {
                    item.checked = true;
                });
            }
        });
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
            let flag = false;
            this.defaultAccesses.forEach(item => {
                if (item.checked) {
                    flag = true;
                }
            });
            if (flag) {
                return true;
            }
        }
        if (
            !this.listAllProjectsComponent ||
            !this.listAllProjectsComponent.selectedRow ||
            !this.listAllProjectsComponent.selectedRow.length
        ) {
            return false;
        }
        for (
            let i = 0;
            i < this.listAllProjectsComponent.selectedRow.length;
            i++
        ) {
            let flag = false;
            for (
                let j = 0;
                j <
                this.listAllProjectsComponent.selectedRow[i].permissions[0]
                    .access.length;
                j++
            ) {
                if (
                    this.listAllProjectsComponent.selectedRow[i].permissions[0]
                        .access[j].checked
                ) {
                    flag = true;
                }
            }
            if (!flag) {
                return false;
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
        if (this.coverAll !== this.coverAllForEdit) {
            if (this.coverAll) {
                let flag = false;
                this.defaultAccesses.forEach(item => {
                    if (item.checked) {
                        flag = true;
                    }
                });
                if (!flag) {
                    return false;
                }
            }
            return true;
        }
        if (this.coverAll === this.coverAllForEdit) {
            if (this.coverAll) {
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
        }
        if (
            this.systemRobot.permissions.length !==
            this.listAllProjectsComponent.selectedRow.length
        ) {
            return true;
        }
        const map = {};
        let accessFlag = true;
        this.listAllProjectsComponent.selectedRow.forEach(item => {
            this.systemRobot.permissions.forEach(item2 => {
                if (item.name === item2.namespace) {
                    map[item.name] = true;
                    if (
                        item2.access.length !==
                        this.getAccessNum(item.permissions[0].access)
                    ) {
                        accessFlag = false;
                    }
                    item2.access.forEach(arr => {
                        item.permissions[0].access.forEach(arr2 => {
                            if (
                                arr.resource === arr2.resource &&
                                arr.action === arr2.action &&
                                !arr2.checked
                            ) {
                                accessFlag = false;
                            }
                        });
                    });
                }
            });
        });
        if (!accessFlag) {
            return true;
        }
        let flag1 = true;
        this.systemRobot.permissions.forEach(item => {
            if (!map[item.namespace]) {
                flag1 = false;
            }
        });
        return !flag1;
    }
    save() {
        const robot: Robot = clone(this.systemRobot);
        robot.disable = false;
        robot.level = PermissionsKinds.SYSTEM;
        robot.duration = +this.systemRobot.duration;
        robot.permissions = [];
        if (this.coverAll) {
            const access: Access[] = [];
            this.defaultAccesses.forEach(item => {
                if (item.checked) {
                    access.push({
                        resource: item.resource,
                        action: item.action,
                    });
                }
            });
            robot.permissions.push({
                kind: PermissionsKinds.PROJECT,
                namespace: NAMESPACE_ALL_PROJECTS,
                access: access,
            });
        } else {
            this.listAllProjectsComponent.selectedRow.forEach(item => {
                const access: Access[] = [];
                item.permissions[0].access.forEach(item2 => {
                    if (item2.checked) {
                        access.push({
                            resource: item2.resource,
                            action: item2.action,
                        });
                    }
                });
                robot.permissions.push({
                    kind: PermissionsKinds.PROJECT,
                    namespace: item.name,
                    access: access,
                });
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
