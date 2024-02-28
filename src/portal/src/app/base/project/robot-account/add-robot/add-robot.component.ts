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
    NEW_EMPTY_ROBOT,
    onlyHasPushPermission,
    PermissionsKinds,
} from '../../../left-side-nav/system-robot-accounts/system-robot-util';
import { Robot } from '../../../../../../ng-swagger-gen/models/robot';
import { NgForm } from '@angular/forms';
import { ClrLoadingState, ClrWizard } from '@clr/angular';
import { Subject, Subscription } from 'rxjs';
import { RobotService } from '../../../../../../ng-swagger-gen/services/robot.service';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { clone, isSameArrayValue } from '../../../../shared/units/utils';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { errorHandler } from '../../../../shared/units/shared.utils';
import { PermissionSelectPanelModes } from '../../../../shared/components/robot-permissions-panel/robot-permissions-panel.component';
import { Permissions } from '../../../../../../ng-swagger-gen/models/permissions';

const MINI_SECONDS_ONE_DAY: number = 60 * 24 * 60 * 1000;

@Component({
    selector: 'add-robot',
    templateUrl: './add-robot.component.html',
    styleUrls: ['./add-robot.component.scss'],
})
export class AddRobotComponent implements OnInit, OnDestroy {
    @Input() projectId: number;
    @Input() projectName: string;
    isEditMode: boolean = false;
    originalRobotForEdit: Robot;
    @Output()
    addSuccess: EventEmitter<Robot> = new EventEmitter<Robot>();
    addRobotOpened: boolean = false;
    robot: Robot = clone(NEW_EMPTY_ROBOT);
    expirationType: string = ExpirationType.DAYS;
    isNameExisting: boolean = false;
    loading: boolean = false;
    checkNameOnGoing: boolean = false;
    @ViewChild(InlineAlertComponent)
    inlineAlertComponent: InlineAlertComponent;
    @ViewChild('robotBasicForm', { static: true }) robotBasicForm: NgForm;
    saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    private _nameSubject: Subject<string> = new Subject<string>();
    private _nameSubscription: Subscription;

    @Input()
    robotMetadata: Permissions;

    @ViewChild('wizard') wizard: ClrWizard;
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
        return this.robot.duration < -1;
    }
    inputExpiration() {
        if (+this.robot.duration === -1) {
            this.expirationType = ExpirationType.NEVER;
        } else {
            this.expirationType = ExpirationType.DAYS;
        }
    }
    changeExpirationType() {
        if (this.expirationType === ExpirationType.DAYS) {
            this.robot.duration = null;
        }
        if (this.expirationType === ExpirationType.NEVER) {
            this.robot.duration = -1;
        }
    }
    inputName() {
        this._nameSubject.next(this.robot.name);
    }

    cancel() {
        this.wizard.reset();
        this.reset();
        this.addRobotOpened = false;
    }

    reset() {
        this.open(false);
        this.robot = clone(NEW_EMPTY_ROBOT);
        this.robotBasicForm.reset();
        this.expirationType = ExpirationType.DAYS;
    }
    resetForEdit(robot: Robot) {
        this.open(true);
        this.originalRobotForEdit = clone(robot);
        this.robot = clone(robot);
        this.expirationType =
            robot.duration === -1 ? ExpirationType.NEVER : ExpirationType.DAYS;
        this.robotBasicForm.reset({
            name: this.robot.name,
            expiration: this.robot.duration,
            description: this.robot.description,
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
        return (
            this.robot?.permissions[0]?.access?.length > 0 &&
            !this.robotBasicForm.invalid
        );
    }
    canEdit() {
        if (!this.canAdd()) {
            return false;
        }
        // eslint-disable-next-line eqeqeq
        if (this.robot.duration != this.originalRobotForEdit.duration) {
            return true;
        }
        // eslint-disable-next-line eqeqeq
        if (this.robot.description != this.originalRobotForEdit.description) {
            return true;
        }
        return !isSameArrayValue(
            this.robot.permissions[0].access,
            this.originalRobotForEdit.permissions[0].access
        );
    }
    save() {
        const robot: Robot = clone(this.robot);
        robot.disable = false;
        robot.level = PermissionsKinds.PROJECT;
        robot.duration = +this.robot.duration;
        robot.permissions[0].kind = PermissionsKinds.PROJECT;
        robot.permissions[0].namespace = this.projectName;
        // Push permission must work with pull permission
        if (onlyHasPushPermission(robot.permissions[0].access)) {
            this.inlineAlertComponent.showInlineError(
                'SYSTEM_ROBOT.PUSH_PERMISSION_TOOLTIP'
            );
            return;
        }
        this.saveBtnState = ClrLoadingState.LOADING;
        if (this.isEditMode) {
            robot.disable = this.robot.disable;
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
        if (this.robot && this.robot.creation_time && this.robot.duration > 0) {
            return new Date(
                new Date(this.robot.creation_time).getTime() +
                    this.robot.duration * MINI_SECONDS_ONE_DAY
            );
        }
        return null;
    }
    shouldShowWarning(): boolean {
        return new Date() >= this.calculateExpiresAt();
    }

    clrWizardPageOnLoad() {
        this.inlineAlertComponent.close();
    }

    protected readonly PermissionSelectPanelModes = PermissionSelectPanelModes;
}
