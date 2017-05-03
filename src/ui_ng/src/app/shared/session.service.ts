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

import { SessionUser } from './session-user';
import { Member } from '../project/member/member';

import { SignInCredential } from './sign-in-credential';
import { enLang } from '../shared/shared.const'

const signInUrl = '/login';
const currentUserEndpint = "/api/users/current";
const signOffEndpoint = "/log_out";
const accountEndpoint = "/api/users/:id";
const langEndpoint = "/language";
const userExistsEndpoint = "/userExists";
const langMap = {
    "zh": "zh-CN",
    "en": "en-US"
};

/**
 * Define related methods to handle account and session corresponding things
 * 
 * @export
 * @class SessionService
 */
@Injectable()
export class SessionService {
    currentUser: SessionUser = null;

    projectMembers: Member[];

    private headers = new Headers({
        "Content-Type": 'application/json'
    });

    private formHeaders = new Headers({
        "Content-Type": 'application/x-www-form-urlencoded'
    });

    constructor(private http: Http) { }

    //Handle the related exceptions
    private handleError(error: any): Promise<any> {
        return Promise.reject(error.message || error);
    }

    //Clear session
    clear(): void {
        this.currentUser = null;
        this.projectMembers = [];
    }

    //Submit signin form to backend (NOT restful service)
    signIn(signInCredential: SignInCredential): Promise<any> {
        //Build the form package
        let queryParam:string = 'principal=' + encodeURIComponent(signInCredential.principal) + 
        '&password=' + encodeURIComponent(signInCredential.password);

        //Trigger Http
        return this.http.post(signInUrl, queryParam, { headers: this.formHeaders })
            .toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    /**
     * Get the related information of current signed in user from backend
     * 
     * @returns {Promise<SessionUser>}
     * 
     * @memberOf SessionService
     */
    retrieveUser(): Promise<SessionUser> {
        return this.http.get(currentUserEndpint, { headers: this.headers }).toPromise()
            .then(response => this.currentUser = response.json() as SessionUser)
            .catch(error => this.handleError(error))
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
        return this.http.get(signOffEndpoint, { headers: this.headers }).toPromise()
            .then(() => {
                //Destroy current session cache
                //this.currentUser = null;
            }) //Nothing returned
            .catch(error => this.handleError(error))
    }

    /**
     * 
     * Update accpunt settings
     * 
     * @param {SessionUser} account
     * @returns {Promise<any>}
     * 
     * @memberOf SessionService
     */
    updateAccountSettings(account: SessionUser): Promise<any> {
        if (!account) {
            return Promise.reject("Invalid account settings");
        }
        let putUrl = accountEndpoint.replace(":id", account.user_id + "");
        return this.http.put(putUrl, JSON.stringify(account), { headers: this.headers }).toPromise()
            .then(() => {
                //Retrieve current session user
                return this.retrieveUser();
            })
            .catch(error => this.handleError(error))
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
        return this.http.get(getUrl).toPromise()
            .then(() => null)
            .catch(error => this.handleError(error))
    }

    checkUserExisting(target: string, value: string): Promise<boolean> {
        //Build the form package
        const body = new URLSearchParams();
        body.set('target', target);
        body.set('value', value);

        //Trigger Http
        return this.http.post(userExistsEndpoint, body.toString(), { headers: this.formHeaders })
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