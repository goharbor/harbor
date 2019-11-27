import { Component, ViewChild, OnInit, OnDestroy } from "@angular/core";
import { Scanner } from "./scanner";
import { NewScannerModalComponent } from "./new-scanner-modal/new-scanner-modal.component";
import { ConfigScannerService } from "./config-scanner.service";
import { clone, ErrorHandler } from "@harbor/ui";
import { finalize } from "rxjs/operators";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { ConfirmationButtons, ConfirmationState, ConfirmationTargets } from "../../shared/shared.const";
import { ConfirmationDialogService } from "../../shared/confirmation-dialog/confirmation-dialog.service";
import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';

@Component({
    selector: 'config-scanner',
    templateUrl: "config-scanner.component.html",
    styleUrls: ['./config-scanner.component.scss', '../config.component.scss']
})
export class ConfigurationScannerComponent implements OnInit, OnDestroy {
    scanners: Scanner[] = [];
    selectedRow: Scanner;
    onGoing: boolean = false;
    @ViewChild(NewScannerModalComponent, {static: false})
    newScannerDialog: NewScannerModalComponent;
    deletionSubscription: any;
    constructor(
        private configScannerService: ConfigScannerService,
        private errorHandler: ErrorHandler,
        private msgHandler: MessageHandlerService,
        private deletionDialogService: ConfirmationDialogService,
    ) {}
    ngOnInit() {
        if (!this.deletionSubscription) {
            this.deletionSubscription = this.deletionDialogService.confirmationConfirm$.subscribe(confirmed => {
                if (confirmed &&
                    confirmed.source === ConfirmationTargets.SCANNER &&
                    confirmed.state === ConfirmationState.CONFIRMED) {
                    this.configScannerService.deleteScanners(confirmed.data)
                        .subscribe(response => {
                            this.msgHandler.showSuccess("SCANNER.DELETE_SUCCESS");
                            this.getScanners();
                        }, error => {
                            this.errorHandler.error(error);
                        });
                }
            });
        }
        this.getScanners();
    }
    ngOnDestroy(): void {
        if (this.deletionSubscription) {
            this.deletionSubscription.unsubscribe();
            this.deletionSubscription = null;
        }
    }
    getScanners() {
        this.onGoing = true;
        this.configScannerService.getScanners()
            .pipe(finalize(() => this.onGoing = false))
            .subscribe(response => {
            this.scanners = response;
            this.getMetadataForAll();
        }, error => {
            this.errorHandler.error(error);
        });
    }
    getMetadataForAll() {
        if (this.scanners && this.scanners.length > 0) {
            this.scanners.forEach((scanner, index) => {
                if (scanner.uuid ) {
                    this.scanners[index].loadingMetadata = true;
                    this.configScannerService.getScannerMetadata(scanner.uuid)
                        .pipe(finalize(() => this.scanners[index].loadingMetadata = false))
                        .subscribe(response => {
                            this.scanners[index].metadata = response;
                        }, error => {
                            this.scanners[index].metadata = null;
                        });
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
            let scanner: Scanner = clone(this.selectedRow);
            scanner.disabled = !scanner.disabled;
            this.configScannerService.updateScanner(scanner)
                .subscribe(response => {
                    this.msgHandler.showSuccess("SCANNER.UPDATE_SUCCESS");
                    this.getScanners();
                }, error => {
                    this.errorHandler.error(error);
                });
        }
    }
    setAsDefault() {
        if (this.selectedRow) {
            this.configScannerService.setAsDefault(this.selectedRow.uuid)
                .subscribe(response => {
                    this.msgHandler.showSuccess("SCANNER.UPDATE_SUCCESS");
                    this.getScanners();
                }, error => {
                    this.errorHandler.error(error);
                });
        }
    }
    deleteScanners() {
        if (this.selectedRow) {
            // Confirm deletion
            let msg: ConfirmationMessage = new ConfirmationMessage(
                "SCANNER.CONFIRM_DELETION",
                "SCANNER.DELETION_SUMMARY",
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
                let username: string = this.selectedRow.access_credential.split(":")[0];
                let password: string = this.selectedRow.access_credential.split(":")[1];
                resetValue['accessCredential'] = {
                    username: username,
                    password: password
                };
            } else if (this.selectedRow.auth === 'Bearer') {
                resetValue['auth'] = 'Bearer';
                resetValue['accessCredential'] = {
                   token: this.selectedRow.access_credential
                };
            } else if (this.selectedRow.auth === 'APIKey') {
                resetValue['auth'] = 'APIKey';
                resetValue['accessCredential'] = {
                    apiKey: this.selectedRow.access_credential
                };
            } else {
                resetValue['auth'] = 'None';
            }
            this.newScannerDialog.newScannerFormComponent.newScannerForm.reset(resetValue);
            this.newScannerDialog.isEdit = true;
            this.newScannerDialog.newScannerFormComponent.isEdit = true;
            this.newScannerDialog.uid = this.selectedRow.uuid;
            this.newScannerDialog.originValue = clone(resetValue);
            this.newScannerDialog.newScannerFormComponent.originValue = clone(resetValue);
            this.newScannerDialog.editScanner = clone(this.selectedRow);
        }
    }
}
