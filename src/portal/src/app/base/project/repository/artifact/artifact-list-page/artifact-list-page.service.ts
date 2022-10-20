import { Injectable } from '@angular/core';
import { ClrLoadingState } from '@clr/angular';
import {
    ScanningResultService,
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../../../shared/services';
import { ErrorHandler } from '../../../../../shared/units/error-handler';

@Injectable()
export class ArtifactListPageService {
    private _scanBtnState: ClrLoadingState;
    private _hasEnabledScanner: boolean = false;
    private _hasAddLabelImagePermission: boolean = false;
    private _hasRetagImagePermission: boolean = false;
    private _hasDeleteImagePermission: boolean = false;
    private _hasScanImagePermission: boolean = false;

    constructor(
        private scanningService: ScanningResultService,
        private userPermissionService: UserPermissionService,
        private errorHandlerService: ErrorHandler
    ) {}

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
                },
                error => this.errorHandlerService.error(error)
            );
    }
}
