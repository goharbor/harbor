import { TestBed, inject } from '@angular/core/testing';

import { ReplicationService, ReplicationDefaultService } from './replication.service';

describe('ReplicationService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [
        ReplicationDefaultService,
        {
          provide: ReplicationService,
          useClass: ReplicationDefaultService
        }]
    });
  });

  it('should be initialized', inject([ReplicationDefaultService], (service: ReplicationService) => {
    expect(service).toBeTruthy();
  }));
});
