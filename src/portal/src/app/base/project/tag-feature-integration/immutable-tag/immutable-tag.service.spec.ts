import { ImmutableTagService } from './immutable-tag.service';
import { TestBed, inject } from '@angular/core/testing';

describe('ImmutableTagService', () => {
    beforeEach(() =>
        TestBed.configureTestingModule({
            providers: [ImmutableTagService],
        })
    );

    it('should be created', () => {
        const service: ImmutableTagService = TestBed.get(ImmutableTagService);
        expect(service).toBeTruthy();
    });
    it('should get rules', inject([ImmutableTagService], service => {
        expect(service).toBeTruthy();
    }));
});
