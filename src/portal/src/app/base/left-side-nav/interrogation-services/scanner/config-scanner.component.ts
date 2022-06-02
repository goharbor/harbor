import { Component, ViewChild, OnInit, OnDestroy } from '@angular/core';
import { Scanner, SCANNERS_DOC } from './scanner';
import { NewScannerModalComponent } from './new-scanner-modal/new-scanner-modal.component';
import { finalize } from 'rxjs/operators';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import {
    clone,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import { ConfirmationDialogService } from '../../../global-confirmation-dialog/confirmation-dialog.service';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../shared/entities/shared.const';
import { ConfirmationMessage } from '../../../global-confirmation-dialog/confirmation-message';
import { ScannerService } from '../../../../../../ng-swagger-gen/services/scanner.service';
import { ClrDatagridStateInterface } from '@clr/angular';
import { ScannerRegistrationReq } from '../../../../../../ng-swagger-gen/models/scanner-registration-req';

@Component({
    selector: 'config-scanner',
    templateUrl: 'config-scanner.component.html',
    styleUrls: [
        './config-scanner.component.scss',
        '../../config/config.component.scss',
    ],
})
export class ConfigurationScannerComponent implements OnInit, OnDestroy {
    scanners: Scanner[] = [];
    selectedRow: Scanner;
    onGoing: boolean = true;
    @ViewChild(NewScannerModalComponent)
    newScannerDialog: NewScannerModalComponent;
    deletionSubscription: any;
    scannerDocUrl: string = SCANNERS_DOC;
    page: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.SYSTEM_SCANNER_COMPONENT
    );
    total: number = 0;
    state: ClrDatagridStateInterface;
    constructor(
        private configScannerService: ScannerService,
        private errorHandler: ErrorHandler,
        private msgHandler: MessageHandlerService,
        private deletionDialogService: ConfirmationDialogService
    ) {}
    ngOnInit() {
        if (!this.deletionSubscription) {
            this.deletionSubscription =
                this.deletionDialogService.confirmationConfirm$.subscribe(
                    confirmed => {
                        if (
                            confirmed &&
                            confirmed.source === ConfirmationTargets.SCANNER &&
                            confirmed.state === ConfirmationState.CONFIRMED
                        ) {
                            this.configScannerService
                                .deleteScanner({
                                    registrationId: confirmed.data[0].uuid,
                                })
                                .subscribe(
                                    response => {
                                        this.msgHandler.showSuccess(
                                            'SCANNER.DELETE_SUCCESS'
                                        );
                                        this.refresh();
                                    },
                                    error => {
                                        this.errorHandler.error(error);
                                    }
                                );
                        }
                    }
                );
        }
    }
    ngOnDestroy(): void {
        if (this.deletionSubscription) {
            this.deletionSubscription.unsubscribe();
            this.deletionSubscription = null;
        }
    }
    refresh() {
        this.page = 1;
        this.selectedRow = null;
        this.total = 0;
        this.getScanners(this.state);
    }
    getScanners(state?: ClrDatagridStateInterface) {
        this.state = state;
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.SYSTEM_SCANNER_COMPONENT,
                this.pageSize
            );
        }
        let q: string;
        if (state && state.filters && state.filters.length) {
            q = encodeURIComponent(
                `${state.filters[0].property}=~${state.filters[0].value}`
            );
        }
        let sort: string;
        if (state && state.sort && state.sort.by) {
            sort = getSortingString(state);
        } else {
            // sort by creation_time desc by default
            sort = `-creation_time`;
        }
        this.onGoing = true;
        this.configScannerService
            .listScannersResponse({
                page: this.page,
                pageSize: this.pageSize,
                q: q,
                sort: sort,
            })
            .pipe(finalize(() => (this.onGoing = false)))
            .subscribe(
                response => {
                    // Get total count
                    if (response.headers) {
                        let xHeader: string =
                            response.headers.get('X-Total-Count');
                        if (xHeader) {
                            this.total = parseInt(xHeader, 0);
                        }
                    }
                    this.scanners = response.body || [];
                    this.getMetadataForAll();
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }
    getMetadataForAll() {
        if (this.scanners && this.scanners.length > 0) {
            this.scanners.forEach((scanner, index) => {
                if (scanner.uuid) {
                    this.scanners[index].loadingMetadata = true;
                    this.configScannerService
                        .getScannerMetadata({
                            registrationId: scanner.uuid,
                        })
                        .pipe(
                            finalize(
                                () =>
                                    (this.scanners[index].loadingMetadata =
                                        false)
                            )
                        )
                        .subscribe(
                            response => {
                                this.scanners[index].metadata = response;
                            },
                            error => {
                                this.scanners[index].metadata = null;
                            }
                        );
                }
            });
        }
    }

    addNewScanner(): void {
        this.newScannerDialog.open();
        this.newScannerDialog.isEdit = false;
        this.newScannerDialog.newScannerFormComponent.isEdit = false;
    }
    addSuccess() {
        this.getScanners();
    }
    changeStat() {
        if (this.selectedRow) {
            let scanner: ScannerRegistrationReq = clone(this.selectedRow);
            scanner.disabled = !scanner.disabled;
            this.configScannerService
                .updateScanner({
                    registrationId: this.selectedRow.uuid,
                    registration: scanner,
                })
                .subscribe(
                    response => {
                        this.msgHandler.showSuccess('SCANNER.UPDATE_SUCCESS');
                        this.refresh();
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }
    setAsDefault() {
        if (this.selectedRow) {
            this.configScannerService
                .setScannerAsDefault({
                    registrationId: this.selectedRow.uuid,
                    payload: {
                        is_default: true,
                    },
                })
                .subscribe(
                    response => {
                        this.msgHandler.showSuccess('SCANNER.UPDATE_SUCCESS');
                        this.refresh();
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }
    deleteScanners() {
        if (this.selectedRow) {
            // Confirm deletion
            let msg: ConfirmationMessage = new ConfirmationMessage(
                'SCANNER.CONFIRM_DELETION',
                'SCANNER.DELETION_SUMMARY',
                this.selectedRow.name,
                [this.selectedRow],
                ConfirmationTargets.SCANNER,
                ConfirmationButtons.DELETE_CANCEL
            );
            this.deletionDialogService.openComfirmDialog(msg);
        }
    }
    editScanner() {
        if (this.selectedRow) {
            this.newScannerDialog.open();
            let resetValue: object = {};
            resetValue['name'] = this.selectedRow.name;
            resetValue['description'] = this.selectedRow.description;
            resetValue['url'] = this.selectedRow.url;
            resetValue['skipCertVerify'] = this.selectedRow.skip_certVerify;
            resetValue['useInner'] = this.selectedRow.use_internal_addr;
            if (this.selectedRow.auth === 'Basic') {
                resetValue['auth'] = 'Basic';
                let username: string =
                    this.selectedRow.access_credential.split(':')[0];
                let password: string =
                    this.selectedRow.access_credential.split(':')[1];
                resetValue['accessCredential'] = {
                    username: username,
                    password: password,
                };
            } else if (this.selectedRow.auth === 'Bearer') {
                resetValue['auth'] = 'Bearer';
                resetValue['accessCredential'] = {
                    token: this.selectedRow.access_credential,
                };
            } else if (this.selectedRow.auth === 'APIKey') {
                resetValue['auth'] = 'APIKey';
                resetValue['accessCredential'] = {
                    apiKey: this.selectedRow.access_credential,
                };
            } else {
                resetValue['auth'] = 'None';
            }
            this.newScannerDialog.newScannerFormComponent.newScannerForm.reset(
                resetValue
            );
            this.newScannerDialog.isEdit = true;
            this.newScannerDialog.newScannerFormComponent.isEdit = true;
            this.newScannerDialog.uid = this.selectedRow.uuid;
            this.newScannerDialog.originValue = clone(resetValue);
            this.newScannerDialog.newScannerFormComponent.originValue =
                clone(resetValue);
            this.newScannerDialog.editScanner = clone(this.selectedRow);
        }
    }
}
