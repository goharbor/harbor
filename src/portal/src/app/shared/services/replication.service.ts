import { HttpClient, HttpResponse } from '@angular/common/http';
import { Injectable } from '@angular/core';
import {
    buildHttpRequestOptions,
    HTTP_JSON_OPTIONS,
    HTTP_GET_OPTIONS,
    buildHttpRequestOptionsWithObserveResponse,
    HTTP_GET_OPTIONS_OBSERVE_RESPONSE,
    CURRENT_BASE_HREF,
} from '../units/utils';
import {
    ReplicationJob,
    ReplicationJobItem,
    ReplicationTasks,
} from './interface';
import { RequestQueryParams } from './RequestQueryParams';
import { map, catchError } from 'rxjs/operators';
import { Observable, throwError as observableThrowError } from 'rxjs';
import { ReplicationPolicy } from '../../../../ng-swagger-gen/models/replication-policy';

/**
 * Define the service methods to handle the replication (rule and job) related things.
 *
 **
 * @abstract
 * class ReplicationService
 */
export abstract class ReplicationService {
    /**
     * Get the replication rules.
     * Set the argument 'projectId' to limit the data scope to the specified project;
     * set the argument 'ruleName' to return the rule only match the name pattern;
     * if pagination needed, use the queryParams to add query parameters.
     *
     * @abstract
     *  ** deprecated param {(number | string)} [projectId]
     *  ** deprecated param {string} [ruleName]
     *  ** deprecated param {RequestQueryParams} [queryParams]
     * returns {(Observable<ReplicationRule[]>)}
     *
     * @memberOf ReplicationService
     */
    abstract getReplicationRules(
        projectId?: number | string,
        ruleName?: string,
        queryParams?: RequestQueryParams
    ): Observable<ReplicationPolicy[]>;
    abstract getReplicationRulesResponse(
        ruleName?: string,
        page?: number,
        pageSize?: number,
        queryParams?: RequestQueryParams
    ): Observable<HttpResponse<ReplicationPolicy[]>>;

    /**
     * Get the specified replication rule.
     *
     * @abstract
     *  ** deprecated param {(number | string)} ruleId
     * returns {(Observable<ReplicationRule>)}
     *
     * @memberOf ReplicationService
     */
    abstract getReplicationRule(
        ruleId: number | string
    ): Observable<ReplicationPolicy>;

    /**
     * Get the specified replication task.
     *
     * @abstract
     * returns {(Observable<ReplicationRule>)}
     *
     * @memberOf ReplicationService
     */
    abstract getReplicationTasks(
        executionId: number | string,
        queryParams?: RequestQueryParams
    ): Observable<ReplicationTasks>;
    /**
     * Create new replication rule.
     *
     * @abstract
     *  ** deprecated param {ReplicationRule} replicationRule
     * returns {(Observable<any>)}
     *
     * @memberOf ReplicationService
     */
    abstract createReplicationRule(
        replicationRule: ReplicationPolicy
    ): Observable<any>;

    /**
     * Update the specified replication rule.
     *
     * @abstract
     *  ** deprecated param {ReplicationRule} replicationRule
     * returns {(Observable<any>)}
     *
     * @memberOf ReplicationService
     */
    abstract updateReplicationRule(
        id: number,
        rep: ReplicationPolicy
    ): Observable<any>;

    /**
     * Delete the specified replication rule.
     *
     * @abstract
     *  ** deprecated param {(number | string)} ruleId
     * returns {(Observable<any>)}
     *
     * @memberOf ReplicationService
     */
    abstract deleteReplicationRule(ruleId: number | string): Observable<any>;

    /**
     * Enable the specified replication rule.
     *
     * @abstract
     *  ** deprecated param {(number | string)} ruleId
     * returns {(Observable<any>)}
     *
     * @memberOf ReplicationService
     */
    abstract enableReplicationRule(
        ruleId: number | string,
        enablement: number
    ): Observable<any>;

    /**
     * Disable the specified replication rule.
     *
     * @abstract
     *  ** deprecated param {(number | string)} ruleId
     * returns {(Observable<any>)}
     *
     * @memberOf ReplicationService
     */
    abstract disableReplicationRule(ruleId: number | string): Observable<any>;

    abstract replicateRule(ruleId: number | string): Observable<any>;

    abstract getRegistryInfo(id: number): Observable<any>;

    /**
     * Get the jobs for the specified replication rule.
     * Set query parameters through 'queryParams', support:
     *   - status
     *   - repository
     *   - startTime and endTime
     *   - page
     *   - pageSize
     *
     * @abstract
     *  ** deprecated param {(number | string)} ruleId
     *  ** deprecated param {RequestQueryParams} [queryParams]
     * returns {(Observable<ReplicationJob>)}
     *
     * @memberOf ReplicationService
     */
    abstract getExecutions(
        ruleId: number | string,
        queryParams?: RequestQueryParams
    ): Observable<ReplicationJob>;

    /**
     * Get the specified execution.
     *
     * @abstract
     *  ** deprecated param {(number | string)} endpointId
     * returns {(Observable<ReplicationJob> | ReplicationJob)}
     *
     * @memberOf ReplicationService
     */
    abstract getExecutionById(
        executionId: number | string
    ): Observable<ReplicationJob>;

    /**
     * Get the log of the specified job.
     *
     * @abstract
     *  ** deprecated param {(number | string)} jobId
     * returns {(Observable<string>)}
     * @memberof ReplicationService
     */
    abstract getJobLog(jobId: number | string): Observable<string>;

    abstract stopJobs(jobId: number | string): Observable<string>;

    abstract getJobBaseUrl(): string;
}

/**
 * Implement default service for replication rule and job.
 *
 **
 * class ReplicationDefaultService
 * extends {ReplicationService}
 */
@Injectable()
export class ReplicationDefaultService extends ReplicationService {
    _ruleBaseUrl: string;
    _replicateUrl: string;
    _baseUrl: string;

    constructor(private http: HttpClient) {
        super();
        this._ruleBaseUrl = CURRENT_BASE_HREF + '/replication/policies';
        this._replicateUrl = CURRENT_BASE_HREF + '/replication';
        this._baseUrl = CURRENT_BASE_HREF + '';
    }

    // Private methods
    // Check if the rule object is valid
    _isValidRule(rule: ReplicationPolicy): boolean {
        return (
            rule !== undefined &&
            rule != null &&
            rule.name !== undefined &&
            rule.name.trim() !== '' &&
            (!!rule.dest_registry || !!rule.src_registry)
        );
    }

    public getRegistryInfo(id): Observable<any> {
        let requestUrl: string = `${this._baseUrl}/registries/${id}/info`;
        return this.http
            .get(requestUrl)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public getJobBaseUrl() {
        return this._replicateUrl;
    }

    public getReplicationRules(
        projectId?: number | string,
        ruleName?: string,
        queryParams?: RequestQueryParams
    ): Observable<ReplicationPolicy[]> {
        if (!queryParams) {
            queryParams = new RequestQueryParams();
        }

        if (projectId) {
            queryParams = queryParams.set('project_id', '' + projectId);
        }

        if (ruleName) {
            queryParams = queryParams.set('name', ruleName);
        }
        return this.http
            .get(this._ruleBaseUrl, buildHttpRequestOptions(queryParams))
            .pipe(
                map(response => response as ReplicationPolicy[]),
                catchError(error => observableThrowError(error))
            );
    }
    public getReplicationRulesResponse(
        ruleName?: string,
        page?: number,
        pageSize?: number,
        queryParams?: RequestQueryParams
    ): Observable<HttpResponse<ReplicationPolicy[]>> {
        if (!queryParams) {
            queryParams = new RequestQueryParams();
        }
        if (ruleName) {
            queryParams = queryParams.set('name', ruleName);
        }
        if (page) {
            queryParams = queryParams.set('page', page + '');
        }
        if (pageSize) {
            queryParams = queryParams.set('page_size', pageSize + '');
        }
        return this.http
            .get<HttpResponse<ReplicationPolicy[]>>(
                this._ruleBaseUrl,
                buildHttpRequestOptionsWithObserveResponse(queryParams)
            )
            .pipe(
                map(response => response as HttpResponse<ReplicationPolicy[]>),
                catchError(error => observableThrowError(error))
            );
    }

    public getReplicationRule(
        ruleId: number | string
    ): Observable<ReplicationPolicy> {
        if (!ruleId) {
            return observableThrowError('Bad argument');
        }

        let url: string = `${this._ruleBaseUrl}/${ruleId}`;
        return this.http.get(url, HTTP_GET_OPTIONS).pipe(
            map(response => response as ReplicationPolicy),
            catchError(error => observableThrowError(error))
        );
    }

    public getReplicationTasks(
        executionId: number | string,
        queryParams?: RequestQueryParams
    ): Observable<ReplicationTasks> {
        if (!executionId) {
            return observableThrowError('Bad argument');
        }
        let url: string = `${this._replicateUrl}/executions/${executionId}/tasks`;
        return this.http
            .get(url, buildHttpRequestOptionsWithObserveResponse(queryParams))
            .pipe(
                map(response => response as ReplicationTasks),
                catchError(error => observableThrowError(error))
            );
    }

    public createReplicationRule(
        replicationRule: ReplicationPolicy
    ): Observable<any> {
        if (!this._isValidRule(replicationRule)) {
            return observableThrowError('Bad argument');
        }

        return this.http
            .post(
                this._ruleBaseUrl,
                JSON.stringify(replicationRule),
                HTTP_JSON_OPTIONS
            )
            .pipe(
                map(response => response),
                catchError(error => observableThrowError(error))
            );
    }

    public updateReplicationRule(
        id: number,
        rep: ReplicationPolicy
    ): Observable<any> {
        if (!this._isValidRule(rep)) {
            return observableThrowError('Bad argument');
        }

        let url = `${this._ruleBaseUrl}/${id}`;
        return this.http.put(url, JSON.stringify(rep), HTTP_JSON_OPTIONS).pipe(
            map(response => response),
            catchError(error => observableThrowError(error))
        );
    }

    public deleteReplicationRule(ruleId: number | string): Observable<any> {
        if (!ruleId || ruleId <= 0) {
            return observableThrowError('Bad argument');
        }

        let url: string = `${this._ruleBaseUrl}/${ruleId}`;
        return this.http.delete(url, HTTP_JSON_OPTIONS).pipe(
            map(response => response),
            catchError(error => observableThrowError(error))
        );
    }

    public replicateRule(ruleId: number | string): Observable<any> {
        if (!ruleId) {
            return observableThrowError('Bad argument');
        }

        let url: string = `${this._replicateUrl}/executions`;
        return this.http
            .post(url, { policy_id: ruleId }, HTTP_JSON_OPTIONS)
            .pipe(
                map(response => response),
                catchError(error => observableThrowError(error))
            );
    }

    public enableReplicationRule(
        ruleId: number | string,
        enablement: number
    ): Observable<any> {
        if (!ruleId || ruleId <= 0) {
            return observableThrowError('Bad argument');
        }

        let url: string = `${this._ruleBaseUrl}/${ruleId}/enablement`;
        return this.http
            .put(url, { enabled: enablement }, HTTP_JSON_OPTIONS)
            .pipe(
                map(response => response),
                catchError(error => observableThrowError(error))
            );
    }

    public disableReplicationRule(ruleId: number | string): Observable<any> {
        if (!ruleId || ruleId <= 0) {
            return observableThrowError('Bad argument');
        }

        let url: string = `${this._ruleBaseUrl}/${ruleId}/enablement`;
        return this.http.put(url, { enabled: 0 }, HTTP_JSON_OPTIONS).pipe(
            map(response => response),
            catchError(error => observableThrowError(error))
        );
    }

    public getExecutions(
        ruleId: number | string,
        queryParams?: RequestQueryParams
    ): Observable<ReplicationJob> {
        if (!ruleId || ruleId <= 0) {
            return observableThrowError('Bad argument');
        }

        if (!queryParams) {
            queryParams = new RequestQueryParams();
        }
        let url: string = `${this._replicateUrl}/executions`;
        queryParams = queryParams.set('policy_id', '' + ruleId);
        return this.http
            .get<HttpResponse<ReplicationJobItem[]>>(
                url,
                buildHttpRequestOptionsWithObserveResponse(queryParams)
            )
            .pipe(
                map(response => {
                    let result: ReplicationJob = {
                        metadata: {
                            xTotalCount: 0,
                        },
                        data: [],
                    };

                    if (response && response.headers) {
                        let xHeader: string =
                            response.headers.get('X-Total-Count');
                        if (xHeader) {
                            result.metadata.xTotalCount = parseInt(xHeader, 0);
                        }
                    }
                    result.data = response.body as ReplicationJobItem[];
                    if (result.metadata.xTotalCount === 0) {
                        if (result.data && result.data.length > 0) {
                            result.metadata.xTotalCount = result.data.length;
                        }
                    }

                    return result;
                }),
                catchError(error => observableThrowError(error))
            );
    }

    public getExecutionById(
        executionId: number | string
    ): Observable<ReplicationJob> {
        if (!executionId || executionId <= 0) {
            return observableThrowError('Bad request argument.');
        }
        let requestUrl: string = `${this._replicateUrl}/executions/${executionId}`;
        return this.http
            .get<HttpResponse<ReplicationJobItem[]>>(
                requestUrl,
                HTTP_GET_OPTIONS_OBSERVE_RESPONSE
            )
            .pipe(
                map(response => {
                    let result: ReplicationJob = {
                        metadata: {
                            xTotalCount: 0,
                        },
                        data: [],
                    };

                    if (response && response.headers) {
                        let xHeader: string =
                            response.headers.get('X-Total-Count');
                        if (xHeader) {
                            result.metadata.xTotalCount = parseInt(xHeader, 0);
                        }
                    }
                    result.data = response.body as ReplicationJobItem[];
                    if (result.metadata.xTotalCount === 0) {
                        if (result.data && result.data.length > 0) {
                            result.metadata.xTotalCount = result.data.length;
                        }
                    }

                    return result;
                }),
                catchError(error => observableThrowError(error))
            );
    }

    public getJobLog(jobId: number | string): Observable<string> {
        if (!jobId || jobId <= 0) {
            return observableThrowError('Bad argument');
        }

        let logUrl = `${this._replicateUrl}/${jobId}/log`;
        return this.http
            .get<string>(logUrl, HTTP_GET_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public stopJobs(jobId: number | string): Observable<any> {
        if (!jobId || jobId <= 0) {
            return observableThrowError('Bad request argument.');
        }
        let requestUrl: string = `${this._replicateUrl}/executions/${jobId}`;

        return this.http.put(requestUrl, HTTP_JSON_OPTIONS).pipe(
            map(response => response),
            catchError(error => observableThrowError(error))
        );
    }
}
