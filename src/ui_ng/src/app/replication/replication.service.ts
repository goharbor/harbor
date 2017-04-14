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
import { Http, URLSearchParams, Response } from '@angular/http';

import { Policy } from './policy';
import { Job } from './job';
import { Target } from './target';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';
import 'rxjs/add/operator/mergeMap';

@Injectable()
export class ReplicationService {
  constructor(private http: Http) {}

  listPolicies(policyName: string, projectId?: any): Observable<Policy[]> {
    if(!projectId) {
      projectId = '';
    }
    return this.http
               .get(`/api/policies/replication?project_id=${projectId}&name=${policyName}`)
               .map(response=>response.json() as Policy[])
               .catch(error=>Observable.throw(error));
  }

  getPolicy(policyId: number): Observable<Policy> {
    return this.http
               .get(`/api/policies/replication/${policyId}`)
               .map(response=>response.json() as Policy)
               .catch(error=>Observable.throw(error));
  }

  createPolicy(policy: Policy): Observable<any> {
    return this.http
               .post(`/api/policies/replication`, JSON.stringify(policy))
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  updatePolicy(policy: Policy): Observable<any> {
    if (policy && policy.id) {
      return this.http
                 .put(`/api/policies/replication/${policy.id}`, JSON.stringify(policy))
                 .map(response=>response.status)
                 .catch(error=>Observable.throw(error));
    } 
    return Observable.throw(new Error("Policy is nil or has no ID set."));
  }

  createOrUpdatePolicyWithNewTarget(policy: Policy, target: Target): Observable<any> {
    return this.http
               .post(`/api/targets`, JSON.stringify(target))
               .map(response=>{
                 return response.status;
               })
               .catch(error=>Observable.throw(error))
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
                     if(policy.id) {
                       return this.http
                                  .put(`/api/policies/replication/${policy.id}`, JSON.stringify(policy))
                                  .map(response=>response.status)
                                  .catch(error=>Observable.throw(error));
                     } else {
                       return this.http
                                  .post(`/api/policies/replication`, JSON.stringify(policy))
                                  .map(response=>response.status)
                                  .catch(error=>Observable.throw(error));
                     }
                   } 
                 }
               })
               .catch(error=>Observable.throw(error));
  }

  enablePolicy(policyId: number, enabled: number): Observable<any> {
    console.log('Enable or disable policy ID:' + policyId + ' with activation status:' + enabled);
    return this.http
               .put(`/api/policies/replication/${policyId}/enablement`, {enabled: enabled})
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  deletePolicy(policyId: number): Observable<any> {
    console.log('Delete policy ID:' + policyId);
    return this.http
               .delete(`/api/policies/replication/${policyId}`)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  // /api/jobs/replication/?page=1&page_size=20&end_time=&policy_id=1&start_time=&status=&repository=
  listJobs(policyId: number, status: string = '', repoName: string = '', startTime: string = '', endTime: string = '', page: number, pageSize: number): Observable<any> {
    return this.http
               .get(`/api/jobs/replication?policy_id=${policyId}&status=${status}&repository=${repoName}&start_time=${startTime}&end_time=${endTime}&page=${page}&page_size=${pageSize}`)
               .map(response=>response)
               .catch(error=>Observable.throw(error));
  }

  listTargets(targetName: string): Observable<Target[]> {
    return this.http
               .get(`/api/targets?name=${targetName}`)
               .map(response=>response.json() as Target[])
               .catch(error=>Observable.throw(error));
  }

  listTargetPolicies(targetId: number): Observable<Policy[]> {
    return this.http
               .get(`/api/targets/${targetId}/policies`)
               .map(response=>response.json() as Policy[])
               .catch(error=>Observable.throw(error));
  }

  getTarget(targetId: number): Observable<Target> {
    return this.http
               .get(`/api/targets/${targetId}`)
               .map(response=>response.json() as Target)
               .catch(error=>Observable.throw(error));
  }

  createTarget(target: Target): Observable<any> {
    return this.http
               .post(`/api/targets`, JSON.stringify(target))
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  pingTarget(target: Target): Observable<any> {
    if(target.id) {
      return this.http
               .post(`/api/targets/${target.id}/ping`, {})
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
    }
    return this.http
               .post(`/api/targets/ping`, target)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  updateTarget(target: Target, targetId: number): Observable<any> {
    return this.http
               .put(`/api/targets/${targetId}`, JSON.stringify(target))
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  deleteTarget(targetId: number): Observable<any> {
    return this.http
               .delete(`/api/targets/${targetId}`)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

}