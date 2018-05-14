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
import { Http, URLSearchParams } from '@angular/http';

import { AuditLog } from './audit-log';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';
import {buildHttpRequestOptions} from '../shared/shared.utils';
import {RequestQueryParams} from 'harbor-ui';

export const logEndpoint = '/api/logs';

@Injectable()
export class AuditLogService {

  constructor(private http: Http) {}

  listAuditLogs(queryParam: AuditLog): Observable<any> {
    let params: URLSearchParams = new URLSearchParams(queryParam.keywords);
    if (queryParam.begin_timestamp) {
      params.set('begin_timestamp', <string>queryParam.begin_timestamp);
    }
    if (queryParam.end_timestamp) {
      params.set('end_timestamp', <string>queryParam.end_timestamp);
    }
    if (queryParam.username) {
      params.set('username', queryParam.username);
    }
    if (queryParam.page) {
      params.set('page', <string>queryParam.page);
    }
    if (queryParam.page_size) {
      params.set('page_size', <string>queryParam.page_size);
    }
    return this.http
      .get(`/api/projects/${queryParam.project_id}/logs`, buildHttpRequestOptions(params))
      .map(response => response)
      .catch(error => Observable.throw(error));
  }

  getRecentLogs(lines: number): Observable<AuditLog[]> {
    let params: RequestQueryParams = new RequestQueryParams();
    params.set('page_size', '' + lines);
    return this.http.get(logEndpoint,  buildHttpRequestOptions(params))
      .map(response => response.json() as AuditLog[])
      .catch(error => Observable.throw(error));
  }
}
