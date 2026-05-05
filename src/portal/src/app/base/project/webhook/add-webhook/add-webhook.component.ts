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
    Input,
    Output,
    ViewChild,
} from '@angular/core';
import { AddWebhookFormComponent } from '../add-webhook-form/add-webhook-form.component';
import { WebhookPolicy } from '../../../../../../ng-swagger-gen/models/webhook-policy';
import { SupportedWebhookEventTypes } from '../../../../../../ng-swagger-gen/models/supported-webhook-event-types';

@Component({
    selector: 'add-webhook',
    templateUrl: './add-webhook.component.html',
    styleUrls: ['./add-webhook.component.scss'],
})
export class AddWebhookComponent {
    isEdit: boolean;
    isOpen: boolean = false;
    closable: boolean = false;
    staticBackdrop: boolean = true;

    @Input() projectId: number;
    webhook: WebhookPolicy;
    @Input()
    metadata: SupportedWebhookEventTypes;
    @ViewChild(AddWebhookFormComponent)
    addWebhookFormComponent: AddWebhookFormComponent;
    @Output() notify = new EventEmitter<WebhookPolicy>();

    constructor() {}

    openAddWebhookModal() {
        this.isOpen = true;
    }

    onCancel() {
        this.isOpen = false;
    }
    notifySuccess() {
        this.isOpen = false;
        this.notify.emit();
    }
    closeModal() {
        this.isOpen = false;
    }
}
