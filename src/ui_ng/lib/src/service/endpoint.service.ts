import { Observable } from 'rxjs/Observable';
import { RequestQueryParams } from './RequestQueryParams';
import { Endpoint, ReplicationRule } from './interface';
import { Injectable } from "@angular/core";
import { Http } from '@angular/http';
import 'rxjs/add/observable/of';

/**
 * Define the service methods to handle the endpoint related things.
 * 
 * @export
 * @abstract
 * @class EndpointService
 */
export abstract class EndpointService {
    /**
     * Get all the endpoints.
     * Set the argument 'endpointName' to return only the endpoints match the name pattern.
     * 
     * @abstract
     * @param {string} [endpointName]
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<Endpoint[]> | Endpoint[])}
     * 
     * @memberOf EndpointService
     */
    abstract getEndpoints(endpointName?: string, queryParams?: RequestQueryParams): Observable<Endpoint[]> | Promise<Endpoint[]> | Endpoint[];

    /**
     * Get the specified endpoint.
     * 
     * @abstract
     * @param {(number | string)} endpointId
     * @returns {(Observable<Endpoint> | Endpoint)}
     * 
     * @memberOf EndpointService
     */
    abstract getEndpoint(endpointId: number | string): Observable<Endpoint> | Promise<Endpoint> | Endpoint;

    /**
     * Create new endpoint.
     * 
     * @abstract
     * @param {Endpoint} endpoint
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    abstract createEndpoint(endpoint: Endpoint): Observable<any> | Promise<any> | any;

    /**
     * Update the specified endpoint.
     * 
     * @abstract
     * @param {(number | string)} endpointId
     * @param {Endpoint} endpoint
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    abstract updateEndpoint(endpointId: number | string, endpoint: Endpoint): Observable<any> | Promise<any> | any;

    /**
     * Delete the specified endpoint.
     * 
     * @abstract
     * @param {(number | string)} endpointId
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    abstract deleteEndpoint(endpointId: number | string): Observable<any> | Promise<any> | any;

    /**
     * Ping the specified endpoint.
     * 
     * @abstract
     * @param {Endpoint} endpoint
     * @returns {(Observable<any> | any)}
     * 
     * @memberOf EndpointService
     */
    abstract pingEndpoint(endpoint: Endpoint): Observable<any> | Promise<any> | any;

    /**
     * Check endpoint whether in used with specific replication rule.
     * 
     * @abstract 
     * @param {{number | string}} endpointId
     * @returns {{Observable<any> | any}}
     */
    abstract getEndpointWithReplicationRules(endpointId: number | string): Observable<any> | Promise<any> | any;
}

/**
 * Implement default service for endpoint.
 * 
 * @export
 * @class EndpointDefaultService
 * @extends {EndpointService}
 */
@Injectable()
export class EndpointDefaultService extends EndpointService {
    
    constructor(private http: Http){
      super();
    }

    public getEndpoints(endpointName?: string, queryParams?: RequestQueryParams): Observable<Endpoint[]> | Promise<Endpoint[]> | Endpoint[] {
        return this.http
               .get(`/api/targets?name=${endpointName}`)
               .toPromise()
               .then(response=>response.json())
               .catch(error=>Promise.reject(error));
    }

    public getEndpoint(endpointId: number | string): Observable<Endpoint> | Promise<Endpoint> | Endpoint {
        return this.http
               .get(`/api/targets/${endpointId}`)
               .toPromise()
               .then(response=>response.json() as Endpoint)
               .catch(error=>Promise.reject(error));
    }

    public createEndpoint(endpoint: Endpoint): Observable<any> | Promise<any> | any {
        return this.http
               .post(`/api/targets`, JSON.stringify(endpoint))
               .toPromise()
               .then(response=>response.status)
               .catch(error=>Promise.reject(error));
    }

    public updateEndpoint(endpointId: number | string, endpoint: Endpoint): Observable<any> | Promise<any> | any {
        return this.http
               .put(`/api/targets/${endpointId}`, JSON.stringify(endpoint))
               .toPromise()
               .then(response=>response.status)
               .catch(error=>Promise.reject(error));
    }

    public deleteEndpoint(endpointId: number | string): Observable<any> | Promise<any> | any {
        return this.http
               .delete(`/api/targets/${endpointId}`)
               .toPromise()
               .then(response=>response.status)
               .catch(error=>Promise.reject(error));
    }

    public pingEndpoint(endpoint: Endpoint): Observable<any> | Promise<any> | any {
        if(endpoint.id) {
          return this.http
                 .post(`/api/targets/${endpoint.id}/ping`, {})
                 .toPromise()
                 .then(response=>response.status)
                 .catch(error=>Promise.reject(error));
        }
        return this.http
               .post(`/api/targets/ping`, endpoint)
               .toPromise()
               .then(response=>response.status)
               .catch(error=>Observable.throw(error));
    }

    public getEndpointWithReplicationRules(endpointId: number | string): Observable<any> | Promise<any> | any {
        return this.http
               .get(`/api/targets/${endpointId}/policies`)
               .toPromise()
               .then(response=>response.json() as ReplicationRule[])
               .catch(error=>Promise.reject(error));
    }
}