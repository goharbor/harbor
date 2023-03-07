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
import { finalize } from 'rxjs/operators';
import { TranslateService } from '@ngx-translate/core';
import { Component, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { AddWebhookComponent } from './add-webhook/add-webhook.component';
import { AddWebhookFormComponent } from './add-webhook-form/add-webhook-form.component';
import { ActivatedRoute, NavigationEnd, Router } from '@angular/router';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { Project } from '../project';
import {
    clone,
    getPageSizeFromLocalStorage,
    getSortingString,
    PageSizeMapKeys,
    setPageSizeToLocalStorage,
} from '../../../shared/units/utils';
import { forkJoin, Observable, Subscription } from 'rxjs';
import {
    UserPermissionService,
    USERSTATICPERMISSION,
} from '../../../shared/services';
import { ClrDatagridStateInterface } from '@clr/angular';
import { ConfirmationDialogComponent } from '../../../shared/components/confirmation-dialog';
import {
    ConfirmationButtons,
    ConfirmationState,
    ConfirmationTargets,
} from '../../../shared/entities/shared.const';
import { ConfirmationMessage } from '../../global-confirmation-dialog/confirmation-message';
import { WebhookService } from '../../../../../ng-swagger-gen/services/webhook.service';
import { WebhookPolicy } from '../../../../../ng-swagger-gen/models/webhook-policy';
import {
    PAYLOAD_FORMATS,
    PAYLOAD_FORMAT_I18N_MAP,
    ProjectWebhookService,
} from './webhook.service';
import { ExecutionsComponent } from './excutions/executions.component';
import {
    EventService,
    HarborEvent,
} from '../../../services/event-service/event.service';
import { SupportedWebhookEventTypes } from '../../../../../ng-swagger-gen/models/supported-webhook-event-types';
// The route path which will display this component
const URL_TO_DISPLAY: RegExp = /^\/harbor\/projects\/(\d+)\/webhook$/;
@Component({
    templateUrl: './webhook.component.html',
    styleUrls: ['./webhook.component.scss'],
})
export class WebhookComponent implements OnInit, OnDestroy {
    @ViewChild(AddWebhookComponent)
    addWebhookComponent: AddWebhookComponent;
    @ViewChild(AddWebhookFormComponent)
    addWebhookFormComponent: AddWebhookFormComponent;
    @ViewChild('confirmationDialogComponent')
    confirmationDialogComponent: ConfirmationDialogComponent;
    projectId: number;
    projectName: string;
    selectedRow: WebhookPolicy;
    webhookList: WebhookPolicy[] = [];
    metadata: SupportedWebhookEventTypes;
    loadingMetadata: boolean = true;
    loadingWebhookList: boolean = true;
    hasCreatPermission: boolean = false;
    hasUpdatePermission: boolean = false;
    page: number = 1;
    pageSize: number = getPageSizeFromLocalStorage(
        PageSizeMapKeys.WEBHOOK_COMPONENT
    );
    total: number = 0;
    state: ClrDatagridStateInterface;
    @ViewChild(ExecutionsComponent)
    executionsComponent: ExecutionsComponent;
    routerSub: Subscription;
    scrollSub: Subscription;
    scrollTop: number;
    constructor(
        private route: ActivatedRoute,
        private translate: TranslateService,
        private webhookService: WebhookService,
        private projectWebhookService: ProjectWebhookService,
        private messageHandlerService: MessageHandlerService,
        private userPermissionService: UserPermissionService,
        private router: Router,
        private event: EventService
    ) {}

    ngOnInit() {
        if (!this.scrollSub) {
            this.scrollSub = this.event.subscribe(HarborEvent.SCROLL, v => {
                if (v && URL_TO_DISPLAY.test(v.url)) {
                    this.scrollTop = v.scrollTop;
                }
            });
        }
        if (!this.routerSub) {
            this.routerSub = this.router.events.subscribe(e => {
                if (e instanceof NavigationEnd) {
                    if (e && URL_TO_DISPLAY.test(e.url)) {
                        // Into view
                        this.event.publish(
                            HarborEvent.SCROLL_TO_POSITION,
                            this.scrollTop
                        );
                    } else {
                        this.event.publish(HarborEvent.SCROLL_TO_POSITION, 0);
                    }
                }
            });
        }
        this.projectId = +this.route.snapshot.parent.parent.params['id'];
        let resolverData = this.route.snapshot.parent.parent.data;
        if (resolverData) {
            let project = <Project>resolverData['projectResolver'];
            this.projectName = project.name;
        }
        this.getData();
        this.getPermissions();
    }

    ngOnDestroy(): void {
        if (this.routerSub) {
            this.routerSub.unsubscribe();
            this.routerSub = null;
        }
        if (this.scrollSub) {
            this.scrollSub.unsubscribe();
            this.scrollSub = null;
        }
    }

    getPermissions() {
        const permissionsList: Observable<boolean>[] = [];
        permissionsList.push(
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.WEBHOOK.KEY,
                USERSTATICPERMISSION.WEBHOOK.VALUE.CREATE
            )
        );
        permissionsList.push(
            this.userPermissionService.getPermission(
                this.projectId,
                USERSTATICPERMISSION.WEBHOOK.KEY,
                USERSTATICPERMISSION.WEBHOOK.VALUE.UPDATE
            )
        );
        forkJoin(...permissionsList).subscribe(
            Rules => {
                [this.hasCreatPermission, this.hasUpdatePermission] = Rules;
            },
            error => {
                this.messageHandlerService.error(error);
            }
        );
    }
    refresh() {
        this.page = 1;
        this.total = 0;
        this.selectedRow = null;
        this.getWebhooks(this.state);
        this.getData();
    }
    getData() {
        this.getMetadata();
        this.selectedRow = null;
    }
    getMetadata() {
        this.loadingMetadata = true;
        this.webhookService
            .GetSupportedEventTypes({
                projectNameOrId: this.projectId.toString(),
            })
            .pipe(finalize(() => (this.loadingMetadata = false)))
            .subscribe(
                response => {
                    this.metadata = response;
                    if (this.metadata && this.metadata.event_type) {
                        // sort by text
                        this.metadata.event_type.sort(
                            (a: string, b: string) => {
                                if (
                                    this.eventTypeToText(a) ===
                                    this.eventTypeToText(b)
                                ) {
                                    return 0;
                                }
                                return this.eventTypeToText(a) >
                                    this.eventTypeToText(b)
                                    ? 1
                                    : -1;
                            }
                        );
                    }
                },
                error => {
                    this.messageHandlerService.handleError(error);
                }
            );
    }

    getWebhooks(state?: ClrDatagridStateInterface) {
        if (state) {
            this.state = state;
        }
        if (state && state.page) {
            this.pageSize = state.page.size;
            setPageSizeToLocalStorage(
                PageSizeMapKeys.WEBHOOK_COMPONENT,
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
        this.loadingWebhookList = true;
        this.webhookService
            .ListWebhookPoliciesOfProjectResponse({
                projectNameOrId: this.projectId.toString(),
                page: this.page,
                pageSize: this.pageSize,
                sort: sort,
                q: q,
            })
            .pipe(finalize(() => (this.loadingWebhookList = false)))
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
                    this.webhookList = response.body || [];
                },
                error => {
                    this.messageHandlerService.handleError(error);
                }
            );
    }

    switchWebhookStatus() {
        let content = '';
        this.translate
            .get(
                !this.selectedRow.enabled
                    ? 'WEBHOOK.ENABLED_WEBHOOK_SUMMARY'
                    : 'WEBHOOK.DISABLED_WEBHOOK_SUMMARY',
                { name: this.selectedRow.name }
            )
            .subscribe(res => {
                content = res;
                let message = new ConfirmationMessage(
                    !this.selectedRow.enabled
                        ? 'WEBHOOK.ENABLED_WEBHOOK_TITLE'
                        : 'WEBHOOK.DISABLED_WEBHOOK_TITLE',
                    content,
                    '',
                    {},
                    ConfirmationTargets.WEBHOOK,
                    !this.selectedRow.enabled
                        ? ConfirmationButtons.ENABLE_CANCEL
                        : ConfirmationButtons.DISABLE_CANCEL
                );
                this.confirmationDialogComponent.open(message);
            });
    }

    confirmSwitch(message) {
        if (
            message &&
            message.source === ConfirmationTargets.WEBHOOK &&
            message.state === ConfirmationState.CONFIRMED
        ) {
            if (JSON.stringify(message.data) === '{}') {
                this.webhookService
                    .UpdateWebhookPolicyOfProject({
                        projectNameOrId: this.projectId.toString(),
                        webhookPolicyId: this.selectedRow.id,
                        policy: Object.assign({}, this.selectedRow, {
                            enabled: !this.selectedRow.enabled,
                        }),
                    })
                    .subscribe(
                        response => {
                            this.refresh();
                        },
                        error => {
                            this.messageHandlerService.handleError(error);
                        }
                    );
            } else {
                this.webhookService
                    .DeleteWebhookPolicyOfProject({
                        projectNameOrId: this.projectId.toString(),
                        webhookPolicyId: message.data.id,
                    })
                    .subscribe({
                        next: res => {
                            this.refresh();
                        },
                        error: err => {
                            this.messageHandlerService.handleError(err);
                        },
                    });
            }
        }
    }

    editWebhook() {
        if (this.metadata) {
            this.addWebhookComponent.isOpen = true;
            this.addWebhookComponent.isEdit = true;
            this.addWebhookComponent.addWebhookFormComponent.isModify = true;
            this.addWebhookComponent.addWebhookFormComponent.webhook = clone(
                this.selectedRow
            );
            this.addWebhookComponent.addWebhookFormComponent.originValue =
                clone(this.selectedRow);
            this.addWebhookComponent.addWebhookFormComponent.webhook.event_types =
                clone(this.selectedRow.event_types);
        }
    }

    openAddWebhookModal(): void {
        this.addWebhookComponent.openAddWebhookModal();
    }
    newWebhook() {
        if (this.metadata) {
            this.addWebhookComponent.isOpen = true;
            this.addWebhookComponent.isEdit = false;
            this.addWebhookComponent.addWebhookFormComponent.isModify = false;
            this.addWebhookComponent.addWebhookFormComponent.currentForm.reset({
                notifyType: this.metadata.notify_type[0],
                payloadFormat: PAYLOAD_FORMATS[0],
            });
            this.addWebhookComponent.addWebhookFormComponent.webhook = {
                enabled: true,
                event_types: [],
                targets: [
                    {
                        type: 'http',
                        address: '',
                        skip_cert_verify: true,
                        payload_format: PAYLOAD_FORMATS[0],
                    },
                ],
            };
            this.addWebhookComponent.addWebhookFormComponent.webhook.event_types =
                clone(this.metadata.event_type);
        }
    }
    success() {
        this.refresh();
    }

    deleteWebhook() {
        const names: string[] = [];
        names.push(this.selectedRow.name);
        let content = '';
        this.translate
            .get('WEBHOOK.DELETE_WEBHOOK_SUMMARY', { names: names.join(',') })
            .subscribe(res => (content = res));
        const msg: ConfirmationMessage = new ConfirmationMessage(
            'SCANNER.CONFIRM_DELETION',
            content,
            names.join(','),
            this.selectedRow,
            ConfirmationTargets.WEBHOOK,
            ConfirmationButtons.DELETE_CANCEL
        );
        this.confirmationDialogComponent.open(msg);
    }
    eventTypeToText(eventType: string): string {
        return this.projectWebhookService.eventTypeToText(eventType);
    }
    refreshExecutions(e: WebhookPolicy) {
        if (e) {
            this.executionsComponent?.refreshExecutions(true, e.id);
        }
    }

    getI18nKey(v: string) {
        if (v && PAYLOAD_FORMAT_I18N_MAP[v]) {
            return PAYLOAD_FORMAT_I18N_MAP[v];
        }
        return v;
    }
}
