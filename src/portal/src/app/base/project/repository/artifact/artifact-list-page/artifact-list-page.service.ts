import { Injectable } from '@angular/core';
import { ClrLoadingState } from '@clr/angular';
import {
    ScanningResultService,
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../../../shared/services';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { Scanner } from '../../../../left-side-nav/interrogation-services/scanner/scanner';

@Injectable()
export class ArtifactListPageService {
    private _scanBtnState: ClrLoadingState;
    private _sbomBtnState: ClrLoadingState;
    private _hasEnabledScanner: boolean = false;
    private _hasScannerSupportVulnerability: boolean = false;
    private _hasScannerSupportSBOM: boolean = false;
    private _hasAddLabelImagePermission: boolean = false;
    private _hasRetagImagePermission: boolean = false;
    private _hasDeleteImagePermission: boolean = false;
    private _hasScanImagePermission: boolean = false;
    private _hasSbomPermission: boolean = false;
    private _scanner: Scanner = undefined;

    constructor(
        private scanningService: ScanningResultService,
        private userPermissionService: UserPermissionService,
        private errorHandlerService: ErrorHandler
    ) {}

    getProjectScanner(): Scanner {
        return this._scanner;
    }

    getScanBtnState(): ClrLoadingState {
        return this._scanBtnState;
    }

    getSbomBtnState(): ClrLoadingState {
        return this._sbomBtnState;
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

    hasSbomPermission(): boolean {
        return this._hasSbomPermission;
    }

    hasScannerSupportVulnerability(): boolean {
        return this._hasScannerSupportVulnerability;
    }

    hasScannerSupportSBOM(): boolean {
        return this._hasScannerSupportSBOM;
    }

    init(projectId: number) {
        this._getProjectScanner(projectId);
        this._getPermissionRule(projectId);
    }

    updateStates(
        enabledScanner: boolean,
        scanState?: ClrLoadingState,
        sbomState?: ClrLoadingState
    ) {
        if (scanState) {
            this._scanBtnState = scanState;
        }
        if (sbomState) {
            this._sbomBtnState = sbomState;
        }
        this._hasEnabledScanner = enabledScanner;
    }

    updateCapabilities(capabilities?: any) {
        if (capabilities) {
            if (capabilities?.support_vulnerability !== undefined) {
                this._hasScannerSupportVulnerability =
                    capabilities.support_vulnerability;
            }
            if (capabilities?.support_sbom !== undefined) {
                this._hasScannerSupportSBOM = capabilities.support_sbom;
            }
        }
    }

    private _getProjectScanner(projectId: number): void {
        this._hasEnabledScanner = false;
        this._scanBtnState = ClrLoadingState.LOADING;
        this._sbomBtnState = ClrLoadingState.LOADING;
        this.scanningService.getProjectScanner(projectId).subscribe(
            response => {
                if (response && '{}' !== JSON.stringify(response)) {
                    this._scanner = response;
                    if (!response.disabled && response.health === 'healthy') {
                        this.updateStates(
                            true,
                            ClrLoadingState.SUCCESS,
                            ClrLoadingState.SUCCESS
                        );
                        if (response?.capabilities) {
                            this.updateCapabilities(response?.capabilities);
                        }
                    } else {
                        this.updateStates(
                            false,
                            ClrLoadingState.ERROR,
                            ClrLoadingState.ERROR
                        );
                    }
                } else {
                    this.updateStates(
                        false,
                        ClrLoadingState.ERROR,
                        ClrLoadingState.ERROR
                    );
                }
            },
            error => {
                this._scanner = null;
                this.updateStates(
                    false,
                    ClrLoadingState.ERROR,
                    ClrLoadingState.ERROR
                );
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
            {
                resource: USERSTATICPERMISSION.REPOSITORY_TAG_SBOM_JOB.KEY,
                action: USERSTATICPERMISSION.REPOSITORY_TAG_SBOM_JOB.VALUE
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
                    this._hasSbomPermission = results?.[4] ?? false;
                    // TODO need to remove the static code
                    this._hasSbomPermission = true;
                },
                error => this.errorHandlerService.error(error)
            );
    }
}
