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
import { Http } from '@angular/http';


import {HTTP_JSON_OPTIONS, HTTP_GET_OPTIONS} from "../shared/shared.utils";
import { User, LDAPUser } from './user';
import LDAPUsertoUser from './user';

const userMgmtEndpoint = '/api/users';
const userListSearch = '/api/users/search';
const ldapUserEndpoint = '/api/ldap/users';

/**
 * Define related methods to handle account and session corresponding things
 *
 **
 * class SessionService
 */
@Injectable()
export class UserService {

    constructor(private http: Http) { }

    // Handle the related exceptions
    handleError(error: any): Promise<any> {
        return Promise.reject(error.message || error);
    }

    // Get the user list
    getUsersNameList(): Promise<User[]> {
        return this.http.get(userListSearch, HTTP_GET_OPTIONS).toPromise()
            .then(response => response.json() as User[])
            .catch(error => this.handleError(error));
    }
    getUsers(): Promise<User[]> {
        return this.http.get(userMgmtEndpoint, HTTP_GET_OPTIONS).toPromise()
            .then(response => response.json() as User[])
            .catch(error => this.handleError(error));
    }

    // Add new user
    addUser(user: User): Promise<any> {
        return this.http.post(userMgmtEndpoint, JSON.stringify(user), HTTP_JSON_OPTIONS).toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    // Delete the specified user
    deleteUser(userId: number): Promise<any> {
        return this.http.delete(userMgmtEndpoint + "/" + userId, HTTP_JSON_OPTIONS)
            .toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    // Update user to enable/disable the admin role
    updateUser(user: User): Promise<any> {
        return this.http.put(userMgmtEndpoint + "/" + user.user_id, JSON.stringify(user), HTTP_JSON_OPTIONS)
            .toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    // Set user admin role
    updateUserRole(user: User): Promise<any> {
        return this.http.put(userMgmtEndpoint + "/" + user.user_id + "/sysadmin", JSON.stringify(user), HTTP_JSON_OPTIONS)
            .toPromise()
            .then(() => null)
            .catch(error => this.handleError(error));
    }

    // admin change normal user pwd
    changePassword(uid: number, newPassword: string, confirmPwd: string): Promise<any> {
        if (!uid || !newPassword) {
            return Promise.reject("Invalid change uid or password");
        }

        return this.http.put(userMgmtEndpoint + '/' + uid + '/password',
            {
                "old_password": newPassword,
                'new_password': confirmPwd
            },
            HTTP_JSON_OPTIONS)
            .toPromise()
            .then(response => response)
            .catch(error => {
                return Promise.reject(error);
            });
    }

    // Get User from LDAP
    getLDAPUsers(username: string): Promise<User[]> {
        return this.http.get(`${ldapUserEndpoint}/search?username=${username}`, HTTP_GET_OPTIONS)
        .toPromise()
        .then(response => {
            let ldapUser = response.json() as LDAPUser[] || [];
            return ldapUser.map(u => LDAPUsertoUser(u));
        })
        .catch( error => this.handleError(error));
    }

    importLDAPUsers(usernames: string[]): Promise<any> {
        return this.http.post(`${ldapUserEndpoint}/import`, JSON.stringify({ldap_uid_list: usernames}), HTTP_JSON_OPTIONS)
        .toPromise()
        .then(() => null )
        .catch(err => this.handleError(err));
    }
}
