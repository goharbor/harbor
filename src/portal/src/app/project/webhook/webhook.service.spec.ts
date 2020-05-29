import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { WebhookService } from './webhook.service';
import { UserPermissionService } from '@harbor/ui';

describe('WebhookService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule
      ],
      providers: [WebhookService, UserPermissionService]
    });
  });

  it('should be created', inject([WebhookService], (service: WebhookService) => {
    expect(service).toBeTruthy();
  }));
});
