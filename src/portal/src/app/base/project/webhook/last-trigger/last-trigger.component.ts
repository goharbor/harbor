import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';
import { ProjectWebhookService } from '../webhook.service';
import { WebhookLastTrigger } from '../../../../../../ng-swagger-gen/models/webhook-last-trigger';

@Component({
    selector: 'last-trigger',
    templateUrl: 'last-trigger.component.html',
    styleUrls: ['./last-trigger.component.scss'],
})
export class LastTriggerComponent implements OnChanges {
    @Input() inputLastTriggers: WebhookLastTrigger[];
    @Input() webhookName: string;
    lastTriggers: WebhookLastTrigger[] = [];
    constructor(private webhookService: ProjectWebhookService) {}
    ngOnChanges(changes: SimpleChanges): void {
        if (changes && changes['inputLastTriggers']) {
            this.lastTriggers = [];
            this.inputLastTriggers.forEach(item => {
                if (this.webhookName === item.policy_name) {
                    this.lastTriggers.push(item);
                }
            });
        }
    }
    eventTypeToText(eventType: string): string {
        return this.webhookService.eventTypeToText(eventType);
    }
}
