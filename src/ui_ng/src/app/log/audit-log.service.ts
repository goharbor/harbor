import { Injectable } from '@angular/core';
import { Http, Headers, RequestOptions } from '@angular/http';

import { BaseService } from '../service/base.service';

import { AuditLog } from './audit-log';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';

export const urlPrefix = '';

@Injectable()
export class AuditLogService extends BaseService {
  
  constructor(private http: Http) {
    super();
  }

  listAuditLogs(queryParam: AuditLog): Observable<AuditLog[]> {
    return this.http
               .post(urlPrefix + `/api/projects/${queryParam.project_id}/logs/filter`, {
                  begin_timestamp: queryParam.begin_timestamp,
                  end_timestamp: queryParam.end_timestamp,
                  keywords: queryParam.keywords,
                  operation: queryParam.operation,
                  project_id: queryParam.project_id,
                  username: queryParam.username })
                .map(response=>response.json() as AuditLog[])
                .catch(error=>this.handleError(error));
  }

}