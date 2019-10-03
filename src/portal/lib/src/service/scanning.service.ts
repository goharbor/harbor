import { HttpClient, HttpHeaders } from "@angular/common/http";
import { Injectable, Inject } from "@angular/core";

import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import { buildHttpRequestOptions, HTTP_JSON_OPTIONS } from "../utils";
import { RequestQueryParams } from "./RequestQueryParams";
import { VulnerabilityItem, VulnerabilitySummary } from "./interface";
import { map, catchError } from "rxjs/operators";
import { Observable, of, throwError as observableThrowError } from "rxjs";

// The default report mime type
const DEFAULT_REPORT_MIME_TYPE = "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"

/**
 * Get the vulnerabilities scanning results for the specified tag.
 *
 **
 * @abstract
 * class ScanningResultService
 */
export abstract class ScanningResultService {
  /**
   * Get the summary of vulnerability scanning result.
   *
   * @abstract
   *  ** deprecated param {string} tagId
   * returns {(Observable<VulnerabilitySummary>)}
   *
   * @memberOf ScanningResultService
   */
  abstract getVulnerabilityScanningSummary(
    repoName: string,
    tagId: string,
    queryParams?: RequestQueryParams
  ):
    | Observable<VulnerabilitySummary>;

  /**
   * Get the detailed vulnerabilities scanning results.
   *
   * @abstract
   *  ** deprecated param {string} tagId
   * returns {(Observable<VulnerabilityItem[]>)}
   *
   * @memberOf ScanningResultService
   */
  abstract getVulnerabilityScanningResults(
    repoName: string,
    tagId: string,
    queryParams?: RequestQueryParams
  ):
    | Observable<any>;

  /**
   * Start a new vulnerability scanning
   *
   * @abstract
   *  ** deprecated param {string} repoName
   *  ** deprecated param {string} tagId
   * returns {(Observable<any>)}
   *
   * @memberOf ScanningResultService
   */
  abstract startVulnerabilityScanning(
    repoName: string,
    tagId: string
  ): Observable<any>;

  /**
   * Trigger the scanning all action.
   *
   * @abstract
   * returns {(Observable<any>)}
   *
   * @memberOf ScanningResultService
   */
  abstract startScanningAll(): Observable<any>;
}

@Injectable()
export class ScanningResultDefaultService extends ScanningResultService {
  _baseUrl: string = "/api/repositories";

  constructor(
    private http: HttpClient,
    @Inject(SERVICE_CONFIG) private config: IServiceConfig
  ) {
    super();
    if (this.config && this.config.vulnerabilityScanningBaseEndpoint) {
      this._baseUrl = this.config.vulnerabilityScanningBaseEndpoint;
    }
  }

  getVulnerabilityScanningSummary(
    repoName: string,
    tagId: string,
    queryParams?: RequestQueryParams
  ):
    | Observable<VulnerabilitySummary>  {
    if (!repoName || repoName.trim() === "" || !tagId || tagId.trim() === "") {
      return observableThrowError("Bad argument");
    }

    return of({} as VulnerabilitySummary);
  }

  getVulnerabilityScanningResults(
    repoName: string,
    tagId: string,
    queryParams?: RequestQueryParams
  ):
    | Observable<any> {
    if (!repoName || repoName.trim() === "" || !tagId || tagId.trim() === "") {
      return observableThrowError("Bad argument");
    }

    let httpOptions = buildHttpRequestOptions(queryParams)
    let requestHeaders = httpOptions.headers as HttpHeaders
    // Change the accept header to the supported report mime types
    httpOptions.headers = requestHeaders.set("Accept", DEFAULT_REPORT_MIME_TYPE)

    return this.http
      .get(
        `${this._baseUrl}/${repoName}/tags/${tagId}/scan`,
        httpOptions
      )
      .pipe(map(response => response as any)
      , catchError(error => observableThrowError(error)));
  }

  startVulnerabilityScanning(
    repoName: string,
    tagId: string
  ): Observable<any> {
    if (!repoName || repoName.trim() === "" || !tagId || tagId.trim() === "") {
      return observableThrowError("Bad argument");
    }

    return this.http
      .post(
        `${this._baseUrl}/${repoName}/tags/${tagId}/scan`,
        HTTP_JSON_OPTIONS
      )
      .pipe(map(() => {
        return true;
      })
      , catchError(error => observableThrowError(error)));
  }

  startScanningAll(): Observable<any> {
    return this.http
      .post(`${this._baseUrl}/scanAll`, HTTP_JSON_OPTIONS)
      .pipe(map(() => {
        return true;
      })
      , catchError(error => observableThrowError(error)));
  }
}
