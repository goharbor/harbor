import { TestBed, inject } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { ReplicationService } from "../../../lib/services";
import { ReplicationTasksRoutingResolverService } from "./replication-tasks-routing-resolver.service";

describe('ReplicationTasksRoutingResolverService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        RouterTestingModule
      ],
      providers: [
        { provide: ReplicationService, useValue: null },
      ]
    });
  });

  it('should be created', inject([ReplicationTasksRoutingResolverService], (service: ReplicationTasksRoutingResolverService) => {
    expect(service).toBeTruthy();
  }));
});
