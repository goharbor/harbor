import { TestBed, inject } from '@angular/core/testing';

import { WebhookService } from './webhook.service';

xdescribe('WebhookService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [WebhookService]
    });
  });

  it('should be created', inject([WebhookService], (service: WebhookService) => {
    expect(service).toBeTruthy();
  }));
});
