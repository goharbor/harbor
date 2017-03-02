import { Injectable } from '@angular/core';
import { Http, URLSearchParams, Response } from '@angular/http';

import { BaseService } from '../service/base.service';

import { Policy } from './policy';
import { Job } from './job';
import { Target } from './target';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';
import 'rxjs/add/operator/mergeMap';

@Injectable()
export class ReplicationService extends BaseService {
  constructor(private http: Http) {
    super();
  }

  listPolicies(projectId: number, policyName: string): Observable<Policy[]> {
    console.log('Get policies with project ID:' + projectId + ', policy name:' + policyName);
    return this.http
               .get(`/api/policies/replication?project_id=${projectId}&name=${policyName}`)
               .map(response=>response.json() as Policy[])
               .catch(error=>Observable.throw(error));
  }

  getPolicy(policyId: number): Observable<Policy> {
    console.log('Get policy with ID:' + policyId);
    return this.http
               .get(`/api/policies/replication/${policyId}`)
               .map(response=>response.json() as Policy)
               .catch(error=>Observable.throw(error));
  }

  createPolicy(policy: Policy): Observable<any> {
    console.log('Create policy with project ID:' + policy.project_id + ', policy:' + JSON.stringify(policy));
    return this.http
               .post(`/api/policies/replication`, JSON.stringify(policy))
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  createTarget(target: Target): Observable<any> {
    console.log('Create target:' + JSON.stringify(target));
    return this.http
               .post(`/api/targets`, JSON.stringify(target))
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  createPolcyWithTarget(policy: Policy, target: Target): Observable<any> {
    return this.http
               .post(`/api/targets`, JSON.stringify(target))
               .map(response=>{
                 return response.status;
               })
               .flatMap((status)=>{
                 if(status === 201) {
                   return this.http
                              .get(`/api/targets?name=${target.name}`)
                              .map(res=>res)
                              .catch(error=>Observable.throw(error));
                 }
               })
               .flatMap((res: Response) => { 
                 if(res.status === 200) {
                   let lastAddedTarget= <Target>res.json()[0];
                   if(lastAddedTarget && lastAddedTarget.id) {
                     policy.target_id = lastAddedTarget.id;
                     return this.http
                                .post(`/api/policies/replication`, JSON.stringify(policy))
                                .map(response=>response.status)
                                .catch(error=>Observable.throw(error));
                   }
                 }
               })
               .catch(error=>Observable.throw(error));
  }

  // /api/jobs/replication/?page=1&page_size=20&end_time=&policy_id=1&start_time=&status=&repository=
  listJobs(policyId: number, status: string = '', repoName: string = '', startTime: string = '', endTime: string = ''): Observable<Job[]> {
    console.log('Get jobs under policy ID:' + policyId);
    return this.http
               .get(`/api/jobs/replication?policy_id=${policyId}&status=${status}&repository=${repoName}&start_time=${startTime}&end_time=${endTime}`)
               .map(response=>response.json() as Job[])
               .catch(error=>Observable.throw(error));
  }

  listTargets(): Observable<Target[]> {
    console.log('Get targets.');
    return this.http
               .get(`/api/targets`)
               .map(response=>response.json() as Target[])
               .catch(error=>Observable.throw(error));
  }
}