
import {throwError as observableThrowError,  Observable} from "rxjs";

import {catchError, map} from 'rxjs/operators';
import { Injectable } from "@angular/core";
import { Http, Response } from "@angular/http";
import { UserGroup } from "./group";
import { HTTP_JSON_OPTIONS, HTTP_GET_OPTIONS } from "../shared/shared.utils";

const userGroupEndpoint = "/api/usergroups";
const ldapGroupSearchEndpoint = "/api/ldap/groups/search?groupname=";

@Injectable()
export class GroupService {
  constructor(private http: Http) {}

  private extractData(res: Response) {
    if (res.text() === '') {return []; }
    return res.json() || [];
  }
  private handleErrorObservable(error: Response | any) {
    console.error(error.message || error);
    return observableThrowError(error.message || error);
  }

  getUserGroups(): Observable<UserGroup[]> {
    return this.http.get(userGroupEndpoint, HTTP_GET_OPTIONS).pipe(
    map(response => {
      return this.extractData(response);
    }),
    catchError(error => {
      return this.handleErrorObservable(error);
    }), );
  }

  createGroup(group: UserGroup): Observable<any> {
    return this.http
      .post(userGroupEndpoint, group, HTTP_JSON_OPTIONS).pipe(
      map(response => {
        return this.extractData(response);
      }),
      catchError(this.handleErrorObservable), );
  }

  getGroup(group_id: number): Observable<UserGroup> {
    return this.http
      .get(`${userGroupEndpoint}/${group_id}`, HTTP_JSON_OPTIONS).pipe(
      map(response => {
        return this.extractData(response);
      }),
      catchError(this.handleErrorObservable), );
  }

  editGroup(group: UserGroup): Observable<any> {
    return this.http
    .put(`${userGroupEndpoint}/${group.id}`, group, HTTP_JSON_OPTIONS).pipe(
    map(response => {
      return this.extractData(response);
    }),
    catchError(this.handleErrorObservable), );
  }

  deleteGroup(group_id: number): Observable<any> {
    return this.http
    .delete(`${userGroupEndpoint}/${group_id}`).pipe(
    map(response => {
      return this.extractData(response);
    }),
    catchError(this.handleErrorObservable), );
  }

  searchGroup(group_name: string): Observable<UserGroup[]> {
    return this.http
    .get(`${ldapGroupSearchEndpoint}${group_name}`, HTTP_GET_OPTIONS).pipe(
    map(response => {
      return this.extractData(response);
    }),
    catchError(this.handleErrorObservable), );
  }
}
