
import {throwError as observableThrowError,  Observable } from "rxjs";

import {map, catchError} from 'rxjs/operators';
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { HttpClient, HttpParams, HttpResponse } from '@angular/common/http';

import { AuditLog } from './audit-log';
import {RequestQueryParams, buildHttpRequestOptions, buildHttpRequestOptionsWithObserveResponse} from '@harbor/ui';

export const logEndpoint = '/api/logs';

@Injectable()
export class AuditLogService {

  constructor(private http: HttpClient) {}

  listAuditLogs(queryParam: AuditLog): Observable<any> {
    let params: HttpParams = new HttpParams({fromString: queryParam.keywords});
    if (queryParam.begin_timestamp) {
      params = params.set('begin_timestamp', <string>queryParam.begin_timestamp);
    }
    if (queryParam.end_timestamp) {
      params = params.set('end_timestamp', <string>queryParam.end_timestamp);
    }
    if (queryParam.username) {
      params = params.set('username', queryParam.username);
    }
    if (queryParam.page) {
      params = params.set('page', <string>queryParam.page);
    }
    if (queryParam.page_size) {
      params = params.set('page_size', <string>queryParam.page_size);
    }
    return this.http
      .get<HttpResponse<AuditLog[]>>(`/api/projects/${queryParam.project_id}/logs`
      , buildHttpRequestOptionsWithObserveResponse(params)).pipe(
      catchError(error => observableThrowError(error)), );
  }

  getRecentLogs(lines: number): Observable<AuditLog[]> {
    let params: RequestQueryParams = new RequestQueryParams();
    params = params.set('page_size', '' + lines);
    return this.http.get(logEndpoint,  buildHttpRequestOptions(params)).pipe(
      map(response => response as AuditLog[]),
      catchError(error => observableThrowError(error)), );
  }
}
