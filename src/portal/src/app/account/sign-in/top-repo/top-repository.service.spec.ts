import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { TopRepoService } from './top-repository.service';

describe('TopRepoService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [TopRepoService],
        });
    });

    it('should be created', inject(
        [TopRepoService],
        (service: TopRepoService) => {
            expect(service).toBeTruthy();
        }
    ));
});
