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
import { NewRobotComponent } from './new-robot/new-robot.component';
import { ViewTokenComponent } from '../../../shared/components/view-token/view-token.component';
import { RobotService } from '../../../../../ng-swagger-gen/services/robot.service';
import { Robot } from '../../../../../ng-swagger-gen/models/robot';
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
    FrontRobot,
    getSystemAccess,
    NAMESPACE_ALL_PROJECTS,
    NEW_EMPTY_ROBOT,
    PermissionsKinds,
} from './system-robot-util';
import { ProjectsModalComponent } from './projects-modal/projects-modal.component';
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
} from '../../../shared/entities/shared.const';
import { errorHandler } from '../../../shared/units/shared.utils';
import { ConfirmationMessage } from '../../global-confirmation-dialog/confirmation-message';
import { RobotPermission } from '../../../../../ng-swagger-gen/models/robot-permission';
import { SysteminfoService } from '../../../../../ng-swagger-gen/services/systeminfo.service';
import { Access } from '../../../../../ng-swagger-gen/models/access';
import { PermissionSelectPanelModes } from '../../../shared/components/robot-permissions-panel/robot-permissions-panel.component';
import { PermissionsService } from '../../../../../ng-swagger-gen/services/permissions.service';
import { Permissions } from '../../../../../ng-swagger-gen/models/permissions';

@Component({
    selector: 'system-robot-accounts',
    templateUrl: './system-robot-accounts.component.html',
    styleUrls: ['./system-robot-accounts.component.scss'],
})
export class SystemRobotAccountsComponent implements OnInit, OnDestroy {
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.SYSTEM_ROBOT_COMPONENT
    );
    currentPage: number = 1;
    total: number = 0;
    robots: FrontRobot[] = [];
    selectedRows: FrontRobot[] = [];
    loading: boolean = true;
    loadingData: boolean = false;
    addBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    @ViewChild(NewRobotComponent, { static: true })
    newRobotComponent: NewRobotComponent;
    @ViewChild(ViewTokenComponent)
    viewTokenComponent: ViewTokenComponent;
    @ViewChild(ProjectsModalComponent)
    projectsModalComponent: ProjectsModalComponent;
    @ViewChild(FilterComponent, { static: true })
    filterComponent: FilterComponent;
    searchSub: Subscription;
    searchKey: string;
    subscription: Subscription;
    deltaTime: number; // the different between server time and local time

    robotMetadata: Permissions;
    loadingMetadata: boolean = false;
    constructor(
        private robotService: RobotService,
        private msgHandler: MessageHandlerService,
        private operateDialogService: ConfirmationDialogService,
        private operationService: OperationService,
        private sanitizer: DomSanitizer,
        private translate: TranslateService,
        private systemInfoService: SysteminfoService,
        private permissionService: PermissionsService
    ) {}
    ngOnInit() {
        this.getRobotPermissions();
        this.getCurrenTime();
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
                                `name=~${this.searchKey}`
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
                        this.calculateProjects();
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
                            message.source === ConfirmationTargets.ROBOT_ACCOUNT
                        ) {
                            this.deleteRobots(message.data);
                        }
                        if (
                            message.state === ConfirmationState.CONFIRMED &&
                            message.source ===
                                ConfirmationTargets.ROBOT_ACCOUNT_ENABLE_OR_DISABLE
                        ) {
                            this.operateRobot();
                        }
                    }
                );
        }
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

    getRobotPermissions() {
        this.loadingData = true;
        this.permissionService
            .getPermissions()
            .pipe(finalize(() => (this.loadingData = false)))
            .subscribe(res => {
                this.robotMetadata = res;
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
        const queryParam: RobotService.ListRobotParams = {
            page: this.currentPage,
            pageSize: this.pageSize,
            sort: getSortingString(state),
        };
        if (this.searchKey) {
            queryParam.q = encodeURIComponent(`name=~${this.searchKey}`);
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
                    this.calculateProjects();
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
    calculateProjects() {
        if (this.robots && this.robots.length) {
            for (let i = 0; i < this.robots.length; i++) {
                if (
                    this.robots[i] &&
                    this.robots[i].permissions &&
                    this.robots[i].permissions.length
                ) {
                    for (
                        let j = 0;
                        j < this.robots[i].permissions.length;
                        j++
                    ) {
                        if (
                            this.robots[i].permissions[j].kind ===
                                PermissionsKinds.PROJECT &&
                            this.robots[i].permissions[j].namespace ===
                                NAMESPACE_ALL_PROJECTS
                        ) {
                            this.robots[i].permissionScope = {
                                coverAll: true,
                                access: this.robots[i].permissions[j].access,
                            };
                            break;
                        }
                    }
                }
            }
        }
    }
    getProjects(r: Robot): RobotPermission[] {
        const arr = [];
        if (r && r.permissions && r.permissions.length) {
            for (let i = 0; i < r.permissions.length; i++) {
                if (r.permissions[i].kind === PermissionsKinds.PROJECT) {
                    arr.push(r.permissions[i]);
                }
            }
        }
        return arr;
    }
    openProjectModal(permissions: RobotPermission[], robotName: string) {
        this.projectsModalComponent.projectsModalOpened = true;
        this.projectsModalComponent.robotName = robotName;
        this.projectsModalComponent.permissions = permissions;
        this.projectsModalComponent.clrDgRefresh();
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
            ConfirmationTargets.ROBOT_ACCOUNT,
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
            ConfirmationTargets.ROBOT_ACCOUNT_ENABLE_OR_DISABLE,
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

    getSystemAccess(r: Robot): Access[] {
        return getSystemAccess(r);
    }

    protected readonly NEW_EMPTY_ROBOT = NEW_EMPTY_ROBOT;
    protected readonly PermissionSelectPanelModes = PermissionSelectPanelModes;
}
