import { Observable } from "rxjs/Observable";
import { RequestQueryParams } from "./RequestQueryParams";
import { AccessLog, AccessLogItem } from "./interface";
import { Injectable, Inject } from "@angular/core";
import "rxjs/add/observable/of";
import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import { Http } from "@angular/http";
import { buildHttpRequestOptions, HTTP_GET_OPTIONS } from "../utils";

/**
 * Define service methods to handle the access log related things.
 *
 * @export
 * @abstract
 * @class AccessLogService
 */
export abstract class AccessLogService {
  /**
   * Get the audit logs for the specified project.
   * Set query parameters through 'queryParams', support:
   *  - page
   *  - pageSize
   *
   * @abstract
   * @param {(number | string)} projectId
   * @param {RequestQueryParams} [queryParams]
   * @returns {(Observable<AccessLog> | Promise<AccessLog> | AccessLog)}
   *
   * @memberOf AccessLogService
   */
  abstract getAuditLogs(
    projectId: number | string,
    queryParams?: RequestQueryParams
  ): Observable<AccessLog> | Promise<AccessLog> | AccessLog;

  /**
   * Get the recent logs.
   *
   * @abstract
   * @param {RequestQueryParams} [queryParams]
   * @returns {(Observable<AccessLog> | Promise<AccessLog> | AccessLog)}
   *
   * @memberOf AccessLogService
   */
  abstract getRecentLogs(
    queryParams?: RequestQueryParams
  ): Observable<AccessLog> | Promise<AccessLog> | AccessLog;
}

/**
 * Implement a default service for access log.
 *
 * @export
 * @class AccessLogDefaultService
 * @extends {AccessLogService}
 */
@Injectable()
export class AccessLogDefaultService extends AccessLogService {
  constructor(
    private http: Http,
    @Inject(SERVICE_CONFIG) private config: IServiceConfig
  ) {
    super();
  }

  public getAuditLogs(
    projectId: number | string,
    queryParams?: RequestQueryParams
  ): Observable<AccessLog> | Promise<AccessLog> | AccessLog {
    return Observable.of({} as AccessLog);
  }

  public getRecentLogs(
    queryParams?: RequestQueryParams
  ): Observable<AccessLog> | Promise<AccessLog> | AccessLog {
    let url: string = this.config.logBaseEndpoint
      ? this.config.logBaseEndpoint
      : "";
    if (url === "") {
      url = "/api/logs";
    }

    return this.http
      .get(
        url,
        queryParams ? buildHttpRequestOptions(queryParams) : HTTP_GET_OPTIONS
      )
      .toPromise()
      .then(response => {
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
            result.data = response.json() as AccessLogItem[];
          }
        }

        return result;
      })
      .catch(error => Promise.reject(error));
  }
}
