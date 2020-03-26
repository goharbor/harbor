import { HttpClient, HttpResponse, HttpParams } from "@angular/common/http";
import { Injectable, Inject } from "@angular/core";
import {
  HTTP_JSON_OPTIONS,
  buildHttpRequestOptionsWithObserveResponse,
} from "../utils/utils";
import {
  QuotaHard
} from "./interface";
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
import { Quota } from "./interface";
import { SERVICE_CONFIG, IServiceConfig } from "../entities/service.config";
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
  abstract getQuotaList(quotaType, page?, pageSize?, sortBy?: any):
    any;

  abstract updateQuota(
    id: number,
    rep: QuotaHard
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
  quotaUrl: string;
  constructor(
    private http: HttpClient,
    @Inject(SERVICE_CONFIG) private config: IServiceConfig
  ) {
    super();
    if (this.config && this.config.quotaUrl) {
      this.quotaUrl = this.config.quotaUrl;
    }
  }

  public getQuotaList(quotaType: string, page?, pageSize?, sortBy?: any):
    any {

    let params = new HttpParams();
    if (quotaType) {
      params = params.set('reference', quotaType);
    }
    if (page && pageSize) {
      params = params.set('page', page + '').set('page_size', pageSize + '');
    }
    if (sortBy) {
      params = params.set('sort', sortBy);
    }

    return this.http
      .get<HttpResponse<Quota[]>>(this.quotaUrl
        , buildHttpRequestOptionsWithObserveResponse(params))
      .pipe(map(response => {
        return response;
      })
        , catchError(error => observableThrowError(error)));
  }

  public updateQuota(
    id: number,
    quotaHardLimit: QuotaHard
  ): Observable<any> {

    let url = `${this.quotaUrl}/${id}`;
    return this.http
      .put(url, quotaHardLimit, HTTP_JSON_OPTIONS)
      .pipe(catchError(error => observableThrowError(error)));
  }

}
