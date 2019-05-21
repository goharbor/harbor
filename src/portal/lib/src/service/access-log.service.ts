import { Observable,  of, throwError as observableThrowError } from "rxjs";
import { RequestQueryParams } from "./RequestQueryParams";
import { AccessLog, AccessLogItem } from "./interface";
import { Injectable, Inject } from "@angular/core";
import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import { HttpClient, HttpResponse } from "@angular/common/http";
import { buildHttpRequestOptionsWithObserveResponse, HTTP_GET_OPTIONS_OBSERVE_RESPONSE } from "../utils";
import { map, catchError } from "rxjs/operators";

/**
 * Define service methods to handle the access log related things.
 *
 **
 * @abstract
 * class AccessLogService
 */
export abstract class AccessLogService {
  /**
   * Get the audit logs for the specified project.
   * Set query parameters through 'queryParams', support:
   *  - page
   *  - pageSize
   *
   * @abstract
   *  ** deprecated param {(number | string)} projectId
   *  ** deprecated param {RequestQueryParams} [queryParams]
   * returns {(Observable<AccessLog> | AccessLog)}
   *
   * @memberOf AccessLogService
   */
  abstract getAuditLogs(
    projectId: number | string,
    queryParams?: RequestQueryParams
  ): Observable<AccessLog>;

  /**
   * Get the recent logs.
   *
   * @abstract
   *  ** deprecated param {RequestQueryParams} [queryParams]
   * returns {(Observable<AccessLog> | AccessLog)}
   *
   * @memberOf AccessLogService
   */
  abstract getRecentLogs(
    queryParams?: RequestQueryParams
  ): Observable<AccessLog>;
}

/**
 * Implement a default service for access log.
 *
 **
 * class AccessLogDefaultService
 * extends {AccessLogService}
 */
@Injectable()
export class AccessLogDefaultService extends AccessLogService {
  constructor(
    private http: HttpClient,
    @Inject(SERVICE_CONFIG) private config: IServiceConfig
  ) {
    super();
  }

  public getAuditLogs(
    projectId: number | string,
    queryParams?: RequestQueryParams
  ): Observable<AccessLog> {
    return of({} as AccessLog);
  }

  public getRecentLogs(
    queryParams?: RequestQueryParams
  ): Observable<AccessLog> {
    let url: string = this.config.logBaseEndpoint
      ? this.config.logBaseEndpoint
      : "";
    if (url === "") {
      url = "/api/logs";
    }

    return this.http
      .get<HttpResponse<AccessLogItem[]>>(
        url,
        queryParams ? buildHttpRequestOptionsWithObserveResponse(queryParams) : HTTP_GET_OPTIONS_OBSERVE_RESPONSE
      )
      .pipe(map(response => {
        let result: AccessLog = {
          metadata: {
            xTotalCount: 0
          },
          data: []
        };
        let xHeader: string | null = "0";
        if (response && response.headers) {
          xHeader = response.headers.get("X-Total-Count");
        }

        if (result && result.metadata) {
          result.metadata.xTotalCount = parseInt(xHeader ? xHeader : "0", 0);
          if (result.metadata.xTotalCount > 0) {
            result.data = response.body as AccessLogItem[];
          }
        }

        return result;
      })
      , catchError(error => observableThrowError(error)));
  }
}
