import { HttpClient, HttpResponse, HttpParams } from "@angular/common/http";
import { Injectable, Inject } from "@angular/core";
import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import {
  buildHttpRequestOptions,
  HTTP_JSON_OPTIONS,
  HTTP_GET_OPTIONS,
  buildHttpRequestOptionsWithObserveResponse,
  HTTP_GET_OPTIONS_OBSERVE_RESPONSE
} from "../utils";
import {
  ReplicationJob,
  ReplicationRule,
  ReplicationJobItem,
  ReplicationTasks,
  QuotaSpec
} from "./interface";
import { RequestQueryParams } from "./RequestQueryParams";
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
import { Quota } from "./interface";
/**
 * Define the service methods to handle the replication (rule and job) related things.
 *
 **
 * @abstract
 * class QuotaService
 */
export abstract class QuotaService {
  /**
   *
   * @abstract
   * returns {(Observable<ReplicationRule[]>)}
   *
   * @memberOf QuotaService
   */
  abstract getQuotaList(page?, pageSize?, sortBy?: any):
    any;

  abstract updateQuota(
    id: number,
    rep: QuotaSpec
  ): Observable<any>;
}

/**
 * Implement default service for replication rule and job.
 *
 **
 * class QuotaDefaultService
 * extends {QuotaService}
 */
@Injectable()
export class QuotaDefaultService extends QuotaService {

  constructor(
    private http: HttpClient
  ) {
    super();
  }

  public getQuotaList(page?, pageSize?, sortBy?: any):
    any {
    const sortByFiled = sortBy ? `&sort=${sortBy}` : '';

    let params = new HttpParams();
    if (page && pageSize) {
      params = params.set('page', page + '').set('page_size', pageSize + '');
    }



    return this.http
      .get<HttpResponse<Quota[]>>(`/api/quotas?reference=project${sortByFiled}`, buildHttpRequestOptionsWithObserveResponse(params))
      .pipe(map(response => {
        return response
      })
        , catchError(error => observableThrowError(error)));
  }


  public updateQuota(
    id: number,
    quotaHardLimit: QuotaSpec
  ): Observable<any> {

    let url = `/api/quotas/${id}`;
    return this.http
      .put(url, { spec: quotaHardLimit }, HTTP_JSON_OPTIONS)
      .pipe(map(response => response)
        , catchError(error => observableThrowError(error)));
  }



}
