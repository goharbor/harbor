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
import { HttpClient } from '@angular/common/http';
import { map, catchError } from 'rxjs/operators';
import { Observable, throwError as observableThrowError } from 'rxjs';
import { Configuration } from '../base/left-side-nav/config/config';
import {
    CURRENT_BASE_HREF,
    HTTP_GET_OPTIONS,
    HTTP_JSON_OPTIONS,
} from '../shared/units/utils';

const configEndpoint = CURRENT_BASE_HREF + '/configurations';
const emailEndpoint = CURRENT_BASE_HREF + '/email/ping';
const ldapEndpoint = CURRENT_BASE_HREF + '/ldap/ping';
const oidcEndpoint = CURRENT_BASE_HREF + '/system/oidc/ping';

@Injectable({
    providedIn: 'root',
})
export class ConfigurationService {
    constructor(private http: HttpClient) {}

    public getConfiguration(): Observable<Configuration> {
        return this.http.get(configEndpoint, HTTP_GET_OPTIONS).pipe(
            map(response => response as Configuration),
            catchError(error => observableThrowError(error))
        );
    }

    public saveConfiguration(values: any): Observable<any> {
        return this.http
            .put(configEndpoint, JSON.stringify(values), HTTP_JSON_OPTIONS)
            .pipe(
                map(response => response),
                catchError(error => observableThrowError(error))
            );
    }

    public testMailServer(mailSettings: any): Observable<any> {
        return this.http
            .post(
                emailEndpoint,
                JSON.stringify(mailSettings),
                HTTP_JSON_OPTIONS
            )
            .pipe(
                map(response => response),
                catchError(error => observableThrowError(error))
            );
    }

    public testLDAPServer(ldapSettings: any): Observable<any> {
        return this.http
            .post(ldapEndpoint, JSON.stringify(ldapSettings), HTTP_JSON_OPTIONS)
            .pipe(
                map(response => response),
                catchError(error => observableThrowError(error))
            );
    }
    public testOIDCServer(oidcSettings: any): Observable<any> {
        return this.http
            .post(oidcEndpoint, JSON.stringify(oidcSettings), HTTP_JSON_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }
}
