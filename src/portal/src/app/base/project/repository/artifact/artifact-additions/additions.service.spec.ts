import { inject, TestBed } from '@angular/core/testing';
import { AdditionsService } from './additions.service';
import {
    HttpClientTestingModule,
    HttpTestingController,
} from '@angular/common/http/testing';

describe('TagRetentionService', () => {
    const testLink: string = '/test';
    const data: string = 'testData';
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [HttpClientTestingModule],
            providers: [AdditionsService],
        });
    });

    it('should be created and get right data', inject(
        [AdditionsService],
        (service: AdditionsService) => {
            expect(service).toBeTruthy();
            service.getDetailByLink(testLink, false, false).subscribe(res => {
                expect(res).toEqual(data);
            });
            const httpTestingController = TestBed.get(HttpTestingController);
            const req = httpTestingController.expectOne(testLink);
            expect(req.request.method).toEqual('GET');
            req.flush(data);
            httpTestingController.verify();
        }
    ));
});
