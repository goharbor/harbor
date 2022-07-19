import { inject, TestBed } from '@angular/core/testing';
import { ProjectWebhookService } from './webhook.service';

describe('WebhookService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [ProjectWebhookService],
        });
    });

    it('should be created', inject(
        [ProjectWebhookService],
        (service: ProjectWebhookService) => {
            expect(service).toBeTruthy();
        }
    ));
    it('function eventTypeToText should work', inject(
        [ProjectWebhookService],
        (service: ProjectWebhookService) => {
            expect(service).toBeTruthy();
            const eventType: string = 'REPLICATION';
            expect(service.eventTypeToText(eventType)).toEqual(
                'Replication finished'
            );
            const mockedEventType: string = 'TEST';
            expect(service.eventTypeToText(mockedEventType)).toEqual('TEST');
        }
    ));
});
