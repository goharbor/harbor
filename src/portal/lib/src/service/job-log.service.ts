import { Observable } from "rxjs";
import { Injectable, Inject } from "@angular/core";
import { Http } from "@angular/http";
import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import { HTTP_GET_OPTIONS } from "../utils";

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
   * returns {(Observable<string> | Promise<string> | string)}
   * @memberof JobLogService
   */

  abstract getScanJobBaseUrl(): string;
  abstract getJobLog(
    jobType: string,
    jobId: number | string
  ): Observable<string> | Promise<string> | string;
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
    private http: Http,
    @Inject(SERVICE_CONFIG) config: IServiceConfig
  ) {
    super();
    this._replicationJobBaseUrl = config.replicationJobEndpoint
      ? config.replicationJobEndpoint
      : "/api/jobs/replication";
    this._scanningJobBaseUrl = config.scanJobEndpoint
      ? config.scanJobEndpoint
      : "/api/jobs/scan";
    this._supportedJobTypes = ["replication", "scan"];
  }

  _getJobLog(logUrl: string): Observable<string> | Promise<string> | string {
    return this.http
      .get(logUrl, HTTP_GET_OPTIONS)
      .toPromise()
      .then(response => response.text())
      .catch(error => Promise.reject(error));
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
  ): Observable<string> | Promise<string> | string {
    if (!this._isSupportedJobType(jobType)) {
      return Promise.reject("Unsupport job type: " + jobType);
    }
    if (!jobId || jobId <= 0) {
      return Promise.reject("Bad argument");
    }

    let logUrl: string = `${this._replicationJobBaseUrl}/${jobId}/log`;
    if (jobType === "scan") {
      logUrl = `${this._scanningJobBaseUrl}/${jobId}/log`;
    }

    return this._getJobLog(logUrl);
  }
}
