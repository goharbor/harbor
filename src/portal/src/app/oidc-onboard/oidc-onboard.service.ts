import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { catchError } from 'rxjs/operators';
import { throwError as observableThrowError, Observable } from 'rxjs';

export const logEndpoint = '/c/oidc/onboard';

@Injectable()
export class OidcOnboardService {
    constructor(private http: HttpClient) {}
    oidcSave(param): Observable<any> {
        return this.http
            .post(logEndpoint, param)
            .pipe(catchError(error => observableThrowError(error)));
    }
}
