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
    subscribe(topic: string, handler: Function): Subscription {
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
    private unsubscribe(topic: string, handler: Function = null) {
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
    publish(topic: string, data?: any) {
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
    STOP_SCAN_ARTIFACT = 'stopScanArtifact',
    UPDATE_VULNERABILITY_INFO = 'UpdateVulnerabilityInfo',
}
