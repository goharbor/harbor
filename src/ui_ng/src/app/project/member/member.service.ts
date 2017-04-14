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
import { Http } from '@angular/http';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';

import { Member } from './member';

@Injectable()
export class MemberService {
  
  constructor(private http: Http) {}

  listMembers(projectId: number, username: string): Observable<Member[]> {
    console.log('Get member from project_id:' + projectId + ', username:' + username);
    return this.http
               .get(`/api/projects/${projectId}/members?username=${username}`)
               .map(response=>response.json() as Member[])
               .catch(error=>Observable.throw(error));            
  }

  addMember(projectId: number, username: string, roleId: number): Observable<any> {
    console.log('Adding member with username:' + username + ', roleId:' + roleId + ' under projectId:' + projectId);
    return this.http
               .post(`/api/projects/${projectId}/members`, { username: username, roles: [ roleId ] })
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  changeMemberRole(projectId: number, userId: number, roleId: number): Observable<any> {
    console.log('Changing member role with userId:' + ' to roleId:' + roleId + ' under projectId:' + projectId);
    return this.http
               .put(`/api/projects/${projectId}/members/${userId}`, { roles: [ roleId ]})
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }

  deleteMember(projectId: number, userId: number): Observable<any> {
    console.log('Deleting member role with userId:' + userId + ' under projectId:' + projectId);
    return this.http
               .delete(`/api/projects/${projectId}/members/${userId}`)
               .map(response=>response.status)
               .catch(error=>Observable.throw(error));
  }
}