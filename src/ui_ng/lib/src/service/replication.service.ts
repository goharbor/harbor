import { Observable } from 'rxjs/Observable';
import { RequestQueryParams } from './RequestQueryParams';
import { ReplicationJob, ReplicationRule } from './interface';
import { Injectable } from "@angular/core";
import 'rxjs/add/observable/of';

/**
 * Define the service methods to handle the replication (rule and job) related things.
 * 
 * @export
 * @abstract
 * @class ReplicationService
 */
export abstract class ReplicationService {
    /**
     * Get the replication rules.
     * Set the argument 'projectId' to limit the data scope to the specified project;
     * set the argument 'ruleName' to return the rule only match the name pattern;
     * if pagination needed, use the queryParams to add query parameters.
     * 
     * @abstract
     * @param {(number | string)} [projectId]
     * @param {string} [ruleName]
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<ReplicationRule[]> | ReplicationRule[])}
     * 
     * @memberOf ReplicationService
     */
    abstract getReplicationRules(projectId?: number | string, ruleName?: string, queryParams?: RequestQueryParams): Observable<ReplicationRule[]> | ReplicationRule[];

    /**
     * Get the specified replication rule.
     * 
     * @abstract
     * @param {(number | string)} ruleId
     * @returns {(Observable<ReplicationRule> | ReplicationRule)}
     * 
     * @memberOf ReplicationService
     */
    abstract getReplicationRule(ruleId: number | string): Observable<ReplicationRule> | ReplicationRule;

    /**
     * Create new replication rule.
     * 
     * @abstract
     * @param {ReplicationRule} replicationRule
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf ReplicationService
     */
    abstract createReplicationRule(replicationRule: ReplicationRule): Observable<any> | any;

    /**
     * Update the specified replication rule.
     * 
     * @abstract
     * @param {ReplicationRule} replicationRule
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf ReplicationService
     */
    abstract updateReplicationRule(replicationRule: ReplicationRule): Observable<any> | any;

    /**
     * Delete the specified replication rule.
     * 
     * @abstract
     * @param {(number | string)} ruleId
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf ReplicationService
     */
    abstract deleteReplicationRule(ruleId: number | string): Observable<any> | any;

    /**
     * Enable the specified replication rule.
     * 
     * @abstract
     * @param {(number | string)} ruleId
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf ReplicationService
     */
    abstract enableReplicationRule(ruleId: number | string): Observable<any> | any;

    /**
     * Disable the specified replication rule.
     * 
     * @abstract
     * @param {(number | string)} ruleId
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf ReplicationService
     */
    abstract disableReplicationRule(ruleId: number | string): Observable<any> | any;

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
     * @param {(number | string)} ruleId
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<ReplicationJob> | ReplicationJob)}
     * 
     * @memberOf ReplicationService
     */
    abstract getJobs(ruleId: number | string, queryParams?: RequestQueryParams): Observable<ReplicationJob[]> | ReplicationJob[];

}

/**
 * Implement default service for replication rule and job.
 * 
 * @export
 * @class ReplicationDefaultService
 * @extends {ReplicationService}
 */
@Injectable()
export class ReplicationDefaultService extends ReplicationService {
    public getReplicationRules(projectId?: number | string, ruleName?: string, queryParams?: RequestQueryParams): Observable<ReplicationRule[]> | ReplicationRule[] {
        return Observable.of([]);
    }

    public getReplicationRule(ruleId: number | string): Observable<ReplicationRule> | ReplicationRule {
        return Observable.of({});
    }

    public createReplicationRule(replicationRule: ReplicationRule): Observable<any> | any {
        return Observable.of({});
    }

    public updateReplicationRule(replicationRule: ReplicationRule): Observable<any> | any {
        return Observable.of({});
    }

    public deleteReplicationRule(ruleId: number | string): Observable<any> | any {
        return Observable.of({});
    }

    public enableReplicationRule(ruleId: number | string): Observable<any> | any {
        return Observable.of({});
    }

    public disableReplicationRule(ruleId: number | string): Observable<any> | any {
        return Observable.of({});
    }

    public getJobs(ruleId: number | string, queryParams?: RequestQueryParams): Observable<ReplicationJob[]> | ReplicationJob[] {
        return Observable.of([]);
    }
}