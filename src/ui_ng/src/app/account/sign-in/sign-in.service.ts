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
import { Headers, Http, URLSearchParams } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { SignInCredential } from '../../shared/sign-in-credential';

const signInUrl = '/login';
/**
 * 
 * Define a service to provide sign in methods
 * 
 * @export
 * @class SignInService
 */
@Injectable()
export class SignInService {
    private headers = new Headers({
        "Content-Type": 'application/x-www-form-urlencoded'
    });

    constructor(private http: Http) {}

    //Handle the related exceptions
    private handleError(error: any): Promise<any>{
        return Promise.reject(error.message || error);
    }

    //Submit signin form to backend (NOT restful service)
    signIn(signInCredential: SignInCredential): Promise<any>{
        //Build the form package
        const body = new URLSearchParams();
        body.set('principal', signInCredential.principal);
        body.set('password', signInCredential.password);

        //Trigger Http
        return this.http.post(signInUrl, body.toString(), { headers: this.headers })
        .toPromise()
        .then(()=>null)
        .catch(this.handleError);
    }
}