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
import {
    Component,
    EventEmitter,
    OnDestroy,
    Output,
    ViewChild,
} from '@angular/core';
import { Router } from '@angular/router';
import { ClrDatagridStateInterface } from '@clr/angular';
import { TranslateService } from '@ngx-translate/core';
import { finalize } from 'rxjs/operators';
import { ModelSyncService } from '../../../../../../ng-swagger-gen/services/model-sync.service';
import { ModelSyncPolicy } from '../../../../../../ng-swagger-gen/models/model-sync-policy';
import { ModelSyncExecution } from '../../../../../../ng-swagger-gen/models/model-sync-execution';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
    PAGE_SIZE_OPTIONS,
    REFRESH_TIME_DIFFERENCE,
} from '../../../../shared/entities/shared.const';
import { ConfirmationDialogComponent } from '../../../../shared/components/confirmation-dialog';
import { ConfirmationMessage } from '../../../global-confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../../../global-confirmation-dialog/confirmation-state-message';
import {
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import { CreateEditModelSyncPolicyComponent } from '../create-edit-policy/create-edit-policy.component';
import { isExecutionInProgress, TRIGGER_SCHEDULED } from '../model-sync';

const ONE_MINUTE_SECONDS = 60;
const STATUS_MAP = {
    Succeed: 'Succeeded',
};
const TRIGGER_I18N_MAP = {
    manual: 'REPLICATION.MANUAL',
    scheduled: 'REPLICATION.SCHEDULED',
};

@Component({
    selector: 'model-sync-policy-list',
    templateUrl: './policy-list.component.html',
    styleUrls: ['./policy-list.component.scss'],
    standalone: false,
})
export class ModelSyncPolicyListComponent implements OnDestroy {
    @Output() goToRegistry = new EventEmitter<void>();
    clrPageSizeOptions: number[] = PAGE_SIZE_OPTIONS;

    // policies
    policies: ModelSyncPolicy[] = [];
    selectedPolicy: ModelSyncPolicy;
    policiesLoading: boolean = true;
    policiesPage: number = 1;
    policiesPageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.MODEL_SYNC_POLICY_LIST_COMPONENT,
        5
    );
    policiesTotal: number = 0;
    searchPolicy: string = '';
    policyOperating: boolean = false;

    // executions of the selected policy
    executions: ModelSyncExecution[] = [];
    selectedExecutions: ModelSyncExecution[] = [];
    executionsLoading: boolean = true;
    executionsPage: number = 1;
    executionsPageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.MODEL_SYNC_POLICY_LIST_COMPONENT_EXECUTIONS
    );
    executionsTotal: number = 0;
    executionsTimeout: any;
    stopOnGoing: boolean = false;
    isOpenFilterTag: boolean = false;
    currentTerm: string = '';
    defaultFilter: string = 'trigger';

    @ViewChild(CreateEditModelSyncPolicyComponent)
    createEditPolicy: CreateEditModelSyncPolicyComponent;
    @ViewChild('policyConfirmDialog')
    policyConfirmDialog: ConfirmationDialogComponent;
    @ViewChild('stopConfirmDialog')
    stopConfirmDialog: ConfirmationDialogComponent;

    constructor(
        private modelSyncService: ModelSyncService,
        private errorHandler: ErrorHandler,
        private translate: TranslateService,
        private router: Router
    ) {}

    ngOnDestroy(): void {
        this.clearExecutionsTimeout();
    }

    // ---------- policies ----------

    clrLoadPolicies(state?: ClrDatagridStateInterface): void {
        if (state && state.page && state.page.size) {
            this.policiesPageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.MODEL_SYNC_POLICY_LIST_COMPONENT,
                this.policiesPageSize
            );
        }
        this.policiesLoading = true;
        const params: ModelSyncService.ListModelSyncPoliciesParams = {
            page: this.policiesPage,
            pageSize: this.policiesPageSize,
            sort: getSortingString(state),
        };
        if (this.searchPolicy) {
            params.q = encodeURIComponent(`name=~${this.searchPolicy}`);
        }
        this.modelSyncService
            .listModelSyncPoliciesResponse(params)
            .pipe(finalize(() => (this.policiesLoading = false)))
            .subscribe(
                res => {
                    const total = res.headers.get('X-Total-Count');
                    this.policiesTotal = total ? parseInt(total, 10) : 0;
                    this.policies = res.body || [];
                    // keep the selection in sync with the refreshed data
                    if (this.selectedPolicy) {
                        const refreshed = this.policies.find(
                            p => p.id === this.selectedPolicy.id
                        );
                        this.selectedPolicy = refreshed || null;
                        if (!this.selectedPolicy) {
                            this.executions = [];
                            this.executionsTotal = 0;
                        }
                    }
                },
                error => this.errorHandler.error(error)
            );
    }

    refreshPolicies(): void {
        this.policiesPage = 1;
        this.searchPolicy = '';
        this.clrLoadPolicies({ page: {} });
    }

    doSearchPolicies(value: string): void {
        this.policiesPage = 1;
        this.searchPolicy = (value || '').trim();
        this.clrLoadPolicies({ page: {} });
    }

    selectPolicy(policy: ModelSyncPolicy): void {
        this.selectedPolicy = policy;
        this.selectedExecutions = [];
        this.executionsPage = 1;
        this.clearExecutionsTimeout();
        if (policy) {
            this.clrLoadExecutions(true, { page: {} });
        } else {
            this.executions = [];
            this.executionsTotal = 0;
        }
    }

    openCreatePolicy(): void {
        this.createEditPolicy.openCreate();
    }

    openEditPolicy(policy: ModelSyncPolicy): void {
        if (policy) {
            this.createEditPolicy.openEdit(policy);
        }
    }

    onPolicySaved(): void {
        this.clrLoadPolicies({ page: {} });
    }

    triggerI18n(policy: ModelSyncPolicy): string {
        const type = policy?.trigger?.type;
        return type ? TRIGGER_I18N_MAP[type] || type : '';
    }

    isScheduled(policy: ModelSyncPolicy): boolean {
        return !!policy?.trigger && policy.trigger.type === TRIGGER_SCHEDULED;
    }

    truncatedDescription(desc: string): string {
        return desc.length > 35 ? desc.substr(0, 35) : desc;
    }

    destination(policy: ModelSyncPolicy): string {
        if (!policy) {
            return '';
        }
        const project =
            policy.dest_project_name || `#${policy.dest_project_id}`;
        return `${project}/${policy.dest_repository || ''}`;
    }

    goRegistry(): void {
        this.goToRegistry.emit();
        this.router.navigate(['harbor', 'registries']);
    }

    revision(policy: ModelSyncPolicy): string {
        return policy?.src_revision || 'main';
    }

    // ---------- policy actions ----------

    confirmSync(policy: ModelSyncPolicy): void {
        if (!policy) {
            return;
        }
        const message = new ConfirmationMessage(
            'MODEL_SYNC.SYNC_TITLE',
            'MODEL_SYNC.SYNC_SUMMARY',
            policy.name,
            policy,
            ConfirmationTargets.MODEL_SYNC_EXECUTE,
            ConfirmationButtons.CONFIRM_CANCEL
        );
        this.policyConfirmDialog.open(message);
    }

    confirmDelete(policy: ModelSyncPolicy): void {
        if (!policy) {
            return;
        }
        const message = new ConfirmationMessage(
            'MODEL_SYNC.DELETION_TITLE',
            'MODEL_SYNC.DELETION_SUMMARY',
            policy.name,
            policy,
            ConfirmationTargets.MODEL_SYNC_DELETE,
            ConfirmationButtons.DELETE_CANCEL
        );
        this.policyConfirmDialog.open(message);
    }

    confirmToggle(policy: ModelSyncPolicy): void {
        if (!policy) {
            return;
        }
        const message = new ConfirmationMessage(
            policy.enabled
                ? 'MODEL_SYNC.DISABLE_TITLE'
                : 'MODEL_SYNC.ENABLE_TITLE',
            policy.enabled
                ? 'MODEL_SYNC.DISABLE_SUMMARY'
                : 'MODEL_SYNC.ENABLE_SUMMARY',
            policy.name,
            policy,
            ConfirmationTargets.MODEL_SYNC_TOGGLE,
            policy.enabled
                ? ConfirmationButtons.DISABLE_CANCEL
                : ConfirmationButtons.ENABLE_CANCEL
        );
        this.policyConfirmDialog.open(message);
    }

    onPolicyConfirm(message: ConfirmationAcknowledgement): void {
        if (!message || message.state !== ConfirmationState.CONFIRMED) {
            return;
        }
        const policy: ModelSyncPolicy = message.data;
        switch (message.source) {
            case ConfirmationTargets.MODEL_SYNC_EXECUTE:
                this.startSync(policy);
                break;
            case ConfirmationTargets.MODEL_SYNC_DELETE:
                this.deletePolicy(policy);
                break;
            case ConfirmationTargets.MODEL_SYNC_TOGGLE:
                this.togglePolicy(policy);
                break;
        }
    }

    startSync(policy: ModelSyncPolicy): void {
        this.policyOperating = true;
        this.modelSyncService
            .startModelSync({ id: policy.id })
            .pipe(finalize(() => (this.policyOperating = false)))
            .subscribe(
                () => {
                    this.translate
                        .get('MODEL_SYNC.SYNC_STARTED', { param: policy.name })
                        .subscribe(res => this.errorHandler.info(res));
                    this.selectedPolicy = policy;
                    this.executionsPage = 1;
                    this.clrLoadExecutions(true, { page: {} });
                },
                error => this.errorHandler.error(error)
            );
    }

    deletePolicy(policy: ModelSyncPolicy): void {
        this.policyOperating = true;
        this.modelSyncService
            .deleteModelSyncPolicy({ id: policy.id })
            .pipe(finalize(() => (this.policyOperating = false)))
            .subscribe(
                () => {
                    this.translate
                        .get('MODEL_SYNC.DELETED_SUCCESS')
                        .subscribe(res => this.errorHandler.info(res));
                    if (this.selectedPolicy?.id === policy.id) {
                        this.selectPolicy(null);
                    }
                    this.clrLoadPolicies({ page: {} });
                },
                error => this.errorHandler.error(error)
            );
    }

    togglePolicy(policy: ModelSyncPolicy): void {
        this.policyOperating = true;
        const updated: ModelSyncPolicy = {
            ...policy,
            enabled: !policy.enabled,
        };
        this.modelSyncService
            .updateModelSyncPolicy({ id: policy.id, policy: updated })
            .pipe(finalize(() => (this.policyOperating = false)))
            .subscribe(
                () => {
                    this.translate
                        .get('MODEL_SYNC.UPDATED_SUCCESS')
                        .subscribe(res => this.errorHandler.info(res));
                    this.clrLoadPolicies({ page: {} });
                },
                error => this.errorHandler.error(error)
            );
    }

    // ---------- executions ----------

    clrLoadExecutions(
        withLoading: boolean,
        state: ClrDatagridStateInterface
    ): void {
        if (!this.selectedPolicy || !state || !state.page) {
            return;
        }
        if (state.page.size) {
            this.executionsPageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.MODEL_SYNC_POLICY_LIST_COMPONENT_EXECUTIONS,
                this.executionsPageSize
            );
        }
        this.clearExecutionsTimeout();
        if (withLoading) {
            this.executionsLoading = true;
        }
        const policyId = this.selectedPolicy.id;
        const params: ModelSyncService.ListModelSyncExecutionsParams = {
            id: policyId,
            page: this.executionsPage,
            pageSize: this.executionsPageSize,
            sort: getSortingString(state) || '-start_time',
        };
        if (this.currentTerm) {
            if (this.defaultFilter === 'trigger') {
                params.trigger = this.currentTerm;
            } else {
                // convert 'Succeeded' back to the API status
                params.status =
                    this.currentTerm === STATUS_MAP.Succeed
                        ? 'Succeed'
                        : this.currentTerm;
            }
        }
        this.modelSyncService
            .listModelSyncExecutionsResponse(params)
            .pipe(finalize(() => (this.executionsLoading = false)))
            .subscribe(
                res => {
                    // the selection may have changed while loading
                    if (this.selectedPolicy?.id !== policyId) {
                        return;
                    }
                    const total = res.headers.get('X-Total-Count');
                    this.executionsTotal = total ? parseInt(total, 10) : 0;
                    this.executions = res.body || [];
                    if (this.executions.some(e => isExecutionInProgress(e))) {
                        this.executionsTimeout = setTimeout(() => {
                            this.clrLoadExecutions(false, { page: {} });
                        }, REFRESH_TIME_DIFFERENCE);
                    }
                },
                error => this.errorHandler.error(error)
            );
    }

    refreshExecutions(): void {
        this.executionsPage = 1;
        this.currentTerm = '';
        this.clrLoadExecutions(true, { page: {} });
    }

    doSearchExecutions(terms: string): void {
        this.currentTerm = (terms || '').trim();
        this.executionsPage = 1;
        this.clrLoadExecutions(true, { page: {} });
    }

    doFilterExecutions($event: any): void {
        this.defaultFilter = $event['target'].value;
        this.doSearchExecutions(this.currentTerm);
    }

    openFilter(isOpen: boolean): void {
        this.isOpenFilterTag = isOpen;
    }

    getStatusStr(status: string, statusText: string): string {
        // an execution marked as failed with "Execution skipped" was skipped, not failed
        if (
            status === 'Failed' &&
            statusText &&
            statusText.startsWith('Execution skipped')
        ) {
            return 'Skipped';
        }
        return STATUS_MAP[status] || status;
    }

    getDuration(e: ModelSyncExecution): string {
        if (!e || !e.start_time) {
            return '-';
        }
        const start = new Date(e.start_time).getTime();
        const end = e.end_time ? new Date(e.end_time).getTime() : Date.now();
        const diff = end - start;
        const seconds = diff / 1000;
        const minutes = Math.floor(seconds / ONE_MINUTE_SECONDS);
        const remain = Math.floor(seconds % ONE_MINUTE_SECONDS);
        if (minutes > 0) {
            return remain === 0 ? minutes + 'm' : minutes + 'm' + remain + 's';
        }
        if (remain > 0) {
            return remain + 's';
        }
        return diff > 0 ? diff + 'ms' : '-';
    }

    clearExecutionsTimeout(): void {
        if (this.executionsTimeout) {
            clearTimeout(this.executionsTimeout);
            this.executionsTimeout = null;
        }
    }

    canStop(): boolean {
        return (
            this.selectedExecutions.length > 0 &&
            this.selectedExecutions.every(e => isExecutionInProgress(e))
        );
    }

    openStopDialog(): void {
        if (!this.canStop()) {
            return;
        }
        const ids = this.selectedExecutions.map(e => e.id).join(',');
        const message = new ConfirmationMessage(
            'MODEL_SYNC.STOP_TITLE',
            'MODEL_SYNC.STOP_SUMMARY',
            ids,
            this.selectedExecutions,
            ConfirmationTargets.MODEL_SYNC_STOP,
            ConfirmationButtons.STOP_CANCEL
        );
        this.stopConfirmDialog.open(message);
    }

    onStopConfirm(message: ConfirmationAcknowledgement): void {
        if (
            !message ||
            message.state !== ConfirmationState.CONFIRMED ||
            message.source !== ConfirmationTargets.MODEL_SYNC_STOP
        ) {
            return;
        }
        const executions: ModelSyncExecution[] = message.data || [];
        this.stopOnGoing = true;
        let remaining = executions.length;
        executions.forEach(e => {
            this.modelSyncService
                .stopModelSync({ executionId: e.id })
                .pipe(
                    finalize(() => {
                        remaining--;
                        if (remaining <= 0) {
                            this.stopOnGoing = false;
                            this.selectedExecutions = [];
                            this.refreshExecutions();
                        }
                    })
                )
                .subscribe(
                    () => {
                        this.translate
                            .get('MODEL_SYNC.STOP_SUCCESS', { param: e.id })
                            .subscribe(res => this.errorHandler.info(res));
                    },
                    error => this.errorHandler.error(error)
                );
        });
    }

    goToTasks(execution: ModelSyncExecution): void {
        this.router.navigate([
            'harbor',
            'model-sync',
            'executions',
            execution.id,
            'tasks',
        ]);
    }
}
