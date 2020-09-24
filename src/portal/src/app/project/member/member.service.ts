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
import { HttpClient } from '@angular/common/http';
import { User } from '../../user/user';
import { Member } from './member';
import {throwError as observableThrowError,  Observable } from "rxjs";
import {map, catchError} from 'rxjs/operators';
import { CURRENT_BASE_HREF, HTTP_GET_OPTIONS, HTTP_JSON_OPTIONS } from "../../../lib/utils/utils";

@Injectable()
export class MemberService {

  constructor(private http: HttpClient) {}

  listMembers(projectId: number, entity_name: string): Observable<Member[]> {
    return this.http
               .get(`${ CURRENT_BASE_HREF }/projects/${projectId}/members?entityname=${entity_name}`, HTTP_GET_OPTIONS).pipe(
               map(response => response as Member[]),
               catchError(error => observableThrowError(error)), );
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
      `${ CURRENT_BASE_HREF }/projects/${projectId}/members`,
      {
        role_id: roleId,
        member_user: member_user
      },
      HTTP_JSON_OPTIONS).pipe(
      catchError(error => observableThrowError(error)), );
  }

  addGroupMember(projectId: number, group: any, roleId: number): Observable<any> {
    return this.http
               .post(`${ CURRENT_BASE_HREF }/projects/${projectId}/members`,
               { role_id: roleId, member_group: group},
               HTTP_JSON_OPTIONS).pipe(
               catchError(error => observableThrowError(error)), );
  }

  changeMemberRole(projectId: number, userId: number, roleId: number): Observable<any> {
    return this.http
               .put(`${ CURRENT_BASE_HREF }/projects/${projectId}/members/${userId}`, { role_id: roleId }, HTTP_JSON_OPTIONS)
               .pipe(catchError(error => observableThrowError(error)));
  }

  deleteMember(projectId: number, memberId: number): Observable<any> {
    return this.http
               .delete(`${ CURRENT_BASE_HREF }/projects/${projectId}/members/${memberId}`)
               .pipe(catchError(error => observableThrowError(error)));
  }
}
