import { Injectable } from '@angular/core';
import { HttpClient, HttpResponse } from '@angular/common/http';
import { throwError as observableThrowError, Observable, pipe } from 'rxjs';
import {
  buildHttpRequestOptionsWithObserveResponse,
  HTTP_JSON_OPTIONS,
  HTTP_GET_OPTIONS_OBSERVE_RESPONSE
} from '../../lib/utils/utils';
import { RequestQueryParams } from '../../lib/services';
import { DistributionHistory, QueryParam } from './distribution-interface';
import {
  DistributionProvider,
  DistributionInstance
} from './distribution-interface';
import { HttpParams } from '@angular/common/http';
import { catchError, map } from 'rxjs/operators';

@Injectable()
export class DistributionService {
  providersEndpoint = '/api/v2.0/distribution/providers';
  instanceEndpoint = '/api/v2.0/distribution/instances';
  preheatEndpoint = '/api/v2.0/distribution/preheats';

  constructor(private http: HttpClient) {}

  getDistributionHistories(
    queryParam?: QueryParam
  ): Observable<HttpResponse<DistributionHistory[]>> {
    let params: HttpParams = new HttpParams();
    if (queryParam.query) {
      params = params.set('q', queryParam.query);
    }
    if (queryParam.page) {
      params = params.set('page', queryParam.page.toString());
    }
    if (queryParam.pageSize) {
      params = params.set('page_size', queryParam.pageSize.toString());
    }

    return this.http
      .get<HttpResponse<DistributionHistory[]>>(
        this.preheatEndpoint,
        params
          ? buildHttpRequestOptionsWithObserveResponse(params)
          : HTTP_JSON_OPTIONS
      )
      .pipe(catchError(error => observableThrowError(error)));
  }

  preheatImages(images: string[]): Observable<any> {
    return this.http
      .post(this.preheatEndpoint, { images: images }, HTTP_JSON_OPTIONS)
      .pipe(
        map(response => {
          return response as any;
        })
      )
      .pipe(catchError(error => observableThrowError(error)));
  }

  getInstances(
    queryParam?: QueryParam
  ): Observable<HttpResponse<DistributionInstance[]>> {
    let params: HttpParams = new HttpParams();
    if (queryParam.query) {
      params = params.set('q', queryParam.query);
    }
    if (queryParam.page) {
      params = params.set('page', queryParam.page.toString());
    }
    if (queryParam.pageSize) {
      params = params.set('page_size', queryParam.pageSize.toString());
    }
    return this.http
      .get<HttpResponse<DistributionInstance[]>>(
        this.instanceEndpoint,
        params
          ? buildHttpRequestOptionsWithObserveResponse(params)
          : HTTP_GET_OPTIONS_OBSERVE_RESPONSE
      )
      .pipe(catchError(error => observableThrowError(error)));
  }

  createInstance(instance: DistributionInstance): Observable<any> {
    return this.http
      .post(this.instanceEndpoint, instance, HTTP_JSON_OPTIONS)
      .pipe(map(response => response as any))
      .pipe(catchError(error => observableThrowError(error)));
  }

  updateInstance(instance: DistributionInstance): Observable<any> {
    let data: DistributionInstance = {
      endpoint: instance.endpoint,
      enabled: instance.enabled,
      description: instance.description,
      auth_mode: instance.auth_mode,
      auth_data: instance.auth_data
    };
    return this.http
      .put(`${this.instanceEndpoint}/${instance.id}`, data, HTTP_JSON_OPTIONS)
      .pipe(map(response => response as any))
      .pipe(catchError(error => observableThrowError(error)));
  }

  deleteInstance(instance: DistributionInstance): Observable<any> {
    return this.http
      .delete(`${this.instanceEndpoint}/${instance.id}`, HTTP_JSON_OPTIONS)
      .pipe(map(response => response as any))
      .pipe(catchError(error => observableThrowError(error)));
  }

  getProviderDrivers(
    params?: RequestQueryParams
  ): Observable<DistributionProvider[]> {
    return this.http
      .get(this.providersEndpoint, HTTP_JSON_OPTIONS)
      .pipe(map(response => response as DistributionProvider[]))
      .pipe(catchError(error => observableThrowError(error)));
  }
}
