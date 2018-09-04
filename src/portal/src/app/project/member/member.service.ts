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

import {HTTP_JSON_OPTIONS, HTTP_GET_OPTIONS} from "../../shared/shared.utils";
import { User } from '../../user/user';
import { Member } from './member';


@Injectable()
export class MemberService {

  constructor(private http: Http) {}

  listMembers(projectId: number, entity_name: string): Observable<Member[]> {
    return this.http
               .get(`/api/projects/${projectId}/members?entityname=${entity_name}`, HTTP_GET_OPTIONS)
               .map(response => response.json() as Member[])
               .catch(error => Observable.throw(error));
  }

  addUserMember(projectId: number, user: User, roleId: number): Observable<any> {
    let member_user = {};
    if (user.user_id) {
      member_user = {user_id: user.user_id};
    } else if (user.username) {
      member_user = {username: user.username};
    } else {
      return;
    }
    return this.http.post(
      `/api/projects/${projectId}/members`,
      {
        role_id: roleId,
        member_user: member_user
      },
      HTTP_JSON_OPTIONS)
      .map(response => response.status)
      .catch(error => Observable.throw(error));
  }

  addGroupMember(projectId: number, group: any, roleId: number): Observable<any> {
    return this.http
               .post(`/api/projects/${projectId}/members`,
               { role_id: roleId, member_group: group},
               HTTP_JSON_OPTIONS)
               .map(response => response.status)
               .catch(error => Observable.throw(error));
  }

  changeMemberRole(projectId: number, userId: number, roleId: number): Promise<any> {
    return this.http
               .put(`/api/projects/${projectId}/members/${userId}`, { role_id: roleId }, HTTP_JSON_OPTIONS).toPromise()
               .then(response => response.status)
               .catch(error => Promise.reject(error));
  }

  deleteMember(projectId: number, memberId: number): Promise<any> {
    return this.http
               .delete(`/api/projects/${projectId}/members/${memberId}`).toPromise()
               .then(response => response.status)
               .catch(error => Promise.reject(error));
  }
}
