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
import { Http, Headers, RequestOptions } from '@angular/http';

import { AuditLog } from './audit-log';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';

export const logEndpoint = "/api/logs";

@Injectable()
export class AuditLogService {
  private httpOptions = new RequestOptions({
    headers: new Headers({
      "Content-Type": 'application/json',
      "Accept": 'application/json'
    })
  });

  constructor(private http: Http) {}

  listAuditLogs(queryParam: AuditLog): Observable<any> {
    return this.http
      .post(`/api/projects/${queryParam.project_id}/logs/filter?page=${queryParam.page}&page_size=${queryParam.page_size}`, {
        begin_timestamp: queryParam.begin_timestamp,
        end_timestamp: queryParam.end_timestamp,
        keywords: queryParam.keywords,
        operation: queryParam.operation,
        project_id: queryParam.project_id,
        username: queryParam.username
      })
      .map(response => response)
      .catch(error => Observable.throw(error));
  }

  getRecentLogs(lines: number): Observable<AuditLog[]> {
    return this.http.get(logEndpoint + "?lines=" + lines, this.httpOptions)
      .map(response => response.json() as AuditLog[])
      .catch(error => Observable.throw(error));
  }
}