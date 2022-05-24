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
import {
    Component,
    EventEmitter,
    Input,
    Output,
    ViewChild,
} from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { catchError, finalize, map } from 'rxjs/operators';
import { forkJoin, Observable, throwError as observableThrowError } from 'rxjs';
import { ConfirmationDialogComponent } from '../../../../../shared/components/confirmation-dialog';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../../shared/entities/shared.const';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import {
    clone,
    getPageSizeFromLocalStorage,
    getQueryString,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../../shared/units/utils';
import {
    operateChanges,
    OperateInfo,
    OperationState,
} from '../../../../../shared/components/operation/operate';
import { OperationService } from '../../../../../shared/components/operation/operation.service';
import { ClrDatagridStateInterface } from '@clr/angular';
import { errorHandler } from '../../../../../shared/units/shared.utils';
import { ConfirmationAcknowledgement } from '../../../../global-confirmation-dialog/confirmation-state-message';
import { ConfirmationMessage } from '../../../../global-confirmation-dialog/confirmation-message';
import { HELM_HUB } from '../../../../../shared/services/endpoint.service';
import { BandwidthUnit, Flatten_I18n_MAP } from '../../replication';
import { KB_TO_MB } from '../create-edit-rule/create-edit-rule.component';
import { ReplicationService } from 'ng-swagger-gen/services/replication.service';
import { ReplicationPolicy } from '../../../../../../../ng-swagger-gen/models/replication-policy';

@Component({
    selector: 'hbr-list-replication-rule',
    templateUrl: './list-replication-rule.component.html',
    styleUrls: ['./list-replication-rule.component.scss'],
})
export class ListReplicationRuleComponent {
    @Input() selectedId: number | string;
    @Input() withReplicationJob: boolean;
    @Input() hasCreateReplicationPermission: boolean;
    @Input() hasUpdateReplicationPermission: boolean;
    @Input() hasDeleteReplicationPermission: boolean;
    @Input() hasExecuteReplicationPermission: boolean;
    @Input() searchString: string;
    @Output() selectOne = new EventEmitter<ReplicationPolicy>();
    @Output() editOne = new EventEmitter<ReplicationPolicy>();
    @Output() toggleOne = new EventEmitter<ReplicationPolicy>();
    @Output() hideJobs = new EventEmitter<any>();
    @Output() redirect = new EventEmitter<ReplicationPolicy>();
    @Output() openNewRule = new EventEmitter<any>();
    @Output() replicateManual = new EventEmitter<ReplicationPolicy>();
    rules: ReplicationPolicy[] = [];
    selectedRow: ReplicationPolicy;
    @ViewChild('toggleConfirmDialog')
    toggleConfirmDialog: ConfirmationDialogComponent;
    @ViewChild('deletionConfirmDialog')
    deletionConfirmDialog: ConfirmationDialogComponent;
    page: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.LIST_REPLICATION_RULE_COMPONENT,
        5
    );
    totalCount: number = 0;
    loading: boolean = true;

    constructor(
        private replicationService: ReplicationService,
        private translateService: TranslateService,
        private errorHandlerEntity: ErrorHandler,
        private operationService: OperationService
    ) {}

    trancatedDescription(desc: string): string {
        if (desc.length > 35) {
            return desc.substr(0, 35);
        } else {
            return desc;
        }
    }

    replicateRule(rule: ReplicationPolicy): void {
        this.replicateManual.emit(rule);
    }

    deletionConfirm(message: ConfirmationAcknowledgement) {
        if (
            message &&
            message.source === ConfirmationTargets.POLICY &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            this.deleteOpe(message.data);
        }
        if (
            message &&
            message.source === ConfirmationTargets.REPLICATION &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            const rule: ReplicationPolicy = clone(message.data);
            rule.enabled = !message.data.enabled;
            const opeMessage = new OperateInfo();
            opeMessage.name = rule.enabled
                ? 'REPLICATION.ENABLE_TITLE'
                : 'REPLICATION.DISABLE_TITLE';
            opeMessage.data.id = rule.id;
            opeMessage.state = OperationState.progressing;
            opeMessage.data.name = rule.name;
            this.operationService.publishInfo(opeMessage);
            this.replicationService
                .updateReplicationPolicy({
                    id: rule.id,
                    policy: rule,
                })
                .subscribe({
                    next: () => {
                        this.translateService
                            .get(
                                rule.enabled
                                    ? 'REPLICATION.ENABLE_SUCCESS'
                                    : 'REPLICATION.DISABLE_SUCCESS'
                            )
                            .subscribe(msg => {
                                operateChanges(
                                    opeMessage,
                                    OperationState.success
                                );
                                this.errorHandlerEntity.info(msg);
                                this.refreshRule();
                            });
                    },
                    error: error => {
                        const errMessage = errorHandler(error);
                        this.translateService
                            .get(
                                rule.enabled
                                    ? 'REPLICATION.ENABLE_FAILED'
                                    : 'REPLICATION.DISABLE_FAILED'
                            )
                            .subscribe(msg => {
                                operateChanges(
                                    opeMessage,
                                    OperationState.failure,
                                    msg
                                );
                                this.errorHandlerEntity.error(errMessage);
                            });
                    },
                });
        }
    }

    selectRule(rule: ReplicationPolicy): void {
        if (rule) {
            this.selectedId = rule.id || '';
            this.selectOne.emit(rule);
        }
    }

    redirectTo(rule: ReplicationPolicy): void {
        this.redirect.emit(rule);
    }

    openModal(): void {
        this.openNewRule.emit();
    }

    editRule(rule: ReplicationPolicy) {
        this.editOne.emit(rule);
    }

    deleteRule(rule: ReplicationPolicy) {
        if (rule) {
            let deletionMessage = new ConfirmationMessage(
                'REPLICATION.DELETION_TITLE',
                'REPLICATION.DELETION_SUMMARY',
                rule.name,
                rule,
                ConfirmationTargets.POLICY,
                ConfirmationButtons.DELETE_CANCEL
            );
            this.deletionConfirmDialog.open(deletionMessage);
        }
    }

    deleteOpe(rule: ReplicationPolicy) {
        if (rule) {
            let observableLists: any[] = [];
            observableLists.push(this.delOperate(rule));

            forkJoin(...observableLists).subscribe(
                item => {
                    this.selectedRow = null;
                    this.refreshRule();
                },
                error => {
                    this.errorHandlerEntity.error(error);
                }
            );
        }
    }

    delOperate(rule: ReplicationPolicy): Observable<any> {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_REPLICATION';
        operMessage.data.id = +rule.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = rule.name;
        this.operationService.publishInfo(operMessage);

        return this.replicationService
            .deleteReplicationPolicy({
                id: rule.id,
            })
            .pipe(
                map(() => {
                    this.translateService
                        .get('BATCH.DELETED_SUCCESS')
                        .subscribe(res =>
                            operateChanges(operMessage, OperationState.success)
                        );
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

    operateRule(operation: string, rule: ReplicationPolicy): void {
        let title: string;
        let summary: string;
        let buttons: ConfirmationButtons;
        switch (operation) {
            case 'enable':
                title = 'REPLICATION.ENABLE_TITLE';
                summary = 'REPLICATION.ENABLE_SUMMARY';
                buttons = ConfirmationButtons.ENABLE_CANCEL;
                break;
            case 'disable':
                title = 'REPLICATION.DISABLE_TITLE';
                summary = 'REPLICATION.DISABLE_SUMMARY';
                buttons = ConfirmationButtons.DISABLE_CANCEL;
                break;

            default:
                return;
        }
        // Confirm
        const msg: ConfirmationMessage = new ConfirmationMessage(
            title,
            summary,
            rule.name,
            rule,
            ConfirmationTargets.REPLICATION,
            buttons
        );
        this.deletionConfirmDialog.open(msg);
    }

    clrLoad(state?: ClrDatagridStateInterface) {
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.LIST_REPLICATION_RULE_COMPONENT,
                this.pageSize
            );
        }
        this.loading = true;
        const param: ReplicationService.ListReplicationPoliciesParams = {
            page: this.page,
            pageSize: this.pageSize,
            sort: getSortingString(state),
        };
        if (this.searchString) {
            param.q = encodeURIComponent(`name=~${this.searchString}`);
        } else {
            param.q = getQueryString(state);
        }
        this.replicationService
            .listReplicationPoliciesResponse(param)
            .pipe(finalize(() => (this.loading = false)))
            .subscribe({
                next: response => {
                    // job list hidden
                    this.hideJobs.emit();
                    // Get total count
                    if (response.headers) {
                        let xHeader: string =
                            response.headers.get('x-total-count');
                        if (xHeader) {
                            this.totalCount = parseInt(xHeader, 0);
                        }
                    }
                    this.rules = response.body as ReplicationPolicy[];
                },
                error: error => {
                    this.errorHandlerEntity.error(error);
                },
            });
    }

    refreshRule() {
        this.page = 1;
        this.totalCount = 0;
        this.selectedRow = null;
        this.searchString = null;
        this.clrLoad();
    }

    isHelmHub(srcRegistry: any): boolean {
        return srcRegistry && srcRegistry.type === HELM_HUB;
    }

    getFlattenLevelString(level: number) {
        if (level !== null && Flatten_I18n_MAP[level]) {
            return Flatten_I18n_MAP[level];
        }
        return level;
    }

    getBandwidthStr(speed: number): string {
        if (speed >= KB_TO_MB) {
            return '' + (speed / KB_TO_MB).toFixed(2) + BandwidthUnit.MB;
        }
        if (speed > 0 && speed < KB_TO_MB) {
            return '' + speed + BandwidthUnit.KB;
        }
        return 'REPLICATION.UNLIMITED';
    }
}
