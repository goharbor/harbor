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
import { Component, EventEmitter, Output, ViewChild } from '@angular/core';
import { Scanner } from '../scanner';
import { NewScannerFormComponent } from '../new-scanner-form/new-scanner-form.component';
import { ClrLoadingState } from '@clr/angular';
import { finalize } from 'rxjs/operators';
import { MessageHandlerService } from '../../../../../shared/services/message-handler.service';
import { TranslateService } from '@ngx-translate/core';
import { InlineAlertComponent } from '../../../../../shared/components/inline-alert/inline-alert.component';
import { ScannerService } from '../../../../../../../ng-swagger-gen/services/scanner.service';
import { ScannerRegistrationReq } from '../../../../../../../ng-swagger-gen/models/scanner-registration-req';
import { clone } from '../../../../../shared/units/utils';

@Component({
    selector: 'new-scanner-modal',
    templateUrl: 'new-scanner-modal.component.html',
    styleUrls: ['../../../../../common.scss'],
})
export class NewScannerModalComponent {
    testMap: any = {};
    opened: boolean = false;
    @Output() notify = new EventEmitter<Scanner>();
    @ViewChild(NewScannerFormComponent, { static: true })
    newScannerFormComponent: NewScannerFormComponent;
    checkBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    onTesting: boolean = false;
    onSaving: boolean = false;
    isEdit: boolean = false;
    originValue: any;
    uid: string;
    editScanner: Scanner;
    @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
    constructor(
        private configScannerService: ScannerService,
        private msgHandler: MessageHandlerService,
        private translate: TranslateService
    ) {}
    open(): void {
        // reset
        this.opened = true;
        this.inlineAlert.close();
        this.testMap = {};
        this.newScannerFormComponent.showEndpointError = false;
        this.newScannerFormComponent.newScannerForm.reset({ auth: 'None' });
    }
    close(): void {
        this.opened = false;
    }
    create(): void {
        this.onSaving = true;
        this.saveBtnState = ClrLoadingState.LOADING;
        const scanner: ScannerRegistrationReq = { name: '', url: '' };
        const value = this.newScannerFormComponent.newScannerForm.value;
        scanner.name = value.name;
        scanner.description = value.description;
        scanner.url = value.url;
        if (value.auth === 'None') {
            scanner.auth = '';
        } else if (value.auth === 'Basic') {
            scanner.auth = value.auth;
            scanner.access_credential =
                value.accessCredential.username +
                ':' +
                value.accessCredential.password;
        } else if (value.auth === 'APIKey') {
            scanner.auth = value.auth;
            scanner.access_credential = value.accessCredential.apiKey;
        } else {
            scanner.auth = value.auth;
            scanner.access_credential = value.accessCredential.token;
        }
        scanner.skip_certVerify = !!value.skipCertVerify;
        scanner.use_internal_addr = !!value.useInner;
        this.configScannerService
            .createScanner({
                registration: scanner,
            })
            .pipe(finalize(() => (this.onSaving = false)))
            .subscribe(
                response => {
                    this.close();
                    this.msgHandler.showSuccess('SCANNER.ADD_SUCCESS');
                    this.notify.emit();
                    this.saveBtnState = ClrLoadingState.SUCCESS;
                },
                error => {
                    this.inlineAlert.showInlineError(error);
                    this.saveBtnState = ClrLoadingState.ERROR;
                }
            );
    }
    get hasPassedTest(): boolean {
        return this.testMap[
            this.newScannerFormComponent.newScannerForm.get('url').value
        ];
    }
    get canTestEndpoint(): boolean {
        if (
            this.newScannerFormComponent.newScannerForm.get('auth').value ===
            'Basic'
        ) {
            return (
                this.newScannerFormComponent.newScannerForm
                    .get('accessCredential')
                    .get('username').valid &&
                this.newScannerFormComponent.newScannerForm
                    .get('accessCredential')
                    .get('password').valid
            );
        }
        if (
            this.newScannerFormComponent.newScannerForm.get('auth').value ===
            'Bearer'
        ) {
            return this.newScannerFormComponent.newScannerForm
                .get('accessCredential')
                .get('token').valid;
        }
        if (
            this.newScannerFormComponent.newScannerForm.get('auth').value ===
            'APIKey'
        ) {
            return this.newScannerFormComponent.newScannerForm
                .get('accessCredential')
                .get('apiKey').valid;
        }
        return (
            !this.onTesting &&
            this.newScannerFormComponent &&
            !this.newScannerFormComponent.checkOnGoing &&
            this.newScannerFormComponent.newScannerForm.get('name').valid &&
            !this.newScannerFormComponent.checkEndpointOnGoing &&
            this.newScannerFormComponent.newScannerForm.get('url').valid
        );
    }
    get valid(): boolean {
        if (
            this.onSaving ||
            this.newScannerFormComponent.isNameExisting ||
            this.newScannerFormComponent.isEndpointUrlExisting ||
            this.onTesting ||
            !this.newScannerFormComponent ||
            this.newScannerFormComponent.checkOnGoing ||
            this.newScannerFormComponent.checkEndpointOnGoing
        ) {
            return false;
        }
        if (this.newScannerFormComponent.newScannerForm.get('name').invalid) {
            return false;
        }
        if (this.newScannerFormComponent.newScannerForm.get('url').invalid) {
            return false;
        }
        if (
            this.newScannerFormComponent.newScannerForm.get('auth').value ===
            'Basic'
        ) {
            return (
                this.newScannerFormComponent.newScannerForm
                    .get('accessCredential')
                    .get('username').valid &&
                this.newScannerFormComponent.newScannerForm
                    .get('accessCredential')
                    .get('password').valid
            );
        }
        if (
            this.newScannerFormComponent.newScannerForm.get('auth').value ===
            'Bearer'
        ) {
            return this.newScannerFormComponent.newScannerForm
                .get('accessCredential')
                .get('token').valid;
        }
        if (
            this.newScannerFormComponent.newScannerForm.get('auth').value ===
            'APIKey'
        ) {
            return this.newScannerFormComponent.newScannerForm
                .get('accessCredential')
                .get('apiKey').valid;
        }
        return true;
    }
    get validForSaving() {
        return this.valid && this.hasChange();
    }
    hasChange(): boolean {
        if (
            this.originValue.name !==
            this.newScannerFormComponent.newScannerForm.get('name').value
        ) {
            return true;
        }
        if (
            this.originValue.description !==
            this.newScannerFormComponent.newScannerForm.get('description').value
        ) {
            return true;
        }
        if (
            this.originValue.url !==
            this.newScannerFormComponent.newScannerForm.get('url').value
        ) {
            return true;
        }
        if (
            this.originValue.auth !==
            this.newScannerFormComponent.newScannerForm.get('auth').value
        ) {
            return true;
        }
        if (
            this.originValue.skipCertVerify !==
            this.newScannerFormComponent.newScannerForm.get('skipCertVerify')
                .value
        ) {
            return true;
        }
        if (
            this.originValue.useInner !==
            this.newScannerFormComponent.newScannerForm.get('useInner').value
        ) {
            return true;
        }
        if (this.originValue.auth === 'Basic') {
            if (
                this.originValue.accessCredential.username !==
                this.newScannerFormComponent.newScannerForm
                    .get('accessCredential')
                    .get('username').value
            ) {
                return true;
            }
            if (
                this.originValue.accessCredential.password !==
                this.newScannerFormComponent.newScannerForm
                    .get('accessCredential')
                    .get('password').value
            ) {
                return true;
            }
        }
        if (this.originValue.auth === 'Bearer') {
            if (
                this.originValue.accessCredential.token !==
                this.newScannerFormComponent.newScannerForm
                    .get('accessCredential')
                    .get('token').value
            ) {
                return true;
            }
        }
        if (this.originValue.auth === 'APIKey') {
            if (
                this.originValue.accessCredential.apiKey !==
                this.newScannerFormComponent.newScannerForm
                    .get('accessCredential')
                    .get('apiKey').value
            ) {
                return true;
            }
        }
        return false;
    }
    onTestEndpoint() {
        this.onTesting = true;
        this.checkBtnState = ClrLoadingState.LOADING;
        const scanner: ScannerRegistrationReq = { name: '', url: '' };
        const value = this.newScannerFormComponent.newScannerForm.value;
        scanner.name = value.name;
        scanner.description = value.description;
        scanner.url = value.url;
        if (value.auth === 'None') {
            scanner.auth = '';
        } else if (value.auth === 'Basic') {
            scanner.auth = value.auth;
            scanner.access_credential =
                value.accessCredential.username +
                ':' +
                value.accessCredential.password;
        } else if (value.auth === 'APIKey') {
            scanner.auth = value.auth;
            scanner.access_credential = value.accessCredential.apiKey;
        } else {
            scanner.auth = value.auth;
            scanner.access_credential = value.accessCredential.token;
        }
        scanner.skip_certVerify = !!value.skipCertVerify;
        scanner.use_internal_addr = !!value.useInner;
        this.configScannerService
            .pingScanner({
                settings: scanner,
            })
            .pipe(finalize(() => (this.onTesting = false)))
            .subscribe(
                response => {
                    this.inlineAlert.showInlineSuccess({
                        message: 'SCANNER.TEST_PASS',
                    });
                    this.checkBtnState = ClrLoadingState.SUCCESS;
                    this.testMap[
                        this.newScannerFormComponent.newScannerForm.get(
                            'url'
                        ).value
                    ] = true;
                },
                error => {
                    this.translate
                        .get('SCANNER.TEST_FAILED', {
                            name: this.newScannerFormComponent.newScannerForm.get(
                                'name'
                            ).value,
                            url: this.newScannerFormComponent.newScannerForm.get(
                                'url'
                            ).value,
                        })
                        .subscribe((res: string) => {
                            this.inlineAlert.showInlineError(res);
                        });
                    this.checkBtnState = ClrLoadingState.ERROR;
                }
            );
    }
    save() {
        this.onSaving = true;
        this.saveBtnState = ClrLoadingState.LOADING;
        let value = this.newScannerFormComponent.newScannerForm.value;
        this.editScanner.name = value.name;
        this.editScanner.description = value.description;
        this.editScanner.url = value.url;
        if (value.auth === 'None') {
            this.editScanner.auth = '';
        } else if (value.auth === 'Basic') {
            this.editScanner.auth = value.auth;
            this.editScanner.access_credential =
                value.accessCredential.username +
                ':' +
                value.accessCredential.password;
        } else if (value.auth === 'APIKey') {
            this.editScanner.auth = value.auth;
            this.editScanner.access_credential = value.accessCredential.apiKey;
        } else {
            this.editScanner.auth = value.auth;
            this.editScanner.access_credential = value.accessCredential.token;
        }
        this.editScanner.skip_certVerify = !!value.skipCertVerify;
        this.editScanner.use_internal_addr = !!value.useInner;
        this.editScanner.uuid = this.uid;
        const scanner: ScannerRegistrationReq = clone(this.editScanner);
        this.configScannerService
            .updateScanner({
                registrationId: this.editScanner.uuid,
                registration: scanner,
            })
            .pipe(finalize(() => (this.onSaving = false)))
            .subscribe(
                response => {
                    this.close();
                    this.msgHandler.showSuccess('SCANNER.UPDATE_SUCCESS');
                    this.notify.emit();
                    this.saveBtnState = ClrLoadingState.SUCCESS;
                },
                error => {
                    this.inlineAlert.showInlineError(error);
                    this.saveBtnState = ClrLoadingState.ERROR;
                }
            );
    }
}
