import { Injectable } from '@angular/core';
import { Headers, Http, RequestOptions } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { Statistics } from './statistics';

export const statisticsEndpoint = "/api/statistics";
/**
 * Declare service to handle the top repositories
 * 
 * 
 * @export
 * @class GlobalSearchService
 */
@Injectable()
export class StatisticsService {
    private headers = new Headers({
        "Content-Type": 'application/json'
    });
    private options = new RequestOptions({
        headers: this.headers
    });

    constructor(private http: Http) { }

    getStatistics(): Promise<Statistics> {
        return this.http.get(statisticsEndpoint, this.options).toPromise()
        .then(response => response.json() as Statistics)
        .catch(error => Promise.reject(error));
    }
}