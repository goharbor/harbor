import { TestBed, inject } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { ReplicationTasksRoutingResolverService } from './replication-tasks-routing-resolver.service';
import { ReplicationService } from '../../../../ng-swagger-gen/services';

describe('ReplicationTasksRoutingResolverService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            providers: [{ provide: ReplicationService, useValue: null }],
        });
    });

    it('should be created', inject(
        [ReplicationTasksRoutingResolverService],
        (service: ReplicationTasksRoutingResolverService) => {
            expect(service).toBeTruthy();
        }
    ));
});
