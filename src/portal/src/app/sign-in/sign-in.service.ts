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
import { Http, URLSearchParams } from '@angular/http';
// import 'rxjs/add/operator/toPromise';

import { SignInCredential } from '../shared/sign-in-credential';
import {HTTP_FORM_OPTIONS} from "../shared/shared.utils";
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
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

    constructor(private http: Http) {}

    // Handle the related exceptions
    handleError(error: any): Observable<any> {
        return observableThrowError(error.message || error);
    }

    // Submit signin form to backend (NOT restful service)
    signIn(signInCredential: SignInCredential): Observable<any> {
        // Build the form package
        const body = new URLSearchParams();
        body.set('principal', signInCredential.principal);
        body.set('password', signInCredential.password);

        // Trigger Http
        return this.http.post(signInUrl, body.toString(), HTTP_FORM_OPTIONS)
        .pipe(map(() => null)
        , catchError(error => observableThrowError(error)));

    }
}
