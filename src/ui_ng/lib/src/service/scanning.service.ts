import { Observable } from 'rxjs/Observable';
import 'rxjs/add/observable/of';
import { Injectable, Inject } from "@angular/core";
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { Http, URLSearchParams } from '@angular/http';
import { HTTP_JSON_OPTIONS } from '../utils';

import { ScanningDetailResult } from './interface';
import { VulnerabilitySeverity, ScanningBaseResult, ScanningResultSummary } from './interface';

/**
 * Get the vulnerabilities scanning results for the specified tag.
 * 
 * @export
 * @abstract
 * @class ScanningResultService
 */
export abstract class ScanningResultService {
    /**
     * Get the summary of vulnerability scanning result.
     * 
     * @abstract
     * @param {string} tagId
     * @returns {(Observable<ScanningResultSummary> | Promise<ScanningResultSummary> | ScanningResultSummary)}
     * 
     * @memberOf ScanningResultService
     */
    abstract getScanningResultSummary(tagId: string): Observable<ScanningResultSummary> | Promise<ScanningResultSummary> | ScanningResultSummary;

    /**
     * Get the detailed vulnerabilities scanning results.
     * 
     * @abstract
     * @param {string} tagId
     * @returns {(Observable<ScanningDetailResult[]> | Promise<ScanningDetailResult[]> | ScanningDetailResult[])}
     * 
     * @memberOf ScanningResultService
     */
    abstract getScanningResults(tagId: string): Observable<ScanningDetailResult[]> | Promise<ScanningDetailResult[]> | ScanningDetailResult[];
}

@Injectable()
export class ScanningResultDefaultService extends ScanningResultService {
    constructor(
        private http: Http,
        @Inject(SERVICE_CONFIG) private config: IServiceConfig) {
        super();
    }

    getScanningResultSummary(tagId: string): Observable<ScanningResultSummary> | Promise<ScanningResultSummary> | ScanningResultSummary {
        if (!tagId || tagId.trim() === '') {
            return Promise.reject('Bad argument');
        }

        return Observable.of({});
    }

    getScanningResults(tagId: string): Observable<ScanningDetailResult[]> | Promise<ScanningDetailResult[]> | ScanningDetailResult[] {
        if (!tagId || tagId.trim() === '') {
            return Promise.reject('Bad argument');
        }

        return Observable.of([]);
    }
}