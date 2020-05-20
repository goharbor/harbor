import { Injectable, Inject } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { SERVICE_CONFIG, IServiceConfig } from "../entities/service.config";
import { CURRENT_BASE_HREF, HTTP_GET_OPTIONS, HTTP_GET_OPTIONS_TEXT } from "../utils/utils";
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
/**
 * Define the service methods to handle the job log related things.
 *
 **
 * @abstract
 * class JobLogService
 */
export abstract class JobLogService {
  /**
   * Get the log of the specified job
   *
   * @abstract
   *  ** deprecated param {string} jobType
   *  ** deprecated param {(number | string)} jobId
   * returns {(Observable<string>)}
   * @memberof JobLogService
   */

  abstract getScanJobBaseUrl(): string;
  abstract getJobLog(
    jobType: string,
    jobId: number | string
  ): Observable<string>;
}

/**
 * Implement default service for job log service.
 *
 **
 * class JobLogDefaultService
 * extends {ReplicationService}
 */
@Injectable()
export class JobLogDefaultService extends JobLogService {
  _replicationJobBaseUrl: string;
  _scanningJobBaseUrl: string;
  _supportedJobTypes: string[];

  constructor(
    private http: HttpClient,
    @Inject(SERVICE_CONFIG) config: IServiceConfig
  ) {
    super();
    this._replicationJobBaseUrl = config.replicationBaseEndpoint
      ? config.replicationBaseEndpoint
      : CURRENT_BASE_HREF + "/replication";
    this._scanningJobBaseUrl = config.scanJobEndpoint
      ? config.scanJobEndpoint
      : CURRENT_BASE_HREF + "/jobs/scan";
    this._supportedJobTypes = ["replication", "scan"];
  }

  _getJobLog(logUrl: string): Observable<string> {
    return this.http
      .get(logUrl, HTTP_GET_OPTIONS_TEXT)
      .pipe(map(response => response)
      , catchError(error => observableThrowError(error)));
  }

  _isSupportedJobType(jobType: string): boolean {
    if (this._supportedJobTypes.find((t: string) => t === jobType)) {
      return true;
    }

    return false;
  }

  public getScanJobBaseUrl() {
    return this._scanningJobBaseUrl;
  }

  public getJobLog(
    jobType: string,
    jobId: number | string
  ): Observable<string> {
    if (!this._isSupportedJobType(jobType)) {
      return observableThrowError("Unsupport job type: " + jobType);
    }
    if (!jobId || jobId <= 0) {
      return observableThrowError("Bad argument");
    }

    let logUrl: string = `${this._replicationJobBaseUrl}/${jobId}/log`;
    if (jobType === "scan") {
      logUrl = `${this._scanningJobBaseUrl}/${jobId}/log`;
    }

    return this._getJobLog(logUrl);
  }
}
