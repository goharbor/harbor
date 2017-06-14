import { Observable } from 'rxjs/Observable';
import 'rxjs/add/observable/of';
import { Injectable, Inject } from "@angular/core";
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { Http, URLSearchParams } from '@angular/http';
import { HTTP_JSON_OPTIONS } from '../utils';

import {
    VulnerabilityItem,
    VulnerabilitySummary
} from './interface';

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
     * @returns {(Observable<VulnerabilitySummary> | Promise<VulnerabilitySummary> | VulnerabilitySummary)}
     * 
     * @memberOf ScanningResultService
     */
    abstract getVulnerabilityScanningSummary(tagId: string): Observable<VulnerabilitySummary> | Promise<VulnerabilitySummary> | VulnerabilitySummary;

    /**
     * Get the detailed vulnerabilities scanning results.
     * 
     * @abstract
     * @param {string} tagId
     * @returns {(Observable<VulnerabilityItem[]> | Promise<VulnerabilityItem[]> | VulnerabilityItem[])}
     * 
     * @memberOf ScanningResultService
     */
    abstract getVulnerabilityScanningResults(tagId: string): Observable<VulnerabilityItem[]> | Promise<VulnerabilityItem[]> | VulnerabilityItem[];
}

@Injectable()
export class ScanningResultDefaultService extends ScanningResultService {
    constructor(
        private http: Http,
        @Inject(SERVICE_CONFIG) private config: IServiceConfig) {
        super();
    }

    getVulnerabilityScanningSummary(tagId: string): Observable<VulnerabilitySummary> | Promise<VulnerabilitySummary> | VulnerabilitySummary {
        if (!tagId || tagId.trim() === '') {
            return Promise.reject('Bad argument');
        }

        return Observable.of({});
    }

    getVulnerabilityScanningResults(tagId: string): Observable<VulnerabilityItem[]> | Promise<VulnerabilityItem[]> | VulnerabilityItem[] {
        if (!tagId || tagId.trim() === '') {
            return Promise.reject('Bad argument');
        }

        return Observable.of([]);
    }
}