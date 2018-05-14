import { Http } from "@angular/http";
import { Injectable, Inject } from "@angular/core";
import { Observable } from "rxjs/Observable";
import "rxjs/add/observable/of";

import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import { buildHttpRequestOptions, HTTP_JSON_OPTIONS } from "../utils";
import { RequestQueryParams } from "./RequestQueryParams";
import { VulnerabilityItem, VulnerabilitySummary } from "./interface";

/**
 * Get the vulnerabilities scanning results for the specified tag.
 *
 * @export
 * @abstract
 * @class ScanningResultService
 */
export abstract class ScanningResultService {
  /**
   * Get the summary of vulnerability scanning result.
   *
   * @abstract
   * @param {string} tagId
   * @returns {(Observable<VulnerabilitySummary> | Promise<VulnerabilitySummary> | VulnerabilitySummary)}
   *
   * @memberOf ScanningResultService
   */
  abstract getVulnerabilityScanningSummary(
    repoName: string,
    tagId: string,
    queryParams?: RequestQueryParams
  ):
    | Observable<VulnerabilitySummary>
    | Promise<VulnerabilitySummary>
    | VulnerabilitySummary;

  /**
   * Get the detailed vulnerabilities scanning results.
   *
   * @abstract
   * @param {string} tagId
   * @returns {(Observable<VulnerabilityItem[]> | Promise<VulnerabilityItem[]> | VulnerabilityItem[])}
   *
   * @memberOf ScanningResultService
   */
  abstract getVulnerabilityScanningResults(
    repoName: string,
    tagId: string,
    queryParams?: RequestQueryParams
  ):
    | Observable<VulnerabilityItem[]>
    | Promise<VulnerabilityItem[]>
    | VulnerabilityItem[];

  /**
   * Start a new vulnerability scanning
   *
   * @abstract
   * @param {string} repoName
   * @param {string} tagId
   * @returns {(Observable<any> | Promise<any> | any)}
   *
   * @memberOf ScanningResultService
   */
  abstract startVulnerabilityScanning(
    repoName: string,
    tagId: string
  ): Observable<any> | Promise<any> | any;

  /**
   * Trigger the scanning all action.
   *
   * @abstract
   * @returns {(Observable<any> | Promise<any> | any)}
   *
   * @memberOf ScanningResultService
   */
  abstract startScanningAll(): Observable<any> | Promise<any> | any;
}

@Injectable()
export class ScanningResultDefaultService extends ScanningResultService {
  _baseUrl: string = "/api/repositories";

  constructor(
    private http: Http,
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
    | Observable<VulnerabilitySummary>
    | Promise<VulnerabilitySummary>
    | VulnerabilitySummary {
    if (!repoName || repoName.trim() === "" || !tagId || tagId.trim() === "") {
      return Promise.reject("Bad argument");
    }

    return Observable.of({} as VulnerabilitySummary);
  }

  getVulnerabilityScanningResults(
    repoName: string,
    tagId: string,
    queryParams?: RequestQueryParams
  ):
    | Observable<VulnerabilityItem[]>
    | Promise<VulnerabilityItem[]>
    | VulnerabilityItem[] {
    if (!repoName || repoName.trim() === "" || !tagId || tagId.trim() === "") {
      return Promise.reject("Bad argument");
    }

    return this.http
      .get(
        `${this._baseUrl}/${repoName}/tags/${tagId}/vulnerability/details`,
        buildHttpRequestOptions(queryParams)
      )
      .toPromise()
      .then(response => response.json() as VulnerabilityItem[])
      .catch(error => Promise.reject(error));
  }

  startVulnerabilityScanning(
    repoName: string,
    tagId: string
  ): Observable<any> | Promise<any> | any {
    if (!repoName || repoName.trim() === "" || !tagId || tagId.trim() === "") {
      return Promise.reject("Bad argument");
    }

    return this.http
      .post(
        `${this._baseUrl}/${repoName}/tags/${tagId}/scan`,
        HTTP_JSON_OPTIONS
      )
      .toPromise()
      .then(() => {
        return true;
      })
      .catch(error => Promise.reject(error));
  }

  startScanningAll(): Observable<any> | Promise<any> | any {
    return this.http
      .post(`${this._baseUrl}/scanAll`, HTTP_JSON_OPTIONS)
      .toPromise()
      .then(() => {
        return true;
      })
      .catch(error => Promise.reject(error));
  }
}
