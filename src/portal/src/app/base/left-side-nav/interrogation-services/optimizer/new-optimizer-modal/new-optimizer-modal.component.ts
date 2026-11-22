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
import { Optimizer } from '../optimizer';
import { NewOptimizerFormComponent } from '../new-optimizer-form/new-optimizer-form.component';
import { ClrLoadingState } from '@clr/angular';
import { finalize } from 'rxjs/operators';
import { MessageHandlerService } from '../../../../../shared/services/message-handler.service';
import { TranslateService } from '@ngx-translate/core';
import { InlineAlertComponent } from '../../../../../shared/components/inline-alert/inline-alert.component';
import { OptimizerService } from '../../../../../../../ng-swagger-gen/services/optimizer.service';
import { OptimizerRegistrationReq } from '../../../../../../../ng-swagger-gen/models/optimizer-registration-req';
import { clone } from '../../../../../shared/units/utils';

@Component({
    selector: 'new-optimizer-modal',
    templateUrl: 'new-optimizer-modal.component.html',
    styleUrls: ['../../../../../common.scss'],
    standalone: false,
})
export class NewOptimizerModalComponent {
    testMap: any = {};
    opened: boolean = false;
    @Output() notify = new EventEmitter<Optimizer>();
    @ViewChild(NewOptimizerFormComponent, { static: true })
    newOptimizerFormComponent: NewOptimizerFormComponent;
    checkBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    saveBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
    onTesting: boolean = false;
    onSaving: boolean = false;
    isEdit: boolean = false;
    originValue: any;
    uid: string;
    editOptimizer: Optimizer;
    @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
    constructor(
        private configOptimizerService: OptimizerService,
        private msgHandler: MessageHandlerService,
        private translate: TranslateService
    ) {}
    open(): void {
        // reset
        this.opened = true;
        this.inlineAlert.close();
        this.testMap = {};
        this.newOptimizerFormComponent.showEndpointError = false;
        this.newOptimizerFormComponent.newOptimizerForm.reset({ auth: 'None' });
    }
    close(): void {
        this.opened = false;
    }
    create(): void {
        this.onSaving = true;
        this.saveBtnState = ClrLoadingState.LOADING;
        const optimizer: OptimizerRegistrationReq = { name: '', url: '' };
        const value = this.newOptimizerFormComponent.newOptimizerForm.value;
        optimizer.name = value.name;
        optimizer.description = value.description;
        optimizer.url = value.url;
        if (value.auth === 'None') {
            optimizer.auth = '';
        } else if (value.auth === 'Basic') {
            optimizer.auth = value.auth;
            optimizer.access_credential =
                value.accessCredential.username +
                ':' +
                value.accessCredential.password;
        } else if (value.auth === 'APIKey') {
            optimizer.auth = value.auth;
            optimizer.access_credential = value.accessCredential.apiKey;
        } else {
            optimizer.auth = value.auth;
            optimizer.access_credential = value.accessCredential.token;
        }
        optimizer.skip_certVerify = !!value.skipCertVerify;
        optimizer.use_internal_addr = !!value.useInner;
        this.configOptimizerService
            .createOptimizer({
                registration: optimizer,
            })
            .pipe(finalize(() => (this.onSaving = false)))
            .subscribe(
                response => {
                    this.close();
                    this.msgHandler.showSuccess('OPTIMIZER.ADD_SUCCESS');
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
            this.newOptimizerFormComponent.newOptimizerForm.get('url').value
        ];
    }
    get canTestEndpoint(): boolean {
        if (
            this.newOptimizerFormComponent.newOptimizerForm.get('auth').value ===
            'Basic'
        ) {
            return (
                this.newOptimizerFormComponent.newOptimizerForm
                    .get('accessCredential')
                    .get('username').valid &&
                this.newOptimizerFormComponent.newOptimizerForm
                    .get('accessCredential')
                    .get('password').valid
            );
        }
        if (
            this.newOptimizerFormComponent.newOptimizerForm.get('auth').value ===
            'Bearer'
        ) {
            return this.newOptimizerFormComponent.newOptimizerForm
                .get('accessCredential')
                .get('token').valid;
        }
        if (
            this.newOptimizerFormComponent.newOptimizerForm.get('auth').value ===
            'APIKey'
        ) {
            return this.newOptimizerFormComponent.newOptimizerForm
                .get('accessCredential')
                .get('apiKey').valid;
        }
        return (
            !this.onTesting &&
            this.newOptimizerFormComponent &&
            !this.newOptimizerFormComponent.checkOnGoing &&
            this.newOptimizerFormComponent.newOptimizerForm.get('name').valid &&
            !this.newOptimizerFormComponent.checkEndpointOnGoing &&
            this.newOptimizerFormComponent.newOptimizerForm.get('url').valid
        );
    }
    get valid(): boolean {
        if (
            this.onSaving ||
            this.newOptimizerFormComponent.isNameExisting ||
            this.newOptimizerFormComponent.isEndpointUrlExisting ||
            this.onTesting ||
            !this.newOptimizerFormComponent ||
            this.newOptimizerFormComponent.checkOnGoing ||
            this.newOptimizerFormComponent.checkEndpointOnGoing
        ) {
            return false;
        }
        if (this.newOptimizerFormComponent.newOptimizerForm.get('name').invalid) {
            return false;
        }
        if (this.newOptimizerFormComponent.newOptimizerForm.get('url').invalid) {
            return false;
        }
        if (
            this.newOptimizerFormComponent.newOptimizerForm.get('auth').value ===
            'Basic'
        ) {
            return (
                this.newOptimizerFormComponent.newOptimizerForm
                    .get('accessCredential')
                    .get('username').valid &&
                this.newOptimizerFormComponent.newOptimizerForm
                    .get('accessCredential')
                    .get('password').valid
            );
        }
        if (
            this.newOptimizerFormComponent.newOptimizerForm.get('auth').value ===
            'Bearer'
        ) {
            return this.newOptimizerFormComponent.newOptimizerForm
                .get('accessCredential')
                .get('token').valid;
        }
        if (
            this.newOptimizerFormComponent.newOptimizerForm.get('auth').value ===
            'APIKey'
        ) {
            return this.newOptimizerFormComponent.newOptimizerForm
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
            this.newOptimizerFormComponent.newOptimizerForm.get('name').value
        ) {
            return true;
        }
        if (
            this.originValue.description !==
            this.newOptimizerFormComponent.newOptimizerForm.get('description').value
        ) {
            return true;
        }
        if (
            this.originValue.url !==
            this.newOptimizerFormComponent.newOptimizerForm.get('url').value
        ) {
            return true;
        }
        if (
            this.originValue.auth !==
            this.newOptimizerFormComponent.newOptimizerForm.get('auth').value
        ) {
            return true;
        }
        if (
            this.originValue.skipCertVerify !==
            this.newOptimizerFormComponent.newOptimizerForm.get('skipCertVerify')
                .value
        ) {
            return true;
        }
        if (
            this.originValue.useInner !==
            this.newOptimizerFormComponent.newOptimizerForm.get('useInner').value
        ) {
            return true;
        }
        if (this.originValue.auth === 'Basic') {
            if (
                this.originValue.accessCredential.username !==
                this.newOptimizerFormComponent.newOptimizerForm
                    .get('accessCredential')
                    .get('username').value
            ) {
                return true;
            }
            if (
                this.originValue.accessCredential.password !==
                this.newOptimizerFormComponent.newOptimizerForm
                    .get('accessCredential')
                    .get('password').value
            ) {
                return true;
            }
        }
        if (this.originValue.auth === 'Bearer') {
            if (
                this.originValue.accessCredential.token !==
                this.newOptimizerFormComponent.newOptimizerForm
                    .get('accessCredential')
                    .get('token').value
            ) {
                return true;
            }
        }
        if (this.originValue.auth === 'APIKey') {
            if (
                this.originValue.accessCredential.apiKey !==
                this.newOptimizerFormComponent.newOptimizerForm
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
        const optimizer: OptimizerRegistrationReq = { name: '', url: '' };
        const value = this.newOptimizerFormComponent.newOptimizerForm.value;
        optimizer.name = value.name;
        optimizer.description = value.description;
        optimizer.url = value.url;
        if (value.auth === 'None') {
            optimizer.auth = '';
        } else if (value.auth === 'Basic') {
            optimizer.auth = value.auth;
            optimizer.access_credential =
                value.accessCredential.username +
                ':' +
                value.accessCredential.password;
        } else if (value.auth === 'APIKey') {
            optimizer.auth = value.auth;
            optimizer.access_credential = value.accessCredential.apiKey;
        } else {
            optimizer.auth = value.auth;
            optimizer.access_credential = value.accessCredential.token;
        }
        optimizer.skip_certVerify = !!value.skipCertVerify;
        optimizer.use_internal_addr = !!value.useInner;
        this.configOptimizerService
            .pingOptimizer({
                settings: optimizer,
            })
            .pipe(finalize(() => (this.onTesting = false)))
            .subscribe(
                response => {
                    this.inlineAlert.showInlineSuccess({
                        message: 'OPTIMIZER.TEST_PASS',
                    });
                    this.checkBtnState = ClrLoadingState.SUCCESS;
                    this.testMap[
                        this.newOptimizerFormComponent.newOptimizerForm.get(
                            'url'
                        ).value
                    ] = true;
                },
                error => {
                    this.translate
                        .get('OPTIMIZER.TEST_FAILED', {
                            name: this.newOptimizerFormComponent.newOptimizerForm.get(
                                'name'
                            ).value,
                            url: this.newOptimizerFormComponent.newOptimizerForm.get(
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
        let value = this.newOptimizerFormComponent.newOptimizerForm.value;
        this.editOptimizer.name = value.name;
        this.editOptimizer.description = value.description;
        this.editOptimizer.url = value.url;
        if (value.auth === 'None') {
            this.editOptimizer.auth = '';
        } else if (value.auth === 'Basic') {
            this.editOptimizer.auth = value.auth;
            this.editOptimizer.access_credential =
                value.accessCredential.username +
                ':' +
                value.accessCredential.password;
        } else if (value.auth === 'APIKey') {
            this.editOptimizer.auth = value.auth;
            this.editOptimizer.access_credential = value.accessCredential.apiKey;
        } else {
            this.editOptimizer.auth = value.auth;
            this.editOptimizer.access_credential = value.accessCredential.token;
        }
        this.editOptimizer.skip_certVerify = !!value.skipCertVerify;
        this.editOptimizer.use_internal_addr = !!value.useInner;
        this.editOptimizer.uuid = this.uid;
        const optimizer: OptimizerRegistrationReq = clone(this.editOptimizer);
        this.configOptimizerService
            .updateOptimizer({
                registrationId: this.editOptimizer.uuid,
                registration: optimizer,
            })
            .pipe(finalize(() => (this.onSaving = false)))
            .subscribe(
                response => {
                    this.close();
                    this.msgHandler.showSuccess('OPTIMIZER.UPDATE_SUCCESS');
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
