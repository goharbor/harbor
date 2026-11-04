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
import { Component, ViewChild, OnInit, OnDestroy } from '@angular/core';
import { Optimizer } from './optimizer';
import { NewOptimizerModalComponent } from './new-optimizer-modal/new-optimizer-modal.component';
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
    PAGE_SIZE_OPTIONS,
} from '../../../../shared/entities/shared.const';
import { ConfirmationMessage } from '../../../global-confirmation-dialog/confirmation-message';
import { OptimizerService } from '../../../../../../ng-swagger-gen/services/optimizer.service';
import { ClrDatagridStateInterface } from '@clr/angular';
import { OptimizerRegistrationReq } from '../../../../../../ng-swagger-gen/models/optimizer-registration-req';

@Component({
    selector: 'config-optimizer',
    templateUrl: 'config-optimizer.component.html',
    styleUrls: [
        './config-optimizer.component.scss',
        '../../config/config.component.scss',
    ],
    standalone: false,
})
export class ConfigurationOptimizerComponent implements OnInit, OnDestroy {
    clrPageSizeOptions: number[] = PAGE_SIZE_OPTIONS;
    optimizers: Optimizer[] = [];
    selectedRow: Optimizer;
    onGoing: boolean = true;
    @ViewChild(NewOptimizerModalComponent)
    newOptimizerDialog: NewOptimizerModalComponent;
    deletionSubscription: any;
    page: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.SYSTEM_OPTIMIZER_COMPONENT
    );
    total: number = 0;
    state: ClrDatagridStateInterface;
    constructor(
        private configOptimizerService: OptimizerService,
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
                            confirmed.source === ConfirmationTargets.OPTIMIZER &&
                            confirmed.state === ConfirmationState.CONFIRMED
                        ) {
                            this.configOptimizerService
                                .deleteOptimizer({
                                    registrationId: confirmed.data[0].uuid,
                                })
                                .subscribe(
                                    response => {
                                        this.msgHandler.showSuccess(
                                            'OPTIMIZER.DELETE_SUCCESS'
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
        this.getOptimizers(this.state);
    }
    getOptimizers(state?: ClrDatagridStateInterface) {
        this.state = state;
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.SYSTEM_OPTIMIZER_COMPONENT,
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
        this.configOptimizerService
            .listOptimizersResponse({
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
                    this.optimizers = response.body || [];
                    this.getMetadataForAll();
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }
    getMetadataForAll() {
        if (this.optimizers && this.optimizers.length > 0) {
            this.optimizers.forEach((optimizer, index) => {
                if (optimizer.uuid) {
                    this.optimizers[index].loadingMetadata = true;
                    this.configOptimizerService
                        .getOptimizerMetadata({
                            registrationId: optimizer.uuid,
                        })
                        .pipe(
                            finalize(
                                () =>
                                    (this.optimizers[index].loadingMetadata =
                                        false)
                            )
                        )
                        .subscribe(
                            response => {
                                this.optimizers[index].metadata = response;
                            },
                            error => {
                                this.optimizers[index].metadata = null;
                            }
                        );
                }
            });
        }
    }

    addNewOptimizer(): void {
        this.newOptimizerDialog.open();
        this.newOptimizerDialog.isEdit = false;
        this.newOptimizerDialog.newOptimizerFormComponent.isEdit = false;
    }
    addSuccess() {
        this.getOptimizers();
    }

    supportCapability(optimizer: Optimizer, capabilityType: string): boolean {
        return optimizer && optimizer.capabilities && capabilityType
            ? optimizer?.capabilities?.[`support_${capabilityType}`] ?? false
            : false;
    }

    changeStat() {
        if (this.selectedRow) {
            let optimizer: OptimizerRegistrationReq = clone(this.selectedRow);
            optimizer.disabled = !optimizer.disabled;
            this.configOptimizerService
                .updateOptimizer({
                    registrationId: this.selectedRow.uuid,
                    registration: optimizer,
                })
                .subscribe(
                    response => {
                        this.msgHandler.showSuccess('OPTIMIZER.UPDATE_SUCCESS');
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
            this.configOptimizerService
                .setOptimizerAsDefault({
                    registrationId: this.selectedRow.uuid,
                    payload: {
                        is_default: true,
                    },
                })
                .subscribe(
                    response => {
                        this.msgHandler.showSuccess('OPTIMIZER.UPDATE_SUCCESS');
                        this.refresh();
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }
    deleteOptimizers() {
        if (this.selectedRow) {
            // Confirm deletion
            let msg: ConfirmationMessage = new ConfirmationMessage(
                'OPTIMIZER.CONFIRM_DELETION',
                'OPTIMIZER.DELETION_SUMMARY',
                this.selectedRow.name,
                [this.selectedRow],
                ConfirmationTargets.OPTIMIZER,
                ConfirmationButtons.DELETE_CANCEL
            );
            this.deletionDialogService.openComfirmDialog(msg);
        }
    }
    editOptimizer() {
        if (this.selectedRow) {
            this.newOptimizerDialog.open();
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
            this.newOptimizerDialog.newOptimizerFormComponent.newOptimizerForm.reset(
                resetValue
            );
            this.newOptimizerDialog.isEdit = true;
            this.newOptimizerDialog.newOptimizerFormComponent.isEdit = true;
            this.newOptimizerDialog.uid = this.selectedRow.uuid;
            this.newOptimizerDialog.originValue = clone(resetValue);
            this.newOptimizerDialog.newOptimizerFormComponent.originValue =
                clone(resetValue);
            this.newOptimizerDialog.editOptimizer = clone(this.selectedRow);
        }
    }
}
