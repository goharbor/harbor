// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { Component, OnInit, OnDestroy, ViewChild } from '@angular/core';
import {
    Subscription,
    Observable,
    forkJoin,
    throwError as observableThrowError,
} from 'rxjs';
import { TranslateService } from '@ngx-translate/core';
import { Comparator } from '../../../shared/services';
import { ErrorHandler } from '../../../shared/units/error-handler';
import { map, catchError, finalize } from 'rxjs/operators';
import { ConfirmationDialogComponent } from '../../../shared/components/confirmation-dialog';
import {
    ConfirmationTargets,
    ConfirmationState,
    ConfirmationButtons,
} from '../../../shared/entities/shared.const';
import { CreateEditEndpointComponent } from './create-edit-endpoint/create-edit-endpoint.component';
import {
    CustomComparator,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../shared/units/utils';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../shared/components/operation/operate';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { errorHandler } from '../../../shared/units/shared.utils';
import { ConfirmationMessage } from '../../global-confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../../global-confirmation-dialog/confirmation-state-message';
import {
    EndpointService,
    HELM_HUB,
} from '../../../shared/services/endpoint.service';
import { RegistryService } from '../../../../../ng-swagger-gen/services/registry.service';
import { ClrDatagridStateInterface } from '@clr/angular';
import { Registry } from '../../../../../ng-swagger-gen/models/registry';

@Component({
    selector: 'hbr-endpoint',
    templateUrl: './endpoint.component.html',
    styleUrls: ['./endpoint.component.scss'],
})
export class EndpointComponent implements OnInit, OnDestroy {
    @ViewChild(CreateEditEndpointComponent)
    createEditEndpointComponent: CreateEditEndpointComponent;

    @ViewChild('confirmationDialog')
    confirmationDialogComponent: ConfirmationDialogComponent;

    targets: Registry[];
    target: Registry;

    targetName: string;
    subscription: Subscription;

    loading: boolean = true;

    creationTimeComparator: Comparator<Registry> =
        new CustomComparator<Registry>('creation_time', 'date');

    timerHandler: any;
    selectedRow: Registry[] = [];

    get initEndpoint(): Registry {
        return {
            credential: {
                access_key: '',
                access_secret: '',
                type: '',
            },
            description: '',
            insecure: false,
            name: '',
            type: '',
            url: '',
        };
    }

    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.SYSTEM_ENDPOINT_COMPONENT
    );
    page: number = 1;
    total: number = 0;
    constructor(
        private endpointService: RegistryService,
        private errorHandlerEntity: ErrorHandler,
        private translateService: TranslateService,
        private operationService: OperationService,
        private oldEndpointService: EndpointService
    ) {}

    ngOnInit(): void {
        this.targetName = '';
    }

    ngOnDestroy(): void {
        if (this.subscription) {
            this.subscription.unsubscribe();
        }
    }
    retrieve(state?: ClrDatagridStateInterface): void {
        this.selectedRow = [];
        let q: string = '';
        if (state && state.filters && state.filters.length) {
            this.targetName = '';
            q = encodeURIComponent(
                `${state.filters[0].property}=~${state.filters[0].value}`
            );
        } else if (this.targetName) {
            q = `name=~${this.targetName}`;
        }
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.SYSTEM_ENDPOINT_COMPONENT,
                this.pageSize
            );
        }
        let sort: string;
        if (state && state.sort && state.sort.by) {
            sort = getSortingString(state);
        } else {
            // sort by creation_time desc by default
            sort = `-creation_time`;
        }
        this.loading = true;
        this.endpointService
            .listRegistriesResponse({
                q: q,
                pageSize: this.pageSize,
                page: this.page,
                sort: sort,
            })
            .pipe(
                finalize(() => {
                    this.loading = false;
                })
            )
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
                    this.targets = response.body || [];
                },
                error => {
                    this.errorHandlerEntity.error(error);
                }
            );
    }

    doSearchTargets(targetName: string) {
        this.targetName = targetName;
        this.page = 1;
        this.total = 0;
        this.selectedRow = [];
        this.retrieve();
    }

    refreshTargets() {
        this.targetName = '';
        this.page = 1;
        this.total = 0;
        this.selectedRow = [];
        this.retrieve();
    }
    openModal() {
        this.createEditEndpointComponent.openCreateEditTarget(true);
        this.target = this.initEndpoint;
    }

    editTargets(targets: Registry[]) {
        if (targets && targets.length === 1) {
            let target = targets[0];
            let editable = true;
            if (!target.id) {
                return;
            }
            let id: number | string = target.id;
            this.createEditEndpointComponent.openCreateEditTarget(editable, id);
        }
    }

    deleteTargets(targets: Registry[]) {
        if (targets && targets.length) {
            let targetNames: string[] = [];
            targets.forEach(target => {
                targetNames.push(target.name);
            });
            let deletionMessage = new ConfirmationMessage(
                'REPLICATION.DELETION_TITLE_TARGET',
                'REPLICATION.DELETION_SUMMARY_TARGET',
                targetNames.join(', ') || '',
                targets,
                ConfirmationTargets.TARGET,
                ConfirmationButtons.DELETE_CANCEL
            );
            this.confirmationDialogComponent.open(deletionMessage);
        }
    }

    confirmDeletion(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.source === ConfirmationTargets.TARGET &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            let targetLists: Registry[] = message.data;
            if (targetLists && targetLists.length) {
                let observableLists: any[] = [];
                targetLists.forEach(target => {
                    observableLists.push(this.delOperate(target));
                });
                forkJoin(...observableLists)
                    .pipe(
                        finalize(() => {
                            this.refreshTargets();
                        })
                    )
                    .subscribe(
                        item => {},
                        error => {
                            this.errorHandlerEntity.error(error);
                        }
                    );
            }
        }
    }
    delOperate(target: Registry): Observable<any> {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_REGISTRY';
        operMessage.data.id = target.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = target.name;
        this.operationService.publishInfo(operMessage);
        return this.endpointService
            .deleteRegistry({
                id: target.id,
            })
            .pipe(
                map(response => {
                    this.translateService
                        .get('BATCH.DELETED_SUCCESS')
                        .subscribe(res => {
                            operateChanges(operMessage, OperationState.success);
                        });
                }),
                catchError(error => {
                    const message = errorHandler(error);
                    this.translateService
                        .get(message)
                        .subscribe(res =>
                            operateChanges(
                                operMessage,
                                OperationState.failure,
                                res
                            )
                        );
                    return observableThrowError(error);
                })
            );
    }
    getAdapterText(adapter: string): string {
        return this.oldEndpointService.getAdapterText(adapter);
    }
    isHelmHub(str: string): boolean {
        return str === HELM_HUB;
    }
}
