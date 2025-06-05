// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { CURRENT_BASE_HREF, HTTP_GET_OPTIONS_TEXT } from '../units/utils';
import { map, catchError } from 'rxjs/operators';
import { Observable, throwError as observableThrowError } from 'rxjs';
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

    constructor(private http: HttpClient) {
        super();
        this._replicationJobBaseUrl = CURRENT_BASE_HREF + '/replication';
        this._scanningJobBaseUrl = CURRENT_BASE_HREF + '/jobs/scan';
        this._supportedJobTypes = ['replication', 'scan'];
    }

    _getJobLog(logUrl: string): Observable<string> {
        return this.http.get(logUrl, HTTP_GET_OPTIONS_TEXT).pipe(
            map(response => response),
            catchError(error => observableThrowError(error))
        );
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
            return observableThrowError('Unsupport job type: ' + jobType);
        }
        if (!jobId || +jobId <= 0) {
            return observableThrowError('Bad argument');
        }

        let logUrl: string = `${this._replicationJobBaseUrl}/${jobId}/log`;
        if (jobType === 'scan') {
            logUrl = `${this._scanningJobBaseUrl}/${jobId}/log`;
        }

        return this._getJobLog(logUrl);
    }
}
