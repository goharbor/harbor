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
import { HttpClient, HttpParams } from '@angular/common/http';
import { map, catchError } from 'rxjs/operators';
import { Observable, throwError as observableThrowError } from 'rxjs';
import { HTTP_FORM_OPTIONS } from '../../shared/units/utils';
import { SignInCredential } from './sign-in-credential';
const signInUrl = '/c/login';
/**
 *
 * Define a service to provide sign in methods
 *
 **
 * class SignInService
 */
@Injectable()
export class SignInService {
    constructor(private http: HttpClient) {}

    // Handle the related exceptions
    handleError(error: any): Observable<any> {
        return observableThrowError(error.error || error);
    }

    // Submit signin form to backend (NOT restful service)
    signIn(signInCredential: SignInCredential): Observable<any> {
        // Build the form package
        let body = new HttpParams();
        body = body.set('principal', signInCredential.principal);
        body = body.set('password', signInCredential.password);

        // Trigger HttpClient
        return this.http
            .post(signInUrl, body.toString(), HTTP_FORM_OPTIONS)
            .pipe(
                map(() => null),
                catchError(error => observableThrowError(error))
            );
    }
}

export const UN_LOGGED_PARAM: string = 'publicAndNotLogged';
export const YES: string = 'yes';
