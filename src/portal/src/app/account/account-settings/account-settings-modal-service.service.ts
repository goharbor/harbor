import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { catchError, map } from 'rxjs/operators';
import { throwError as observableThrowError, Observable } from 'rxjs';
import { CURRENT_BASE_HREF } from "../../../lib/utils/utils";




@Injectable()
export class AccountSettingsModalService {

  constructor(private http: HttpClient) { }
  saveNewCli(userId, secretObj): Observable<any> {
    return this.http.put(`${ CURRENT_BASE_HREF }/users/${userId}/cli_secret`, secretObj).pipe( map(response => response)
    , catchError(error => observableThrowError(error)));
  }
}
