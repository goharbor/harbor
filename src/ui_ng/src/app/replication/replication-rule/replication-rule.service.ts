/**
 * Created by pengf on 12/5/2017.
 */

import {Injectable} from "@angular/core";
import {Http, RequestOptions, Headers, URLSearchParams} from "@angular/http";
import {Observable} from "rxjs/Observable";
import {ReplicationRule, Target} from "./replication-rule";
import {HTTP_GET_OPTIONS, HTTP_JSON_OPTIONS} from "../../shared/shared.utils";
import {Project} from "../../project/project";

@Injectable()
export class ReplicationRuleServie {
    headers = new Headers({'Content-type': 'application/json'});
    options = new RequestOptions({'headers': this.headers});
    baseurl =  '/api/policies/replication';
    targetUrl= '/api/targets';

    constructor(private http: Http) {}

    public createReplicationRule(replicationRule: ReplicationRule): Observable<any> | Promise<any> | any {
        /*if (!this._isValidRule(replicationRule)) {
            return Promise.reject('Bad argument');
        }*/

        return this.http.post(this.baseurl, JSON.stringify(replicationRule), this.options).toPromise()
            .then(response => response)
            .catch(error => Promise.reject(error));
    }

    public getReplicationRules(projectId?: number | string, ruleName?: string): Promise<ReplicationRule[]> | ReplicationRule[] {
        let queryParams = new URLSearchParams();
        if (projectId) {
            queryParams.set('project_id', '' + projectId);
        }

        if (ruleName) {
            queryParams.set('name', ruleName);
        }

        return this.http.get(this.baseurl, {search: queryParams}).toPromise()
            .then(response => response.json() as ReplicationRule[])
            .catch(error => Promise.reject(error));
    }

    public getReplicationRule(policyId: number): Promise<ReplicationRule> {
        let url: string = `${this.baseurl}/${policyId}`;
        return this.http.get(url, HTTP_GET_OPTIONS).toPromise()
            .then(response => response.json() as ReplicationRule)
            .catch(error => Promise.reject(error));
    }


    public getEndpoints(): Promise<Target[]> | Target[] {
        return this.http
            .get(this.targetUrl)
            .toPromise()
            .then(response => response.json())
            .catch(error => Promise.reject(error));
    }

    public listProjects(): Promise<Project[]> | Project[] {
        return this.http.get(`/api/projects`, HTTP_GET_OPTIONS).toPromise()
            .then(response => response.json())
            .catch(error => Promise.reject(error));
    }

    public updateReplicationRule(id: number, rep: {[key: string]: any | any[] }): Observable<any> | Promise<any> | any {
        let url: string = `${this.baseurl}/${id}`;
        return this.http.put(url, JSON.stringify(rep), HTTP_JSON_OPTIONS).toPromise()
            .then(response => response)
            .catch(error => Promise.reject(error));
    }

    public updateEndpoint(endpointId: number | string, endpoint: any): Promise<any> | any {
        if (!endpointId || endpointId <= 0) {
            return Promise.reject('Bad request argument.');
        }
        if (!endpoint) {
            return Promise.reject('Invalid endpoint.');
        }
        let requestUrl: string = `/api/targets/${endpointId}`;
        return this.http
            .put(requestUrl, JSON.stringify(endpoint), HTTP_JSON_OPTIONS)
            .toPromise()
            .then(response=>response.status)
            .catch(error=>Promise.reject(error));
    }

    public pingEndpoint(endpoint: any): Promise<any> | any {
        if (!endpoint) {
            return Promise.reject('Invalid endpoint.');
        }
        let requestUrl: string = `/api/targets/ping`;
        return this.http
            .post(requestUrl, endpoint, HTTP_JSON_OPTIONS)
            .toPromise()
            .then(response=>response.status)
            .catch(error=>Promise.reject(error));
    }

}
