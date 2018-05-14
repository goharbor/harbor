import { Injectable, Inject } from "@angular/core";
import { Http } from "@angular/http";
import { Observable } from "rxjs/Observable";
import "rxjs/add/observable/of";

import { IServiceConfig, SERVICE_CONFIG } from "../service.config";
import {
  buildHttpRequestOptions,
  HTTP_JSON_OPTIONS,
  HTTP_GET_OPTIONS
} from "../utils";
import { RequestQueryParams } from "./RequestQueryParams";
import { Endpoint, ReplicationRule } from "./interface";

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
  abstract getEndpoints(
    endpointName?: string,
    queryParams?: RequestQueryParams
  ): Observable<Endpoint[]> | Promise<Endpoint[]> | Endpoint[];

  /**
   * Get the specified endpoint.
   *
   * @abstract
   * @param {(number | string)} endpointId
   * @returns {(Observable<Endpoint> | Endpoint)}
   *
   * @memberOf EndpointService
   */
  abstract getEndpoint(
    endpointId: number | string
  ): Observable<Endpoint> | Promise<Endpoint> | Endpoint;

  /**
   * Create new endpoint.
   *
   * @abstract
   * @param {Endpoint} endpoint
   * @returns {(Observable<any> | any)}
   *
   * @memberOf EndpointService
   */
  abstract createEndpoint(
    endpoint: Endpoint
  ): Observable<any> | Promise<any> | any;

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
  abstract updateEndpoint(
    endpointId: number | string,
    endpoint: Endpoint
  ): Observable<any> | Promise<any> | any;

  /**
   * Delete the specified endpoint.
   *
   * @abstract
   * @param {(number | string)} endpointId
   * @returns {(Observable<any> | any)}
   *
   * @memberOf EndpointService
   */
  abstract deleteEndpoint(
    endpointId: number | string
  ): Observable<any> | Promise<any> | any;

  /**
   * Ping the specified endpoint.
   *
   * @abstract
   * @param {Endpoint} endpoint
   * @returns {(Observable<any> | any)}
   *
   * @memberOf EndpointService
   */
  abstract pingEndpoint(
    endpoint: Endpoint
  ): Observable<any> | Promise<any> | any;

  /**
   * Check endpoint whether in used with specific replication rule.
   *
   * @abstract
   * @param {{number | string}} endpointId
   * @returns {{Observable<any> | any}}
   */
  abstract getEndpointWithReplicationRules(
    endpointId: number | string
  ): Observable<any> | Promise<any> | any;
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
  _endpointUrl: string;

  constructor(
    @Inject(SERVICE_CONFIG) config: IServiceConfig,
    private http: Http
  ) {
    super();
    this._endpointUrl = config.targetBaseEndpoint
      ? config.targetBaseEndpoint
      : "/api/targets";
  }

  public getEndpoints(
    endpointName?: string,
    queryParams?: RequestQueryParams
  ): Observable<Endpoint[]> | Promise<Endpoint[]> | Endpoint[] {
    if (!queryParams) {
      queryParams = new RequestQueryParams();
    }
    if (endpointName) {
      queryParams.set("name", endpointName);
    }
    let requestUrl: string = `${this._endpointUrl}`;
    return this.http
      .get(requestUrl, buildHttpRequestOptions(queryParams))
      .toPromise()
      .then(response => response.json())
      .catch(error => Promise.reject(error));
  }

  public getEndpoint(
    endpointId: number | string
  ): Observable<Endpoint> | Promise<Endpoint> | Endpoint {
    if (!endpointId || endpointId <= 0) {
      return Promise.reject("Bad request argument.");
    }
    let requestUrl: string = `${this._endpointUrl}/${endpointId}`;
    return this.http
      .get(requestUrl, HTTP_GET_OPTIONS)
      .toPromise()
      .then(response => response.json() as Endpoint)
      .catch(error => Promise.reject(error));
  }

  public createEndpoint(
    endpoint: Endpoint
  ): Observable<any> | Promise<any> | any {
    if (!endpoint) {
      return Promise.reject("Invalid endpoint.");
    }
    let requestUrl: string = `${this._endpointUrl}`;
    return this.http
      .post(requestUrl, JSON.stringify(endpoint), HTTP_JSON_OPTIONS)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }

  public updateEndpoint(
    endpointId: number | string,
    endpoint: Endpoint
  ): Observable<any> | Promise<any> | any {
    if (!endpointId || endpointId <= 0) {
      return Promise.reject("Bad request argument.");
    }
    if (!endpoint) {
      return Promise.reject("Invalid endpoint.");
    }
    let requestUrl: string = `${this._endpointUrl}/${endpointId}`;
    return this.http
      .put(requestUrl, JSON.stringify(endpoint), HTTP_JSON_OPTIONS)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }

  public deleteEndpoint(
    endpointId: number | string
  ): Observable<any> | Promise<any> | any {
    if (!endpointId || endpointId <= 0) {
      return Promise.reject("Bad request argument.");
    }
    let requestUrl: string = `${this._endpointUrl}/${endpointId}`;
    return this.http
      .delete(requestUrl)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }

  public pingEndpoint(
    endpoint: Endpoint
  ): Observable<any> | Promise<any> | any {
    if (!endpoint) {
      return Promise.reject("Invalid endpoint.");
    }
    let requestUrl: string = `${this._endpointUrl}/ping`;
    return this.http
      .post(requestUrl, endpoint, HTTP_JSON_OPTIONS)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }

  public getEndpointWithReplicationRules(
    endpointId: number | string
  ): Observable<any> | Promise<any> | any {
    if (!endpointId || endpointId <= 0) {
      return Promise.reject("Bad request argument.");
    }
    let requestUrl: string = `${this._endpointUrl}/${endpointId}/policies`;
    return this.http
      .get(requestUrl, HTTP_GET_OPTIONS)
      .toPromise()
      .then(response => response.json() as ReplicationRule[])
      .catch(error => Promise.reject(error));
  }
}
