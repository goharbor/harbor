import { Component, Input, OnDestroy } from '@angular/core';
import { WebhookPolicy } from '../../../../../../ng-swagger-gen/models/webhook-policy';
import { finalize } from 'rxjs/operators';
import { Router } from '@angular/router';
import { WebhookService } from '../../../../../../ng-swagger-gen/services/webhook.service';
import {
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../../shared/units/utils';
import { Execution } from '../../../../../../ng-swagger-gen/models/execution';
import { ClrDatagridStateInterface } from '@clr/angular';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import {
    EXECUTION_STATUS,
    TIME_OUT,
} from '../../p2p-provider/p2p-provider.service';
import { ProjectWebhookService, VendorType } from '../webhook.service';

@Component({
    selector: 'app-executions',
    templateUrl: './executions.component.html',
    styleUrls: ['./executions.component.scss'],
})
export class ExecutionsComponent implements OnDestroy {
    @Input()
    selectedWebhook: WebhookPolicy;
    executions: Execution[] = [];
    loading: boolean = true;
    currentPage: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.WEBHOOK_EXECUTIONS_COMPONENT
    );
    total: number = 0;
    state: ClrDatagridStateInterface;
    timeout: any;
    constructor(
        private webhookService: WebhookService,
        private messageHandlerService: MessageHandlerService,
        private router: Router,
        private projectWebhookService: ProjectWebhookService
    ) {}

    ngOnDestroy(): void {
        this.clearLoop();
    }
    clrLoadExecutions(
        state: ClrDatagridStateInterface,
        withLoading: boolean,
        policyId: number
    ) {
        if (state) {
            this.state = state;
        }
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.WEBHOOK_EXECUTIONS_COMPONENT,
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
            // sort by start_time desc by default
            sort = `-start_time`;
        }
        if (withLoading) {
            this.loading = true;
        }
        this.webhookService
            .ListExecutionsOfWebhookPolicyResponse({
                webhookPolicyId: policyId ? policyId : this.selectedWebhook.id,
                projectNameOrId: this.selectedWebhook.project_id.toString(),
                pageSize: this.pageSize,
                page: this.currentPage,
                sort: sort,
                q: q,
            })
            .pipe(
                finalize(() => {
                    this.loading = false;
                })
            )
            .subscribe({
                next: res => {
                    if (res.headers) {
                        let xHeader: string = res.headers.get('X-Total-Count');
                        if (xHeader) {
                            this.total = parseInt(xHeader, 0);
                        }
                    }
                    this.executions = res.body || [];
                    this.setLoop();
                },
                error: err => {
                    this.messageHandlerService.error(err);
                },
            });
    }
    goToLink(id: number) {
        const linkUrl = [
            'harbor',
            'projects',
            `${this.selectedWebhook.project_id}`,
            'webhook',
            `${this.selectedWebhook.id}`,
            'executions',
            `${id}`,
            'tasks',
        ];
        this.router.navigate(linkUrl);
    }
    toString(v: any) {
        if (v) {
            return JSON.stringify(v);
        }
        return '';
    }
    toJson(v: any) {
        if (v) {
            return JSON.parse(v);
        }
        return null;
    }
    refreshExecutions(shouldReset: boolean, policyId: number) {
        if (shouldReset) {
            this.executions = [];
            this.currentPage = 1;
        }
        this.clrLoadExecutions(this.state, true, policyId);
    }

    clearLoop() {
        if (this.timeout) {
            clearTimeout(this.timeout);
            this.timeout = null;
        }
    }

    setLoop() {
        this.clearLoop();
        if (this.executions && this.executions.length) {
            for (let i = 0; i < this.executions.length; i++) {
                if (this.willChangStatus(this.executions[i].status)) {
                    if (!this.timeout) {
                        this.timeout = setTimeout(() => {
                            this.clrLoadExecutions(this.state, false, null);
                        }, TIME_OUT);
                    }
                }
            }
        }
    }

    willChangStatus(status: string): boolean {
        return (
            status === EXECUTION_STATUS.PENDING ||
            status === EXECUTION_STATUS.RUNNING ||
            status === EXECUTION_STATUS.SCHEDULED
        );
    }

    eventTypeToText(eventType: any): string {
        return this.projectWebhookService.eventTypeToText(eventType);
    }

    useJsonFormat(vendorType: string): boolean {
        return vendorType === VendorType.WEBHOOK;
    }
}
