import { TestBed, inject } from '@angular/core/testing';

import { SearchTriggerService } from './search-trigger.service';

describe('SearchTriggerService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [SearchTriggerService]
    });
  });

  it('should be created', inject([SearchTriggerService], (service: SearchTriggerService) => {
    expect(service).toBeTruthy();
  }));
});
