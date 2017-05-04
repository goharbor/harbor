import { Observable } from 'rxjs/Observable';
import { RequestQueryParams } from './RequestQueryParams';
import { AccessLog } from './interface';
import { Injectable, Inject } from "@angular/core";
import 'rxjs/add/observable/of';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { Http, URLSearchParams } from '@angular/http';
import { HTTP_JSON_OPTIONS } from '../utils';

/**
 * Define service methods to handle the access log related things.
 * 
 * @export
 * @abstract
 * @class AccessLogService
 */
export abstract class AccessLogService {
    /**
     * Get the audit logs for the specified project.
     * Set query parameters through 'queryParams', support:
     *  - page
     *  - pageSize
     * 
     * @abstract
     * @param {(number | string)} projectId
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<AccessLog[]> | AccessLog[])}
     * 
     * @memberOf AccessLogService
     */
    abstract getAuditLogs(projectId: number | string, queryParams?: RequestQueryParams): Observable<AccessLog[]> | Promise<AccessLog[]> | AccessLog[];

    /**
     * Get the recent logs.
     * 
     * @abstract
     * @param {number} lines : Specify how many lines should be returned.
     * @returns {(Observable<AccessLog[]> | AccessLog[])}
     * 
     * @memberOf AccessLogService
     */
    abstract getRecentLogs(lines: number): Observable<AccessLog[]> | Promise<AccessLog[]> | AccessLog[];
}

/**
 * Implement a default service for access log.
 * 
 * @export
 * @class AccessLogDefaultService
 * @extends {AccessLogService}
 */
@Injectable()
export class AccessLogDefaultService extends AccessLogService {
    constructor(
        private http: Http,
        @Inject(SERVICE_CONFIG) private config: IServiceConfig) {
        super();
    }

    public getAuditLogs(projectId: number | string, queryParams?: RequestQueryParams): Observable<AccessLog[]> | Promise<AccessLog[]> | AccessLog[] {
        return Observable.of([]);
    }

    public getRecentLogs(lines: number): Observable<AccessLog[]> | Promise<AccessLog[]> | AccessLog[] {
        let url: string = this.config.logBaseEndpoint ? this.config.logBaseEndpoint : "";
        if (url === '') {
            url = '/api/logs';
        }

        return this.http.get(url+`?lines=${lines}`, HTTP_JSON_OPTIONS).toPromise()
            .then(response => response.json() as AccessLog[])
            .catch(error => Promise.reject(error));
    }
}