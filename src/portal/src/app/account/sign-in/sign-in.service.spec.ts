import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { SignInService } from './sign-in.service';

describe('SignInService', () => {
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [SignInService],
        });
    });

    it('should be created', inject(
        [SignInService],
        (service: SignInService) => {
            expect(service).toBeTruthy();
        }
    ));
});
