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
import { TestBed, inject, getTestBed } from '@angular/core/testing';
import {
    HttpClientTestingModule,
    HttpTestingController,
} from '@angular/common/http/testing';
import { GlobalSearchService } from './global-search.service';
import { Injector } from '@angular/core';
import { SearchResults } from './search-results';
import { CURRENT_BASE_HREF } from '../../units/utils';

describe('GlobalSearchService', () => {
    let injector: TestBed;
    let service: GlobalSearchService;
    let httpMock: HttpTestingController;

    beforeEach(() => {
        TestBed.configureTestingModule({
            providers: [GlobalSearchService],
            imports: [HttpClientTestingModule],
        });
        injector = getTestBed();
        service = injector.get(GlobalSearchService);
        httpMock = injector.get(HttpTestingController);
    });

    it('should be created', inject(
        [GlobalSearchService],
        (service1: GlobalSearchService) => {
            expect(service1).toBeTruthy();
        }
    ));
    it('doSearch should return data', () => {
        service.doSearch('library').subscribe(res => {
            expect(res).toEqual(new SearchResults());
        });

        const req = httpMock.expectOne(CURRENT_BASE_HREF + '/search?q=library');
        expect(req.request.method).toBe('GET');
        req.flush(new SearchResults());
    });
});
