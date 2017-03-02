import { Injectable } from '@angular/core';
import { Http } from '@angular/http';

import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';

import { BaseService } from '../../service/base.service';
import { Member } from './member';

@Injectable()
export class MemberService extends BaseService {
  
  constructor(private http: Http) {
    super();
  }

  listMembers(projectId: number, username: string): Observable<Member[]> {
    console.log('Get member from project_id:' + projectId + ', username:' + username);
    return this.http
               .get(`/api/projects/${projectId}/members?username=${username}`)
               .map(response=>response.json())
               .catch(error=>this.handleError(error));            
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