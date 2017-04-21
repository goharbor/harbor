// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { Headers, Http, RequestOptions } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { SearchResults } from './search-results';

const searchEndpoint = "/api/search";
/**
 * Declare service to handle the global search
 * 
 * 
 * @export
 * @class GlobalSearchService
 */
@Injectable()
export class GlobalSearchService {
    private headers = new Headers({
        "Content-Type": 'application/json'
    });
    private options = new RequestOptions({
        headers: this.headers
    });

    constructor(private http: Http) { }

    /**
     * Search related artifacts with the provided keyword
     * 
     * @param {string} keyword
     * @returns {Promise<SearchResults>}
     * 
     * @memberOf GlobalSearchService
     */
    doSearch(term: string): Promise<SearchResults> {
        let searchUrl = searchEndpoint + "?q=" + term;

        return this.http.get(searchUrl, this.options).toPromise()
            .then(response => response.json() as SearchResults)
            .catch(error => Promise.reject(error));
    }
}