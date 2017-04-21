import { Observable } from 'rxjs/Observable';
import { RequestQueryParams } from './RequestQueryParams';
import { AccessLog } from './interface';
import { Injectable } from "@angular/core";
import 'rxjs/add/observable/of';

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
    abstract getAuditLogs(projectId: number | string, queryParams?: RequestQueryParams): Observable<AccessLog[]> | AccessLog[];

    /**
     * Get the recent logs.
     * 
     * @abstract
     * @param {number} lines : Specify how many lines should be returned.
     * @returns {(Observable<AccessLog[]> | AccessLog[])}
     * 
     * @memberOf AccessLogService
     */
    abstract getRecentLogs(lines: number): Observable<AccessLog[]> | AccessLog[];
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
    public getAuditLogs(projectId: number | string, queryParams?: RequestQueryParams): Observable<AccessLog[]> | AccessLog[] {
        return Observable.of([]);
    }

    public getRecentLogs(lines: number): Observable<AccessLog[]> | AccessLog[] {
        return Observable.of([]);
    }
}