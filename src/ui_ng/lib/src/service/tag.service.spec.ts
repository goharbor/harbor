import { TestBed, inject } from '@angular/core/testing';

import { TagService, TagDefaultService } from './tag.service';

describe('TagService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [{
        provide: TagService,
        useClass: TagDefaultService
      }]
    });
  });

  it('should ...', inject([TagDefaultService], (service: TagService) => {
    expect(service).toBeTruthy();
  }));
});
