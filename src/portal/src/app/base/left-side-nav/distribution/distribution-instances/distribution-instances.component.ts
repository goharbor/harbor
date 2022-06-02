import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';
import {
    Subscription,
    Observable,
    forkJoin,
    throwError as observableThrowError,
    of,
} from 'rxjs';
import { DistributionSetupModalComponent } from '../distribution-setup-modal/distribution-setup-modal.component';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../shared/components/operation/operate';
import { TranslateService } from '@ngx-translate/core';
import { map, catchError, finalize } from 'rxjs/operators';
import {
    clone,
    getPageSizeFromLocalStorage,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import { Instance } from '../../../../../../ng-swagger-gen/models/instance';
import { PreheatService } from '../../../../../../ng-swagger-gen/services/preheat.service';
import { Metadata } from '../../../../../../ng-swagger-gen/models/metadata';
import { FrontInstance, HEALTHY, UNHEALTHY } from '../distribution-interface';
import { ClrDatagridStateInterface } from '@clr/angular';
import { ConfirmationDialogService } from '../../../global-confirmation-dialog/confirmation-dialog.service';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../shared/entities/shared.const';
import { errorHandler } from '../../../../shared/units/shared.utils';
import { ConfirmationMessage } from '../../../global-confirmation-dialog/confirmation-message';
import { HttpErrorResponse } from '@angular/common/http';

interface MultiOperateData {
    operation: string;
    instances: Instance[];
}

const DEFAULT_ICON: string = 'images/harbor-logo.svg';
const KRAKEN_ICON: string = 'images/kraken-logo-color.svg';
const ONE_THOUSAND: number = 1000;
const KRAKEN: string = 'kraken';

@Component({
    selector: 'dist-instances',
    templateUrl: './distribution-instances.component.html',
    styleUrls: ['./distribution-instances.component.scss'],
})
export class DistributionInstancesComponent implements OnInit, OnDestroy {
    instances: FrontInstance[] = [];
    selectedRow: FrontInstance[] = [];

    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.DISTRIBUTION_INSTANCE_COMPONENT
    );
    currentPage: number = 1;
    totalCount: number = 0;
    queryString: string;

    chanSub: Subscription;

    private loading: boolean = true;
    private operationSubscription: Subscription;

    @ViewChild('setupModal')
    setupModal: DistributionSetupModalComponent;
    providerMap: { [key: string]: Metadata } = {};
    providers: Metadata[] = [];
    constructor(
        private disService: PreheatService,
        private msgHandler: MessageHandlerService,
        private translate: TranslateService,
        private operationDialogService: ConfirmationDialogService,
        private operationService: OperationService
    ) {
        // subscribe operation
        this.operationSubscription =
            operationDialogService.confirmationConfirm$.subscribe(confirmed => {
                if (
                    confirmed &&
                    confirmed.source === ConfirmationTargets.INSTANCE &&
                    confirmed.state === ConfirmationState.CONFIRMED
                ) {
                    this.operateInstance(confirmed.data);
                }
            });
    }

    public get inProgress(): boolean {
        return this.loading;
    }

    ngOnInit() {
        this.getProviders();
    }

    ngOnDestroy() {
        if (this.operationSubscription) {
            this.operationSubscription.unsubscribe();
        }
        if (this.chanSub) {
            this.chanSub.unsubscribe();
        }
    }

    getProviders() {
        this.disService.ListProviders().subscribe(
            providers => {
                if (providers && providers.length) {
                    this.providers = providers;
                    providers.forEach(item => {
                        this.providerMap[item.id] = item;
                    });
                }
            },
            err => this.msgHandler.error(err)
        );
    }

    loadData(state?: ClrDatagridStateInterface) {
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.DISTRIBUTION_INSTANCE_COMPONENT,
                this.pageSize
            );
        }
        this.selectedRow = [];
        const queryParam: PreheatService.ListInstancesParams = {
            page: this.currentPage,
            pageSize: this.pageSize,
        };
        if (this.queryString) {
            queryParam.q = encodeURIComponent(`name=~${this.queryString}`);
        }
        this.loading = true;
        this.disService
            .ListInstancesResponse(queryParam)
            .pipe(finalize(() => (this.loading = false)))
            .subscribe(
                response => {
                    this.totalCount = Number.parseInt(
                        response.headers.get('x-total-count'),
                        10
                    );
                    this.instances = response.body as Instance[];
                    this.pingInstances();
                },
                err => this.msgHandler.error(err)
            );
    }
    pingInstances() {
        if (this.instances && this.instances.length) {
            this.instances.forEach((item, index) => {
                this.disService
                    .PingInstances({ instance: this.handleInstance(item) })
                    .pipe(
                        finalize(
                            () => (this.instances[index].hasCheckHealth = true)
                        )
                    )
                    .subscribe(
                        res => {
                            this.instances[index].pingStatus = HEALTHY;
                        },
                        error => {
                            this.instances[index].pingStatus = UNHEALTHY;
                        }
                    );
            });
        }
    }

    refresh() {
        this.queryString = null;
        this.currentPage = 1;
        this.loadData();
    }

    doFilter($evt: any) {
        this.currentPage = 1;
        this.queryString = $evt;
        this.loadData();
    }

    addInstance() {
        this.setupModal.openSetupModal(false);
    }

    editInstance() {
        if (this.selectedRow && this.selectedRow.length === 1) {
            this.setupModal.openSetupModal(true, clone(this.selectedRow[0]));
        }
    }

    setAsDefault() {
        if (this.selectedRow && this.selectedRow.length === 1) {
            const operMessage = new OperateInfo();
            operMessage.name = 'DISTRIBUTION.SET_AS_DEFAULT';
            operMessage.data.id = this.selectedRow[0].id;
            operMessage.state = OperationState.progressing;
            operMessage.data.name = this.selectedRow[0].name;
            this.operationService.publishInfo(operMessage);
            const instance: Instance = clone(this.selectedRow[0]);
            instance.default = true;
            this.disService
                .UpdateInstance({
                    instance: this.handleInstance(instance),
                    preheatInstanceName: this.selectedRow[0].name,
                })
                .subscribe(
                    () => {
                        this.translate
                            .get('DISTRIBUTION.SET_DEFAULT_SUCCESS')
                            .subscribe(msg => {
                                operateChanges(
                                    operMessage,
                                    OperationState.success
                                );
                                this.msgHandler.info(msg);
                            });
                        this.refresh();
                    },
                    error => {
                        const message = errorHandler(error);
                        this.translate
                            .get('DISTRIBUTION.SET_DEFAULT_FAILED')
                            .subscribe(msg => {
                                operateChanges(
                                    operMessage,
                                    OperationState.failure,
                                    msg
                                );
                                this.translate
                                    .get(message)
                                    .subscribe(errMsg => {
                                        this.msgHandler.error(
                                            msg + ': ' + errMsg
                                        );
                                    });
                            });
                    }
                );
        }
    }
    // Operate the specified Instance
    operateInstances(operation: string, instances: Instance[]): void {
        let arr: string[] = [];
        let title: string;
        let summary: string;
        let buttons: ConfirmationButtons;

        switch (operation) {
            case 'delete':
                title = 'DISTRIBUTION.DELETION_TITLE';
                summary = 'DISTRIBUTION.DELETION_SUMMARY';
                buttons = ConfirmationButtons.DELETE_CANCEL;
                break;
            case 'enable':
                title = 'DISTRIBUTION.ENABLE_TITLE';
                summary = 'DISTRIBUTION.ENABLE_SUMMARY';
                buttons = ConfirmationButtons.ENABLE_CANCEL;
                break;
            case 'disable':
                title = 'DISTRIBUTION.DISABLE_TITLE';
                summary = 'DISTRIBUTION.DISABLE_SUMMARY';
                buttons = ConfirmationButtons.DISABLE_CANCEL;
                break;

            default:
                return;
        }

        if (instances && instances.length) {
            instances.forEach(instance => {
                arr.push(instance.name);
            });
        }
        // Confirm
        let msg: ConfirmationMessage = new ConfirmationMessage(
            title,
            summary,
            arr.join(','),
            { operation: operation, instances: instances },
            ConfirmationTargets.INSTANCE,
            buttons
        );
        this.operationDialogService.openComfirmDialog(msg);
    }

    operateInstance(data: MultiOperateData) {
        let observableLists: any[] = [];
        if (data.instances && data.instances.length) {
            switch (data.operation) {
                case 'delete':
                    data.instances.forEach(instance => {
                        observableLists.push(this.deleteInstance(instance));
                    });
                    break;

                case 'enable':
                    data.instances.forEach(instance => {
                        observableLists.push(this.enableInstance(instance));
                    });
                    break;

                case 'disable':
                    data.instances.forEach(instance => {
                        observableLists.push(this.disableInstance(instance));
                    });
                    break;
            }

            forkJoin(...observableLists).subscribe(
                resArr => {
                    if (data.operation === 'delete') {
                        let error;
                        let errorCount: number = 0;
                        if (resArr && resArr.length) {
                            resArr.forEach(item => {
                                // only record the last error
                                if (item instanceof HttpErrorResponse) {
                                    error = errorHandler(item);
                                    errorCount += 1;
                                }
                            });
                        }
                        if (errorCount === 0) {
                            // All successful
                            this.translate
                                .get('DISTRIBUTION.DELETED_SUCCESS')
                                .subscribe(msg => {
                                    this.msgHandler.info(msg);
                                });
                            this.selectedRow = [];
                            this.refresh();
                        } else if (
                            resArr &&
                            resArr.length === errorCount &&
                            error
                        ) {
                            // All failed
                            this.msgHandler.handleError(error);
                        } else if (error) {
                            // Partly failed
                            this.msgHandler.handleError(error);
                            this.selectedRow = [];
                            this.refresh();
                        }
                    } else {
                        this.selectedRow = [];
                        this.refresh();
                    }
                },
                error => {
                    this.msgHandler.error(error);
                }
            );
        }
    }
    deleteInstance(instance: Instance): Observable<any> {
        let operMessage = new OperateInfo();
        operMessage.name = 'DISTRIBUTION.DELETE_INSTANCE';
        operMessage.data.id = instance.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = instance.name;
        this.operationService.publishInfo(operMessage);

        return this.disService
            .DeleteInstance({ preheatInstanceName: instance.name })
            .pipe(
                map(() => {
                    this.translate
                        .get('DISTRIBUTION.DELETED_SUCCESS')
                        .subscribe(msg => {
                            operateChanges(operMessage, OperationState.success);
                        });
                }),
                catchError(error => {
                    const message = errorHandler(error);
                    this.translate
                        .get('DISTRIBUTION.DELETED_FAILED')
                        .subscribe(msg => {
                            this.translate.get(message).subscribe(errMsg => {
                                operateChanges(
                                    operMessage,
                                    OperationState.failure,
                                    msg + ': ' + errMsg
                                );
                            });
                        });
                    return of(error);
                })
            );
    }

    enableInstance(instance: Instance) {
        let operMessage = new OperateInfo();
        operMessage.name = 'DISTRIBUTION.ENABLE_INSTANCE';
        operMessage.data.id = instance.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = instance.name;
        this.operationService.publishInfo(operMessage);
        const copiedInstance: Instance = clone(instance);
        copiedInstance.enabled = true;
        return this.disService
            .UpdateInstance({
                instance: this.handleInstance(copiedInstance),
                preheatInstanceName: instance.name,
            })
            .pipe(
                map(() => {
                    this.translate
                        .get('DISTRIBUTION.ENABLE_SUCCESS')
                        .subscribe(msg => {
                            operateChanges(operMessage, OperationState.success);
                            this.msgHandler.info(msg);
                        });
                }),
                catchError(error => {
                    const message = errorHandler(error);
                    this.translate
                        .get('DISTRIBUTION.ENABLE_FAILED')
                        .subscribe(msg => {
                            operateChanges(
                                operMessage,
                                OperationState.failure,
                                msg
                            );
                            this.translate.get(message).subscribe(errMsg => {
                                this.msgHandler.error(msg + ': ' + errMsg);
                            });
                        });
                    return observableThrowError(error);
                })
            );
    }

    disableInstance(instance: Instance) {
        let operMessage = new OperateInfo();
        operMessage.name = 'DISTRIBUTION.DISABLE_INSTANCE';
        operMessage.data.id = instance.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = instance.name;
        this.operationService.publishInfo(operMessage);
        const copiedInstance: Instance = clone(instance);
        copiedInstance.enabled = false;
        return this.disService
            .UpdateInstance({
                instance: this.handleInstance(copiedInstance),
                preheatInstanceName: instance.name,
            })
            .pipe(
                map(() => {
                    this.translate
                        .get('DISTRIBUTION.DISABLE_SUCCESS')
                        .subscribe(msg => {
                            operateChanges(operMessage, OperationState.success);
                            this.msgHandler.info(msg);
                        });
                }),
                catchError(error => {
                    const message = errorHandler(error);
                    this.translate
                        .get('DISTRIBUTION.DISABLE_FAILED')
                        .subscribe(msg => {
                            operateChanges(
                                operMessage,
                                OperationState.failure,
                                msg
                            );
                            this.translate.get(message).subscribe(errMsg => {
                                this.msgHandler.error(msg + ': ' + errMsg);
                            });
                        });
                    return observableThrowError(error);
                })
            );
    }

    fmtTime(time: number) {
        let date = new Date();
        return date.setTime(time * ONE_THOUSAND);
    }
    showDefaultIcon(event: any, vendor: string) {
        if (event && event.target) {
            if (KRAKEN === vendor) {
                event.target.src = KRAKEN_ICON;
            } else {
                event.target.src = DEFAULT_ICON;
            }
        }
    }
    handleInstance(instance: FrontInstance): FrontInstance {
        if (instance) {
            const copyOne: FrontInstance = clone(instance);
            delete copyOne.hasCheckHealth;
            delete copyOne.pingStatus;
            return copyOne;
        }
        return instance;
    }
}
