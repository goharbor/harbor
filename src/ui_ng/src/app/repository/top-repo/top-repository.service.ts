import { Injectable } from '@angular/core';
import { Headers, Http, RequestOptions } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { TopRepo } from './top-repository';

export const topRepoEndpoint = "/api/repositories/top";
/**
 * Declare service to handle the top repositories
 * 
 * 
 * @export
 * @class GlobalSearchService
 */
@Injectable()
export class TopRepoService {
    private headers = new Headers({
        "Content-Type": 'application/json'
    });
    private options = new RequestOptions({
        headers: this.headers
    });

    constructor(private http: Http) { }

    /**
     * Get top popular repositories
     * 
     * @param {string} keyword
     * @returns {Promise<TopRepo>}
     * 
     * @memberOf GlobalSearchService
     */
    getTopRepos(): Promise<TopRepo[]> {
        return this.http.get(topRepoEndpoint, this.options).toPromise()
            .then(response => response.json() as TopRepo[])
            .catch(error => Promise.reject(error));
    }
}