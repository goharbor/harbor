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
import { Subscription } from 'rxjs';

@Injectable({
    providedIn: 'root',
})
export class EventService {
    private _channels: any = [];
    /**
     * Subscribe to an event topic. Events that get posted to that topic will trigger the provided handler.
     *
     * @param {string} topic the topic to subscribe to
     * @param {function} handler the event handler
     * @return A Subscription to unsubscribe
     */
    subscribe(topic: HarborEvent, handler: Function): Subscription {
        if (!this._channels[topic]) {
            this._channels[topic] = [];
        }
        this._channels[topic].push(handler);
        return new Subscription(() => {
            this.unsubscribe(topic, handler);
        });
    }

    /**
     * Unsubscribe from the given topic. Your handler will no longer receive events published to this topic.
     *
     * @param {string} topic the topic to unsubscribe from
     * @param {function} handler the event handler
     *
     */
    private unsubscribe(topic: HarborEvent, handler: Function = null) {
        let t = this._channels[topic];
        if (!t) {
            // Wasn't found, wasn't removed
            return;
        }
        if (!handler) {
            // Remove all handlers for this topic
            delete this._channels[topic];
            return;
        }
        // We need to find and remove a specific handler
        let i = t.indexOf(handler);
        if (i < 0) {
            // Wasn't found, wasn't removed
            return;
        }
        t.splice(i, 1);
        // If the channel is empty now, remove it from the channel map
        if (!t.length) {
            delete this._channels[topic];
        }
        return;
    }

    /**
     * Publish an event to the given topic.
     * @param topic
     * @param data
     */
    publish(topic: HarborEvent, data?: any) {
        const t = this._channels[topic];
        if (!t) {
            return;
        }
        t.forEach((handler: any) => {
            handler(data);
        });
    }
}

export enum HarborEvent {
    SCROLL = 'scroll',
    SCROLL_TO_POSITION = 'scrollToPosition',
    REFRESH_PROJECT_INFO = 'refreshProjectInfo',
    START_SCAN_ARTIFACT = 'startScanArtifact',
    START_GENERATE_SBOM = 'startGenerateSbom',
    STOP_SCAN_ARTIFACT = 'stopScanArtifact',
    STOP_SBOM_ARTIFACT = 'stopSbomArtifact',
    UPDATE_VULNERABILITY_INFO = 'UpdateVulnerabilityInfo',
    UPDATE_SBOM_INFO = 'UpdateSbomInfo',
    REFRESH_EXPORT_JOBS = 'refreshExportJobs',
    DELETE_ACCESSORY = 'deleteAccessory',
    COPY_DIGEST = 'copyDigest',
    REFRESH_BANNER_MESSAGE = 'refreshBannerMessage',
    RETRIEVED_ICON = 'retrievedIcon',
    THEME_CHANGE = 'themeChange',
}
