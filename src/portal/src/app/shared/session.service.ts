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


import { SessionUser } from './session-user';
import { Member } from '../project/member/member';

import { SignInCredential } from './sign-in-credential';
import { enLang } from '../shared/shared.const';
import { HTTP_FORM_OPTIONS, HTTP_JSON_OPTIONS, HTTP_GET_OPTIONS } from "./shared.utils";

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

    constructor(private http: Http) { }

    // Handle the related exceptions
    handleError(error: any): Promise<any> {
        return Promise.reject(error.message || error);
    }

    // Clear session
    clear(): void {
        this.currentUser = null;
        this.projectMembers = [];
    }

    // Submit signin form to backend (NOT restful service)
    signIn(signInCredential: SignInCredential): Promise<any> {
        // Build the form package
        let queryParam: string = 'principal=' + encodeURIComponent(signInCredential.principal) +
            '&password=' + encodeURIComponent(signInCredential.password);

        // Trigger Http
        return this.http.post(signInUrl, queryParam, HTTP_FORM_OPTIONS)
            .toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    /**
     * Get the related information of current signed in user from backend
     *
     * returns {Promise<SessionUser>}
     *
     * @memberOf SessionService
     */
    retrieveUser(): Promise<SessionUser> {
        return this.http.get(currentUserEndpoint, HTTP_GET_OPTIONS).toPromise()
            .then(response => this.currentUser = response.json() as SessionUser)
            .catch(error => this.handleError(error));
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
    signOff(): Promise<any> {
        return this.http.get(signOffEndpoint, HTTP_GET_OPTIONS).toPromise()
            .then(() => {
                // Destroy current session cache
                // this.currentUser = null;
            })  // Nothing returned
            .catch(error => this.handleError(error));
    }

    /**
     *
     * Update accpunt settings
     *
     *  ** deprecated param {SessionUser} account
     * returns {Promise<any>}
     *
     * @memberOf SessionService
     */
    updateAccountSettings(account: SessionUser): Promise<any> {
        if (!account) {
            return Promise.reject("Invalid account settings");
        }
        let putUrl = accountEndpoint.replace(":id", account.user_id + "");
        return this.http.put(putUrl, JSON.stringify(account), HTTP_JSON_OPTIONS).toPromise()
            .then(() => {
                // Retrieve current session user
                return this.retrieveUser();
            })
            .catch(error => this.handleError(error));
    }

    /**
     *
     * Update accpunt settings
     *
     *  ** deprecated param {SessionUser} account
     * returns {Promise<any>}
     *
     * @memberOf SessionService
     */
    renameAdmin(account: SessionUser): Promise<any> {
        if (!account) {
            return Promise.reject("Invalid account settings");
        }
        return this.http.post(renameAdminEndpoint, JSON.stringify({}), HTTP_JSON_OPTIONS)
            .toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    /**
     * Switch the backend language profile
     */
    switchLanguage(lang: string): Promise<any> {
        if (!lang) {
            return Promise.reject("Invalid language");
        }

        let backendLang = langMap[lang];
        if (!backendLang) {
            backendLang = langMap[enLang];
        }

        let getUrl = langEndpoint + "?lang=" + backendLang;
        return this.http.get(getUrl, HTTP_GET_OPTIONS).toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    checkUserExisting(target: string, value: string): Promise<boolean> {
        // Build the form package
        const body = new URLSearchParams();
        body.set('target', target);
        body.set('value', value);

        // Trigger Http
        return this.http.post(userExistsEndpoint, body.toString(), HTTP_FORM_OPTIONS)
            .toPromise()
            .then(response => {
                return response.json();
            })
            .catch(error => this.handleError(error));
    }

    setProjectMembers(projectMembers: Member[]): void {
        this.projectMembers = projectMembers;
    }

    getProjectMembers(): Member[] {
        return this.projectMembers;
    }

}
