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
import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { map, catchError } from 'rxjs/operators';
import { Observable, throwError as observableThrowError } from 'rxjs';
import { SearchResults } from './search-results';
import { CURRENT_BASE_HREF, HTTP_GET_OPTIONS } from '../../units/utils';

const searchEndpoint = CURRENT_BASE_HREF + '/search';
/**
 * Declare service to handle the global search
 *
 *
 **
 * class GlobalSearchService
 */
@Injectable()
export class GlobalSearchService {
    constructor(private http: HttpClient) {}

    /**
     * Search related artifacts with the provided keyword
     *
     *  ** deprecated param {string} keyword
     * returns {Observable<SearchResults>}
     *
     * @memberOf GlobalSearchService
     */
    doSearch(term: string): Observable<SearchResults> {
        let searchUrl = searchEndpoint + '?q=' + term;

        return this.http.get(searchUrl, HTTP_GET_OPTIONS).pipe(
            map(response => response as SearchResults),
            catchError(error => observableThrowError(error))
        );
    }
}
