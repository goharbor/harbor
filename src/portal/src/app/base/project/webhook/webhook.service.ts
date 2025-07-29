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
import { Injectable } from '@angular/core';
import { MarkdownPipe } from 'ngx-markdown/src/markdown.pipe';

const EVENT_TYPES_TEXT_MAP = {
    REPLICATION: 'Replication status changed',
    PUSH_ARTIFACT: 'Artifact pushed',
    PULL_ARTIFACT: 'Artifact pulled',
    DELETE_ARTIFACT: 'Artifact deleted',
    DOWNLOAD_CHART: 'Chart downloaded',
    UPLOAD_CHART: 'Chart uploaded',
    DELETE_CHART: 'Chart deleted',
    QUOTA_EXCEED: 'Quota exceed',
    QUOTA_WARNING: 'Quota near threshold',
    SCANNING_FAILED: 'Scanning failed',
    SCANNING_STOPPED: 'Scanning stopped',
    SCANNING_COMPLETED: 'Scanning finished',
    TAG_RETENTION: 'Tag retention finished',
    ROBOT_EXPIRED: 'Robot account expired',
};

export const PAYLOAD_FORMATS: string[] = ['Default', 'CloudEvents'];

export const PAYLOAD_FORMAT_I18N_MAP = {
    [PAYLOAD_FORMATS[0]]: 'SCANNER.DEFAULT',
    [PAYLOAD_FORMATS[1]]: 'WEBHOOK.CLOUD_EVENT',
};

export enum WebhookType {
    HTTP = 'http',
    SLACK = 'slack',
}

export enum VendorType {
    WEBHOOK = 'WEBHOOK',
    SLACK = 'SLACK',
}

@Injectable()
export class ProjectWebhookService {
    constructor() {}
    public eventTypeToText(eventType: string): string {
        if (EVENT_TYPES_TEXT_MAP[eventType]) {
            return EVENT_TYPES_TEXT_MAP[eventType];
        }
        return eventType;
    }
}
