import { Component, ViewChild, OnInit, OnDestroy } from "@angular/core";
import {Scanner} from "./scanner";
import {NewScannerModalComponent} from "./new-scanner-modal/new-scanner-modal.component";
import {ConfigScannerService} from "./config-scanner.service";
import { clone, ErrorHandler } from "@harbor/ui";
import { finalize } from "rxjs/operators";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { ConfirmationButtons, ConfirmationState, ConfirmationTargets } from "../../shared/shared.const";
import { ConfirmationDialogService } from "../../shared/confirmation-dialog/confirmation-dialog.service";
import { ConfirmationMessage } from '../../shared/confirmation-dialog/confirmation-message';
import { ScannerMetadata } from "./scanner-metadata";

@Component({
    selector: 'config-scanner',
    templateUrl: "config-scanner.component.html",
    styleUrls: ['./config-scanner.component.scss', '../config.component.scss']
})
export class ConfigurationScannerComponent implements OnInit, OnDestroy {
    scanners: Scanner[] = [];
    selectedRow: Scanner[] = [];
    onGoing: boolean = false;
    @ViewChild(NewScannerModalComponent, {static: false})
    newScannerDialog: NewScannerModalComponent;
    deletionSubscription: any;
    constructor(
        private configScannerService: ConfigScannerService,
        private scannerService: ConfigScannerService,
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
                            this.msgHandler.showSuccess("Delete Success");
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
        this.scannerService.getScanners()
            .pipe(finalize(() => this.onGoing = false))
            .subscribe(response => {
            this.scanners = response;
        }, error => {
            this.errorHandler.error(error);
        });
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
        if (this.selectedRow && this.selectedRow.length === 1) {
            let scanner: Scanner = clone(this.selectedRow[0]);
            scanner.disabled = !scanner.disabled;
            this.configScannerService.updateScanner(scanner)
                .subscribe(response => {
                    this.msgHandler.showSuccess("Update Success");
                    this.getScanners();
                }, error => {
                    this.errorHandler.error(error);
                });
        }
    }
    setAsDefault() {
        if (this.selectedRow && this.selectedRow.length === 1) {
            let scanner: Scanner = clone(this.selectedRow[0]);
            scanner.isDefault = true;
            this.configScannerService.updateScanner(scanner)
                .subscribe(response => {
                    this.msgHandler.showSuccess("Update Success");
                    this.getScanners();
                }, error => {
                    this.errorHandler.error(error);
                });
        }
    }
    deleteScanners() {
        if (this.selectedRow && this.selectedRow.length > 0) {
            let endPoints: string[] = [];
            this.selectedRow.forEach(s => {
                endPoints.push(s.url);
            });
            // Confirm deletion
            let msg: ConfirmationMessage = new ConfirmationMessage(
                "USER.DELETION_TITLE",
                "USER.DELETION_SUMMARY",
                endPoints.join(','),
                this.selectedRow,
                ConfirmationTargets.SCANNER,
                ConfirmationButtons.DELETE_CANCEL
            );
            this.deletionDialogService.openComfirmDialog(msg);
        }
    }
    editScanner() {
        if (this.selectedRow && this.selectedRow.length === 1) {
            this.newScannerDialog.open();
            let resetValue: object = {};
            resetValue['name'] = this.selectedRow[0].name;
            resetValue['description'] = this.selectedRow[0].description;
            resetValue['url'] = this.selectedRow[0].url;
            resetValue['skipCertVerify'] = this.selectedRow[0].skipCertVerify;
            if (this.selectedRow[0].auth === 'Basic') {
                resetValue['auth'] = 'Basic';
                let username: string = this.selectedRow[0].accessCredential.split(":")[0];
                let password: string = this.selectedRow[0].accessCredential.split(":")[1];
                resetValue['accessCredential'] = {
                    username: username,
                    password: password
                };
            } else if (this.selectedRow[0].auth === 'Bearer') {
                resetValue['auth'] = 'Bearer';
                resetValue['accessCredential'] = {
                   token: this.selectedRow[0].accessCredential
                };
            } else if (this.selectedRow[0].auth === 'APIKey') {
                resetValue['auth'] = 'APIKey';
                resetValue['accessCredential'] = {
                    apiKey: this.selectedRow[0].accessCredential
                };
            } else {
                resetValue['auth'] = 'None';
            }
            this.newScannerDialog.newScannerFormComponent.newScannerForm.reset(resetValue);
            this.newScannerDialog.isEdit = true;
            this.newScannerDialog.newScannerFormComponent.isEdit = true;
            this.newScannerDialog.uid = this.selectedRow[0].uid;
            this.newScannerDialog.originValue = clone(resetValue);
            this.newScannerDialog.newScannerFormComponent.originValue = clone(resetValue);
        }
    }
}
