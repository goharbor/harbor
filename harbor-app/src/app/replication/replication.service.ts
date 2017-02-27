import { Injectable } from '@angular/core';
import { Http, URLSearchParams } from '@angular/http';

import { BaseService } from '../service/base.service';

import { Policy } from './policy';
import { Job } from './job';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';

export const urlPrefix = '';

@Injectable()
export class ReplicationService extends BaseService {
  constructor(private http: Http) {
    super();
  }

  listPolicies(projectId: number, policyName: string): Observable<Policy[]> {
    console.log('Get policies with project ID:' + projectId + ', policy name:' + policyName);
    return this.http
               .get(urlPrefix + `/api/policies/replication?project_id=${projectId}`)
               .map(response=>response.json() as Policy[])
               .catch(error=>Observable.throw(error));
  }

  // /api/jobs/replication/?page=1&page_size=20&end_time=&policy_id=1&start_time=&status=
  listJobs(policyId: number, status: string = ''): Observable<Job[]> {
    console.log('Get jobs under policy ID:' + policyId);
    return this.http
               .get(urlPrefix + `/api/jobs/replication?policy_id=${policyId}&status=${status}`)
               .map(response=>response.json() as Job[])
               .catch(error=>Observable.throw(error));
  }
}