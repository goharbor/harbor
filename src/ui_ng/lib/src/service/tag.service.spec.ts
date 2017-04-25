import { TestBed, inject } from '@angular/core/testing';

import { TagService, TagDefaultService } from './tag.service';

describe('TagService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        TagDefaultService,
        {
          provide: TagService,
          useClass: TagDefaultService
        }]
    });
  });

  it('should be initialized', inject([TagDefaultService], (service: TagService) => {
    expect(service).toBeTruthy();
  }));
});
