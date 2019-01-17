import { Injectable } from "@angular/core";
import { Http } from "@angular/http";
import { throwError as observableThrowError, Observable, pipe } from "rxjs";
import { catchError, map } from "rxjs/operators";
import { Robot } from './robot';
import { HTTP_JSON_OPTIONS } from "../../shared/shared.utils";

@Injectable()
export class RobotApiRepository {
  constructor(private http: Http) {}

  public postRobot(projectId, param): Observable<any> {
    return this.http
      .post(`/api/projects/${projectId}/robots`, param)
      .pipe(map(response => response.json()))
      .pipe(catchError(error => observableThrowError(error)));
  }

  public deleteRobot(projectId, id): Observable<any> {
    return this.http
      .delete(`/api/projects/${projectId}/robots/${id}`)
      .pipe(catchError(error => observableThrowError(error)));
  }

  public listRobot(projectId): Observable<Robot[]> {
    return this.http
      .get(`/api/projects/${projectId}/robots`)
      .pipe(map(response => response.json() as Robot[]))
      .pipe(catchError(error => observableThrowError(error)));
  }

  public getRobot(projectId, id): Observable<Robot[]> {
    return this.http
      .get(`/api/projects/${projectId}/robots/${id}`)
      .pipe(map(response => response.json() as Robot[]))
      .pipe(catchError(error => observableThrowError(error)));
  }

  public toggleDisabledAccount(projectId, id, data): Observable<any> {
    return this.http
      .put(`/api/projects/${projectId}/robots/${id}`, data)
      .pipe(catchError(error => observableThrowError(error)));
  }
}
