import { Observable } from 'rxjs/Observable';
import { RequestQueryParams } from './RequestQueryParams';
import { Endpoint } from './interface';
import { Injectable } from "@angular/core";
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
    public getEndpoints(endpointName?: string, queryParams?: RequestQueryParams): Observable<Endpoint[]> | Promise<Endpoint[]> | Endpoint[] {
        return Observable.of([]);
    }

    public getEndpoint(endpointId: number | string): Observable<Endpoint> | Promise<Endpoint> | Endpoint {
        return Observable.of({});
    }

    public createEndpoint(endpoint: Endpoint): Observable<any> | Promise<any> | any {
        return Observable.of({});
    }

    public updateEndpoint(endpointId: number | string, endpoint: Endpoint): Observable<any> | Promise<any> | any {
        return Observable.of({});
    }

    public deleteEndpoint(endpointId: number | string): Observable<any> | Promise<any> | any {
        return Observable.of({});
    }

    public pingEndpoint(endpoint: Endpoint): Observable<any> | Promise<any> | any {
        return Observable.of({});
    }
}

