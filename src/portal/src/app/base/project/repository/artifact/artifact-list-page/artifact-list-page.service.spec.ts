import { inject, TestBed } from '@angular/core/testing';
import { ArtifactListPageService } from './artifact-list-page.service';
import { SharedTestingModule } from '../../../../../shared/shared.module';

describe('ArtifactListPageService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [ArtifactListPageService],
        });
    });

    it('should be initialized', inject(
        [ArtifactListPageService],
        (service: ArtifactListPageService) => {
            expect(service).toBeTruthy();
        }
    ));
});
