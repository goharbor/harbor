import { Component, EventEmitter, Output, ViewChild } from '@angular/core';
import { RobotService } from '../../../../../ng-swagger-gen/services/robot.service';
import { ClrLoadingState } from '@clr/angular';
import { Robot } from '../../../../../ng-swagger-gen/models/robot';
import { clone } from '../../units/utils';
import { NgForm } from '@angular/forms';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../operation/operate';
import { OperationService } from '../operation/operation.service';
import { MessageHandlerService } from '../../services/message-handler.service';
import { DomSanitizer, SafeUrl } from '@angular/platform-browser';
import { TranslateService } from '@ngx-translate/core';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';
import { errorHandler } from '../../units/shared.utils';
import { CopyInputComponent } from '../push-image/copy-input.component';

@Component({
    selector: 'view-token',
    templateUrl: './view-token.component.html',
    styleUrls: ['./view-token.component.scss'],
})
export class ViewTokenComponent {
    showNewPwd: boolean = false;
    showConfirmPwd: boolean = false;
    tokenModalOpened: boolean = false;
    robot: Robot;
    newSecret: string;
    confirmSecret: string;
    btnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    @ViewChild(InlineAlertComponent)
    inlineAlertComponent: InlineAlertComponent;
    @ViewChild('secretForm', { static: true }) secretForm: NgForm;
    @ViewChild('copyInputComponent')
    copyInputComponent: CopyInputComponent;
    @Output()
    refreshSuccess: EventEmitter<boolean> = new EventEmitter<boolean>();
    copyToken: boolean = false;
    createSuccess: string;
    downLoadFileName: string = '';
    downLoadHref: SafeUrl = '';
    enableNewSecret: boolean = false;
    constructor(
        private robotService: RobotService,
        private operationService: OperationService,
        private msgHandler: MessageHandlerService,
        private sanitizer: DomSanitizer,
        private translate: TranslateService
    ) {}

    cancel() {
        this.tokenModalOpened = false;
    }
    open() {
        this.showNewPwd = false;
        this.showConfirmPwd = false;
        this.tokenModalOpened = true;
        this.inlineAlertComponent.close();
        this.copyToken = false;
        this.createSuccess = null;
        this.newSecret = null;
        this.confirmSecret = null;
        this.downLoadFileName = '';
        this.downLoadHref = '';
        this.secretForm.reset();
    }
    refreshToken() {
        this.btnState = ClrLoadingState.LOADING;
        const robot: Robot = clone(this.robot);
        const opeMessage = new OperateInfo();
        opeMessage.name = 'SYSTEM_ROBOT.REFRESH_SECRET';
        opeMessage.data.id = robot.id;
        opeMessage.state = OperationState.progressing;
        opeMessage.data.name = robot.name;
        this.operationService.publishInfo(opeMessage);
        if (this.newSecret) {
            robot.secret = this.newSecret;
        }
        this.robotService
            .RefreshSec({
                robotId: robot.id,
                robotSec: {
                    secret: robot.secret,
                },
            })
            .subscribe(
                res => {
                    this.btnState = ClrLoadingState.SUCCESS;
                    operateChanges(opeMessage, OperationState.success);
                    this.refreshSuccess.emit(true);
                    this.cancel();
                    if (res && res.secret) {
                        this.robot.secret = res.secret;
                        this.copyToken = true;
                        this.createSuccess =
                            'SYSTEM_ROBOT.REFRESH_SECRET_SUCCESS';
                        // export to token file
                        robot.secret = res.secret;
                        const downLoadUrl = `data:text/json;charset=utf-8, ${encodeURIComponent(
                            JSON.stringify(robot)
                        )}`;
                        this.downLoadHref =
                            this.sanitizer.bypassSecurityTrustUrl(downLoadUrl);
                        this.downLoadFileName = `${robot.name}.json`;
                    } else {
                        this.msgHandler.showSuccess(
                            'SYSTEM_ROBOT.REFRESH_SECRET_SUCCESS'
                        );
                    }
                },
                error => {
                    this.btnState = ClrLoadingState.ERROR;
                    this.inlineAlertComponent.showInlineError(error);
                    operateChanges(
                        opeMessage,
                        OperationState.failure,
                        errorHandler(error)
                    );
                }
            );
    }
    canRefresh() {
        if (this.enableNewSecret && !this.newSecret && !this.confirmSecret) {
            return false;
        }
        if (!this.newSecret && !this.confirmSecret) {
            return true;
        }
        return (
            this.newSecret &&
            this.confirmSecret &&
            this.newSecret === this.confirmSecret &&
            this.secretForm.valid
        );
    }
    onCpSuccess($event: any): void {
        this.copyToken = false;
        this.tokenModalOpened = false;
        if (this.copyInputComponent) {
            this.copyInputComponent.reset();
        }
        this.translate
            .get('ROBOT_ACCOUNT.COPY_SUCCESS', { param: this.robot.name })
            .subscribe((res: string) => {
                this.msgHandler.showSuccess(res);
            });
    }

    closeModal() {
        this.copyToken = false;
        this.tokenModalOpened = false;
    }

    notSame(): boolean {
        return (
            this.secretForm.valid &&
            this.newSecret &&
            this.confirmSecret &&
            this.newSecret !== this.confirmSecret
        );
    }
    changeEnableNewSecret() {
        this.secretForm.reset({
            enableNewSecret: this.enableNewSecret,
        });
    }
}
