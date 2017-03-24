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