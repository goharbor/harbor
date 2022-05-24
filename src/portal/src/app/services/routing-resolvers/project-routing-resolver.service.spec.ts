import { TestBed, inject } from '@angular/core/testing';
import { SessionService } from '../../shared/services/session.service';
import { ProjectRoutingResolver } from './project-routing-resolver.service';
import { RouterTestingModule } from '@angular/router/testing';
import { ProjectService } from '../../shared/services';

describe('ProjectRoutingResolverService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [RouterTestingModule],
            providers: [
                ProjectRoutingResolver,
                { provide: SessionService, useValue: null },
                { provide: ProjectService, useValue: null },
            ],
        });
    });

    it('should be created', inject(
        [ProjectRoutingResolver],
        (service: ProjectRoutingResolver) => {
            expect(service).toBeTruthy();
        }
    ));
});
