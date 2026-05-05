// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
