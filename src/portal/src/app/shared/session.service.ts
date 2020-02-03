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
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
import { SessionUser, SessionUserBackend } from './session-user';
import { Member } from '../project/member/member';
import { SignInCredential } from './sign-in-credential';
import { enLang } from './shared.const';
import { SessionViewmodelFactory } from './session.viewmodel.factory';
import { HTTP_FORM_OPTIONS, HTTP_GET_OPTIONS, HTTP_JSON_OPTIONS, clone } from "../../lib/utils/utils";
import { FlushAll } from "../../lib/utils/cache-util";

const signInUrl = '/c/login';
const currentUserEndpoint = "/api/users/current";
const signOffEndpoint = "/c/log_out";
const accountEndpoint = "/api/users/:id";
const langEndpoint = "/language";
const userExistsEndpoint = "/c/userExists";
const renameAdminEndpoint = '/api/internal/renameadmin';
const langMap = {
    "zh": "zh-CN",
    "en": "en-US"
};

/**
 * Define related methods to handle account and session corresponding things
 *
 **
 * class SessionService
 */
@Injectable()
export class SessionService {
    currentUser: SessionUser = null;

    projectMembers: Member[];

    /*formHeaders = new Headers({
        "Content-Type": 'application/x-www-form-urlencoded'
    });*/

    constructor(private http: HttpClient, public sessionViewmodel: SessionViewmodelFactory) { }

    // Handle the related exceptions
    handleError(error: any): Observable<any> {
        return observableThrowError(error.error || error);
    }

    // Clear session
    clear(): void {
        this.currentUser = null;
        this.projectMembers = [];
        FlushAll();
    }

    // Submit signin form to backend (NOT restful service)
    signIn(signInCredential: SignInCredential): Observable<any> {
        // Build the form package
        let queryParam: string = 'principal=' + encodeURIComponent(signInCredential.principal) +
            '&password=' + encodeURIComponent(signInCredential.password);

        // Trigger HttpClient
        return this.http.post(signInUrl, queryParam, HTTP_FORM_OPTIONS)
            .pipe(map(() => null)
            , catchError(error => observableThrowError(error)));
    }

    /**
     * Get the related information of current signed in user from backend
     *
     * returns {Observable<SessionUser>}
     *
     * @memberOf SessionService
     */
    retrieveUser(): Observable<SessionUserBackend> {
        return this.http.get(currentUserEndpoint, HTTP_GET_OPTIONS)
            .pipe(map((response: SessionUserBackend) => this.currentUser = this.sessionViewmodel.getCurrentUser(response) as SessionUser)
            , catchError(error => this.handleError(error)));
    }
    /**
     * For getting info
     */
    getCurrentUser(): SessionUser {
        return this.currentUser;
    }

    /**
     * Log out the system
     */
    signOff(): Observable<any> {
        return this.http.get(signOffEndpoint, HTTP_GET_OPTIONS)
            .pipe(map(() => {
                // Destroy current session cache
                // this.currentUser = null;
            })  // Nothing returned
            , catchError(error => this.handleError(error)));
    }

    /**
     *
     * Update accpunt settings
     *
     *  ** deprecated param {SessionUser} account
     * returns {Observable<any>}
     *
     * @memberOf SessionService
     */
    updateAccountSettings(account: SessionUser): Observable<any> {
        if (!account) {
            return observableThrowError("Invalid account settings");
        }
        let putUrl = accountEndpoint.replace(":id", account.user_id + "");
        return this.http.put(putUrl, JSON.stringify(account), HTTP_JSON_OPTIONS)
            .pipe(map(() => {
                // Retrieve current session user
                return this.retrieveUser();
            })
            , catchError(error => this.handleError(error)));
    }

    /**
     *
     * Update accpunt settings
     *
     *  ** deprecated param {SessionUser} account
     * returns {Observable<any>}
     *
     * @memberOf SessionService
     */
    renameAdmin(account: SessionUser): Observable<any> {
        if (!account) {
            return observableThrowError("Invalid account settings");
        }
        return this.http.post(renameAdminEndpoint, JSON.stringify({}), HTTP_JSON_OPTIONS)
            .pipe(map(() => null)
            , catchError(error => this.handleError(error)));
    }

    /**
     * Switch the backend language profile
     */
    switchLanguage(lang: string): Observable<any> {
        if (!lang) {
            return observableThrowError("Invalid language");
        }

        let backendLang = langMap[lang];
        if (!backendLang) {
            backendLang = langMap[enLang];
        }

        let getUrl = langEndpoint + "?lang=" + backendLang;
        return this.http.get(getUrl, HTTP_GET_OPTIONS)
            .pipe(map(() => null)
            , catchError(error => this.handleError(error)));
    }

    checkUserExisting(target: string, value: string): Observable<boolean> {
        // Build the form package
        let body = new HttpParams();
        body = body.set('target', target);
        body = body.set('value', value);

        // Trigger HttpClient
        return this.http.post(userExistsEndpoint, body.toString(), HTTP_FORM_OPTIONS)
            .pipe(catchError(error => this.handleError(error)));
    }

    setProjectMembers(projectMembers: Member[]): void {
        this.projectMembers = projectMembers;
    }

    getProjectMembers(): Member[] {
        return this.projectMembers;
    }

}
