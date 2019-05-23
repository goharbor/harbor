import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { catchError, map } from 'rxjs/operators';
import { throwError as observableThrowError, Observable } from 'rxjs';




@Injectable()
export class AccountSettingsModalService {

  constructor(private http: HttpClient) { }
  generateCli(userId): Observable<any> {
    return this.http.post(`/api/users/${userId}/gen_cli_secret`, {}).pipe( map(response => response)
    , catchError(error => observableThrowError(error)));
  }
}
