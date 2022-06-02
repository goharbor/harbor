import { TestBed, inject } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { ProjectService } from '../../shared/services';
import { ArtifactService } from '../../../../ng-swagger-gen/services/artifact.service';
import { ArtifactDetailRoutingResolverService } from './artifact-detail-routing-resolver.service';

describe('ArtifactDetailRoutingResolverService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            providers: [
                { provide: ProjectService, useValue: null },
                { provide: ArtifactService, useValue: null },
            ],
        });
    });

    it('should be created', inject(
        [ArtifactDetailRoutingResolverService],
        (service: ArtifactDetailRoutingResolverService) => {
            expect(service).toBeTruthy();
        }
    ));
});
