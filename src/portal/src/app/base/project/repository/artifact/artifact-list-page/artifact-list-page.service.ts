import { Injectable } from '@angular/core';
import { ClrLoadingState } from '@clr/angular';
import {
    ScanningResultService,
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../../../shared/services';
import { LabelState } from './artifact-list/artifact-list-tab/artifact-list-tab.component';
import { forkJoin, Observable } from 'rxjs';
import { LabelService } from 'ng-swagger-gen/services/label.service';
import { Label } from 'ng-swagger-gen/models/label';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { clone } from '../../../../../shared/units/utils';

const PAGE_SIZE: number = 100;

@Injectable()
export class ArtifactListPageService {
    private _scanBtnState: ClrLoadingState;
    private _allLabels: LabelState[] = [];
    imageStickLabels: LabelState[] = [];
    imageFilterLabels: LabelState[] = [];
    private _hasEnabledScanner: boolean = false;
    private _hasAddLabelImagePermission: boolean = false;
    private _hasRetagImagePermission: boolean = false;
    private _hasDeleteImagePermission: boolean = false;
    private _hasScanImagePermission: boolean = false;

    constructor(
        private scanningService: ScanningResultService,
        private labelService: LabelService,
        private userPermissionService: UserPermissionService,
        private errorHandlerService: ErrorHandler
    ) {}
    resetClonedLabels() {
        this.imageStickLabels = clone(this._allLabels);
        this.imageFilterLabels = clone(this._allLabels);
    }
    getScanBtnState(): ClrLoadingState {
        return this._scanBtnState;
    }

    hasEnabledScanner(): boolean {
        return this._hasEnabledScanner;
    }

    hasAddLabelImagePermission(): boolean {
        return this._hasAddLabelImagePermission;
    }

    hasRetagImagePermission(): boolean {
        return this._hasRetagImagePermission;
    }

    hasDeleteImagePermission(): boolean {
        return this._hasDeleteImagePermission;
    }

    hasScanImagePermission(): boolean {
        return this._hasScanImagePermission;
    }

    init(projectId: number) {
        this._getProjectScanner(projectId);
        this._getPermissionRule(projectId);
    }

    private _getProjectScanner(projectId: number): void {
        this._hasEnabledScanner = false;
        this._scanBtnState = ClrLoadingState.LOADING;
        this.scanningService.getProjectScanner(projectId).subscribe(
            response => {
                if (
                    response &&
                    '{}' !== JSON.stringify(response) &&
                    !response.disabled &&
                    response.health === 'healthy'
                ) {
                    this._scanBtnState = ClrLoadingState.SUCCESS;
                    this._hasEnabledScanner = true;
                } else {
                    this._scanBtnState = ClrLoadingState.ERROR;
                }
            },
            error => {
                this._scanBtnState = ClrLoadingState.ERROR;
            }
        );
    }

    private _getAllLabels(projectId: number): void {
        // get all project labels
        this._allLabels = []; // reset
        this.labelService
            .ListLabelsResponse({
                pageSize: PAGE_SIZE,
                page: 1,
                scope: 'p',
                projectId: projectId,
            })
            .subscribe(res => {
                if (res.headers) {
                    const xHeader: string = res.headers.get('X-Total-Count');
                    const totalCount = parseInt(xHeader, 0);
                    let arr = res.body || [];
                    if (totalCount <= PAGE_SIZE) {
                        // already gotten all project labels
                        if (arr && arr.length) {
                            arr.forEach(data => {
                                this._allLabels.push({
                                    iconsShow: false,
                                    label: data,
                                    show: true,
                                });
                            });
                            this.resetClonedLabels();
                        }
                    } else {
                        // get all the project labels in specified times
                        const times: number = Math.ceil(totalCount / PAGE_SIZE);
                        const observableList: Observable<Label[]>[] = [];
                        for (let i = 2; i <= times; i++) {
                            observableList.push(
                                this.labelService.ListLabels({
                                    page: i,
                                    pageSize: PAGE_SIZE,
                                    scope: 'p',
                                    projectId: projectId,
                                })
                            );
                        }
                        this._handleLabelRes(observableList, arr);
                    }
                }
            });
        // get all global labels
        this.labelService
            .ListLabelsResponse({
                pageSize: PAGE_SIZE,
                page: 1,
                scope: 'g',
            })
            .subscribe(res => {
                if (res.headers) {
                    const xHeader: string = res.headers.get('X-Total-Count');
                    const totalCount = parseInt(xHeader, 0);
                    let arr = res.body || [];
                    if (totalCount <= PAGE_SIZE) {
                        // already gotten all global labels
                        if (arr && arr.length) {
                            arr.forEach(data => {
                                this._allLabels.push({
                                    iconsShow: false,
                                    label: data,
                                    show: true,
                                });
                            });
                            this.resetClonedLabels();
                        }
                    } else {
                        // get all the global labels in specified times
                        const times: number = Math.ceil(totalCount / PAGE_SIZE);
                        const observableList: Observable<Label[]>[] = [];
                        for (let i = 2; i <= times; i++) {
                            observableList.push(
                                this.labelService.ListLabels({
                                    page: i,
                                    pageSize: PAGE_SIZE,
                                    scope: 'g',
                                })
                            );
                        }
                        this._handleLabelRes(observableList, arr);
                    }
                }
            });
    }

    private _handleLabelRes(
        observableList: Observable<Label[]>[],
        arr: Label[]
    ) {
        forkJoin(observableList).subscribe(response => {
            if (response && response.length) {
                response.forEach(item => {
                    arr = arr.concat(item);
                });
                arr.forEach(data => {
                    this._allLabels.push({
                        iconsShow: false,
                        label: data,
                        show: true,
                    });
                });
                this.resetClonedLabels();
            }
        });
    }

    private _getPermissionRule(projectId: number): void {
        const permissions = [
            {
                resource: USERSTATICPERMISSION.REPOSITORY_ARTIFACT_LABEL.KEY,
                action: USERSTATICPERMISSION.REPOSITORY_ARTIFACT_LABEL.VALUE
                    .CREATE,
            },
            {
                resource: USERSTATICPERMISSION.REPOSITORY.KEY,
                action: USERSTATICPERMISSION.REPOSITORY.VALUE.PULL,
            },
            {
                resource: USERSTATICPERMISSION.ARTIFACT.KEY,
                action: USERSTATICPERMISSION.ARTIFACT.VALUE.DELETE,
            },
            {
                resource: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.KEY,
                action: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.VALUE
                    .CREATE,
            },
        ];
        this.userPermissionService
            .hasProjectPermissions(projectId, permissions)
            .subscribe(
                (results: Array<boolean>) => {
                    this._hasAddLabelImagePermission = results[0];
                    this._hasRetagImagePermission = results[1];
                    this._hasDeleteImagePermission = results[2];
                    this._hasScanImagePermission = results[3];
                    // only has label permission
                    if (this._hasAddLabelImagePermission) {
                        this._getAllLabels(projectId);
                    }
                },
                error => this.errorHandlerService.error(error)
            );
    }
}
