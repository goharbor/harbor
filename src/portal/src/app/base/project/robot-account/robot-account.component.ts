import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { ClrDatagridStateInterface, ClrLoadingState } from '@clr/angular';
import {
    catchError,
    debounceTime,
    distinctUntilChanged,
    finalize,
    map,
    switchMap,
} from 'rxjs/operators';
import { forkJoin, Observable, of, Subscription } from 'rxjs';
import {
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../shared/services';
import {
    ACTION_RESOURCE_I18N_MAP,
    PermissionsKinds,
} from '../../left-side-nav/system-robot-accounts/system-robot-util';
import {
    clone,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../shared/units/utils';
import { ViewTokenComponent } from '../../../shared/components/view-token/view-token.component';
import { FilterComponent } from '../../../shared/components/filter/filter.component';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { RobotService } from '../../../../../ng-swagger-gen/services/robot.service';
import { Robot } from '../../../../../ng-swagger-gen/models/robot';
import { ActivatedRoute } from '@angular/router';
import { Project } from '../../../../../ng-swagger-gen/models/project';
import { HttpErrorResponse } from '@angular/common/http';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../shared/components/operation/operate';
import { AddRobotComponent } from './add-robot/add-robot.component';
import { TranslateService } from '@ngx-translate/core';
import { DomSanitizer } from '@angular/platform-browser';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../shared/entities/shared.const';
import { errorHandler } from '../../../shared/units/shared.utils';
import { ConfirmationMessage } from '../../global-confirmation-dialog/confirmation-message';
import { SysteminfoService } from '../../../../../ng-swagger-gen/services/systeminfo.service';

@Component({
    selector: 'app-robot-account',
    templateUrl: './robot-account.component.html',
    styleUrls: ['./robot-account.component.scss'],
})
export class RobotAccountComponent implements OnInit, OnDestroy {
    i18nMap = ACTION_RESOURCE_I18N_MAP;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.PROJECT_ROBOT_COMPONENT
    );
    currentPage: number = 1;
    total: number = 0;
    robots: Robot[] = [];
    selectedRows: Robot[] = [];
    loading: boolean = true;
    addBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    @ViewChild(AddRobotComponent)
    newRobotComponent: AddRobotComponent;
    @ViewChild(ViewTokenComponent)
    viewTokenComponent: ViewTokenComponent;
    @ViewChild(FilterComponent, { static: true })
    filterComponent: FilterComponent;
    searchSub: Subscription;
    searchKey: string;
    subscription: Subscription;
    hasRobotCreatePermission: boolean;
    hasRobotUpdatePermission: boolean;
    hasRobotDeletePermission: boolean;
    hasRobotReadPermission: boolean;
    projectId: number;
    projectName: string;
    deltaTime: number; // the different between server time and local time
    constructor(
        private robotService: RobotService,
        private msgHandler: MessageHandlerService,
        private operateDialogService: ConfirmationDialogService,
        private operationService: OperationService,
        private userPermissionService: UserPermissionService,
        private route: ActivatedRoute,
        private translate: TranslateService,
        private sanitizer: DomSanitizer,
        private systemInfoService: SysteminfoService
    ) {}
    ngOnInit() {
        this.getCurrenTime();
        this.projectId = +this.route.snapshot.parent.parent.params['id'];
        let resolverData = this.route.snapshot.parent.parent.data;
        if (resolverData) {
            let project = <Project>resolverData['projectResolver'];
            this.projectName = project.name;
        }
        this.getPermissionsList();
        if (!this.searchSub) {
            this.searchSub = this.filterComponent.filterTerms
                .pipe(
                    debounceTime(500),
                    distinctUntilChanged(),
                    switchMap(robotSearchName => {
                        this.currentPage = 1;
                        this.selectedRows = [];
                        const queryParam: RobotService.ListRobotParams = {
                            page: this.currentPage,
                            pageSize: this.pageSize,
                        };
                        this.searchKey = robotSearchName;
                        if (this.searchKey) {
                            queryParam.q = encodeURIComponent(
                                `Level=${PermissionsKinds.PROJECT},ProjectID=${this.projectId},name=~${this.searchKey}`
                            );
                        } else {
                            queryParam.q = encodeURIComponent(
                                `Level=${PermissionsKinds.PROJECT},ProjectID=${this.projectId}`
                            );
                        }
                        this.loading = true;
                        return this.robotService
                            .ListRobotResponse(queryParam)
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
                        this.robots = response.body as Robot[];
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
                            message.source ===
                                ConfirmationTargets.PROJECT_ROBOT_ACCOUNT
                        ) {
                            this.deleteRobots(message.data);
                        }
                        if (
                            message.state === ConfirmationState.CONFIRMED &&
                            message.source ===
                                ConfirmationTargets.PROJECT_ROBOT_ACCOUNT_ENABLE_OR_DISABLE
                        ) {
                            this.operateRobot();
                        }
                    }
                );
        }
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
    getPermissionsList(): void {
        let permissionsList = [];
        permissionsList.push(
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.ROBOT.KEY,
                USERSTATICPERMISSION.ROBOT.VALUE.CREATE
            )
        );
        permissionsList.push(
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.ROBOT.KEY,
                USERSTATICPERMISSION.ROBOT.VALUE.UPDATE
            )
        );
        permissionsList.push(
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.ROBOT.KEY,
                USERSTATICPERMISSION.ROBOT.VALUE.DELETE
            )
        );
        permissionsList.push(
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.ROBOT.KEY,
                USERSTATICPERMISSION.ROBOT.VALUE.READ
            )
        );

        forkJoin(...permissionsList).subscribe(
            Rules => {
                this.hasRobotCreatePermission = Rules[0] as boolean;
                this.hasRobotUpdatePermission = Rules[1] as boolean;
                this.hasRobotDeletePermission = Rules[2] as boolean;
                this.hasRobotReadPermission = Rules[3] as boolean;
            },
            error => this.msgHandler.error(error)
        );
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
    clrLoad(state?: ClrDatagridStateInterface) {
        if (state && state.page && state.page.size) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.PROJECT_ROBOT_COMPONENT,
                this.pageSize
            );
        }
        this.selectedRows = [];
        const queryParam: RobotService.ListRobotParams = {
            page: this.currentPage,
            pageSize: this.pageSize,
            sort: getSortingString(state),
            q: encodeURIComponent(
                `Level=${PermissionsKinds.PROJECT},ProjectID=${this.projectId}`
            ),
        };
        if (this.searchKey) {
            queryParam.q += encodeURIComponent(`,name=~${this.searchKey}`);
        }
        this.loading = true;
        this.robotService
            .ListRobotResponse(queryParam)
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                response => {
                    this.total = Number.parseInt(
                        response.headers.get('x-total-count'),
                        10
                    );
                    this.robots = response.body as Robot[];
                },
                err => {
                    this.msgHandler.error(err);
                }
            );
    }
    openNewRobotModal(isEditMode: boolean) {
        if (isEditMode) {
            this.newRobotComponent.resetForEdit(clone(this.selectedRows[0]));
        } else {
            this.newRobotComponent.reset();
        }
    }
    openTokenModal() {
        this.viewTokenComponent.open();
        this.viewTokenComponent.robot = clone(this.selectedRows[0]);
    }
    refresh() {
        this.currentPage = 1;
        this.selectedRows = [];
        this.clrLoad();
    }
    deleteRobots(robots: Robot[]) {
        let observableLists: Observable<any>[] = [];
        if (robots && robots.length) {
            robots.forEach(item => {
                observableLists.push(this.deleteRobot(item));
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
                        'SYSTEM_ROBOT.DELETE_ROBOT_SUCCESS'
                    );
                }
                this.refresh();
            });
        }
    }
    deleteRobot(robot: Robot): Observable<any> {
        let operMessage = new OperateInfo();
        operMessage.name = 'SYSTEM_ROBOT.DELETE_ROBOT';
        operMessage.data.id = robot.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = robot.name;
        this.operationService.publishInfo(operMessage);
        return this.robotService.DeleteRobot({ robotId: robot.id }).pipe(
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
    openDeleteRobotsDialog() {
        const robotNames = this.selectedRows.map(robot => robot.name).join(',');
        const deletionMessage = new ConfirmationMessage(
            'ROBOT_ACCOUNT.DELETION_TITLE',
            'ROBOT_ACCOUNT.DELETION_SUMMARY',
            robotNames,
            this.selectedRows,
            ConfirmationTargets.PROJECT_ROBOT_ACCOUNT,
            ConfirmationButtons.DELETE_CANCEL
        );
        this.operateDialogService.openComfirmDialog(deletionMessage);
    }

    disableOrEnable() {
        const title: string = this.selectedRows[0].disable
            ? 'SYSTEM_ROBOT.ENABLE_TITLE'
            : 'SYSTEM_ROBOT.DISABLE_TITLE';
        const summary: string = this.selectedRows[0].disable
            ? 'SYSTEM_ROBOT.ENABLE_SUMMARY'
            : 'SYSTEM_ROBOT.DISABLE_SUMMARY';
        const deletionMessage = new ConfirmationMessage(
            title,
            summary,
            this.selectedRows[0].name,
            this.selectedRows[0],
            ConfirmationTargets.PROJECT_ROBOT_ACCOUNT_ENABLE_OR_DISABLE,
            this.selectedRows[0].disable
                ? ConfirmationButtons.ENABLE_CANCEL
                : ConfirmationButtons.DISABLE_CANCEL
        );
        this.operateDialogService.openComfirmDialog(deletionMessage);
    }

    operateRobot() {
        const robot: Robot = clone(this.selectedRows[0]);
        const successMessage: string = robot.disable
            ? 'SYSTEM_ROBOT.ENABLE_ROBOT_SUCCESSFULLY'
            : 'SYSTEM_ROBOT.DISABLE_ROBOT_SUCCESSFULLY';
        robot.disable = !robot.disable;
        delete robot.secret;
        const opeMessage = new OperateInfo();
        opeMessage.name = robot.disable
            ? 'SYSTEM_ROBOT.DISABLE_TITLE'
            : 'SYSTEM_ROBOT.ENABLE_TITLE';
        opeMessage.data.id = robot.id;
        opeMessage.state = OperationState.progressing;
        opeMessage.data.name = robot.name;
        this.operationService.publishInfo(opeMessage);
        this.robotService
            .UpdateRobot({
                robot: robot,
                robotId: robot.id,
            })
            .subscribe(
                res => {
                    operateChanges(opeMessage, OperationState.success);
                    this.msgHandler.showSuccess(successMessage);
                    this.refresh();
                },
                error => {
                    operateChanges(
                        opeMessage,
                        OperationState.failure,
                        errorHandler(error)
                    );
                    this.msgHandler.showSuccess(error);
                }
            );
    }
    addSuccess(robot: Robot) {
        if (robot) {
            this.viewTokenComponent.open();
            this.viewTokenComponent.tokenModalOpened = false;
            this.viewTokenComponent.robot = clone(robot);
            this.viewTokenComponent.copyToken = true;
            this.translate
                .get('ROBOT_ACCOUNT.CREATED_SUCCESS', { param: robot.name })
                .subscribe((res: string) => {
                    this.viewTokenComponent.createSuccess = res;
                });
            // export to token file
            const downLoadUrl = `data:text/json;charset=utf-8, ${encodeURIComponent(
                JSON.stringify(robot)
            )}`;
            this.viewTokenComponent.downLoadHref =
                this.sanitizer.bypassSecurityTrustUrl(downLoadUrl);
            this.viewTokenComponent.downLoadFileName = `${robot.name}.json`;
        }
        this.refresh();
    }
}
