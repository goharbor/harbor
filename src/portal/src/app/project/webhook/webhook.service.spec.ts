import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { WebhookService } from './webhook.service';

describe('WebhookService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule
      ],
      providers: [WebhookService]
    });
  });

  it('should be created', inject([WebhookService], (service: WebhookService) => {
    expect(service).toBeTruthy();
  }));
  it('function eventTypeToText should work', inject([WebhookService], (service: WebhookService) => {
    expect(service).toBeTruthy();
    const eventType: string = 'REPLICATION';
    expect(service.eventTypeToText(eventType)).toEqual('Replication finished');
    const mockedEventType: string = 'TEST';
    expect(service.eventTypeToText(mockedEventType)).toEqual('TEST');
  }));
});
