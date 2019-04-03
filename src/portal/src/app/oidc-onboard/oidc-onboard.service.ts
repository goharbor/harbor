import { Injectable } from '@angular/core';
import { Http, URLSearchParams } from '@angular/http';
import { catchError } from 'rxjs/operators';
import { throwError as observableThrowError, Observable } from 'rxjs';


export const logEndpoint = "/c/oidc/onboard";

@Injectable()
export class OidcOnboardService {

  constructor(private http: Http) { }
  oidcSave(param): Observable<any> {
    return this.http.post(logEndpoint, param).pipe(catchError(error => observableThrowError(error)));
  }
}
