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
            .catch(error => error);
    }
}