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
import { finalize } from "rxjs/operators";
import { TranslateService } from '@ngx-translate/core';
import { Component, OnInit, ViewChild } from '@angular/core';
import { AddWebhookComponent } from "./add-webhook/add-webhook.component";
import { AddWebhookFormComponent } from "./add-webhook-form/add-webhook-form.component";
import { ActivatedRoute } from '@angular/router';
import { Webhook, LastTrigger } from './webhook';
import { WebhookService } from './webhook.service';
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { Project } from '../project';
import {
  ConfirmationTargets,
  ConfirmationState,
  ConfirmationButtons
} from "../../shared/shared.const";

import { ConfirmationMessage } from "../../shared/confirmation-dialog/confirmation-message";
import { ConfirmationAcknowledgement } from "../../shared/confirmation-dialog/confirmation-state-message";
import { ConfirmationDialogComponent } from "../../shared/confirmation-dialog/confirmation-dialog.component";

@Component({
  templateUrl: './webhook.component.html',
  styleUrls: ['./webhook.component.scss'],
  // changeDetection: ChangeDetectionStrategy.OnPush
})
export class WebhookComponent implements OnInit {
  @ViewChild(AddWebhookComponent)
  addWebhookComponent: AddWebhookComponent;
  @ViewChild(AddWebhookFormComponent)
  addWebhookFormComponent: AddWebhookFormComponent;
  @ViewChild("confirmationDialogComponent")
  confirmationDialogComponent: ConfirmationDialogComponent;
  webhook: Webhook;
  endpoint: string = '';
  lastTriggers: LastTrigger[] = [];
  lastTriggerCount: number = 0;
  isEnabled: boolean;
  loading: boolean = false;
  showCreate: boolean = false;
  projectId: number;
  projectName: string;
  constructor(
    private route: ActivatedRoute,
    private translate: TranslateService,
    private webhookService: WebhookService,
    private messageHandlerService: MessageHandlerService) {}

  ngOnInit() {
    this.projectId = +this.route.snapshot.parent.params['id'];
    let resolverData = this.route.snapshot.parent.data;
    if (resolverData) {
      let project = <Project>(resolverData["projectResolver"]);
      this.projectName = project.name;
    }
    this.getData(this.projectId);
  }

  getData(projectId: number) {
    this.getLastTriggers(projectId);
    this.getWebhook(projectId);
  }

  getLastTriggers(projectId: number) {
    this.loading = true;
    this.webhookService
      .listLastTrigger(projectId)
      .pipe(finalize(() => (this.loading = false)))
      .subscribe(
        response => {
          this.lastTriggers = response;
          this.lastTriggerCount = response.length;
        },
        error => {
          this.messageHandlerService.handleError(error);
        }
      );
  }

  getWebhook(projectId: number) {
    this.webhookService
      .listWebhook(projectId)
      .subscribe(
        response => {
          if (response.length) {
            this.webhook = response[0];
            this.endpoint = this.webhook.targets[0].address;
            this.isEnabled = this.webhook.enabled;
            this.showCreate = false;
          } else {
            this.showCreate = true;
          }
        },
        error => {
          this.messageHandlerService.handleError(error);
        }
      );
  }

  switchWebhookStatus(enabled = false) {
    let content = '';
    this.translate.get(
      enabled
      ? 'WEBHOOK.ENABLED_WEBHOOK_SUMMARY'
      : 'WEBHOOK.DISABLED_WEBHOOK_SUMMARY'
    ).subscribe((res) => content = res + this.projectName);
    let message = new ConfirmationMessage(
      enabled ? 'WEBHOOK.ENABLED_WEBHOOK_TITLE' : 'WEBHOOK.DISABLED_WEBHOOK_TITLE',
      content,
      '',
      {},
      ConfirmationTargets.WEBHOOK,
      enabled ? ConfirmationButtons.ENABLE_CANCEL : ConfirmationButtons.DISABLE_CANCEL
    );
    this.confirmationDialogComponent.open(message);
  }

  confirmSwitch(message: ConfirmationAcknowledgement) {
    if (message &&
        message.source === ConfirmationTargets.WEBHOOK &&
        message.state === ConfirmationState.CONFIRMED) {
        this.webhookService
          .editWebhook(this.projectId, this.webhook.id, Object.assign({}, this.webhook, { enabled: !this.isEnabled }))
          .subscribe(
            response => {
              this.getData(this.projectId);
            },
            error => {
              this.messageHandlerService.handleError(error);
            }
          );
    }
}

  editWebhook(isModify: boolean): void {
    this.getData(this.projectId);
  }

  openAddWebhookModal(): void {
    this.addWebhookComponent.openAddWebhookModal();
  }
}
