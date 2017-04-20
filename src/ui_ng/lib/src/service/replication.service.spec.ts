import { TestBed, inject } from '@angular/core/testing';

import { ReplicationService, ReplicationDefaultService } from './replication.service';

describe('ReplicationService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [{
        provide: ReplicationService,
        useClass: ReplicationDefaultService
      }]
    });
  });

  it('should ...', inject([ReplicationDefaultService], (service: ReplicationService) => {
    expect(service).toBeTruthy();
  }));
});
