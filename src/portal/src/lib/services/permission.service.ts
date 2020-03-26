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
import { Observable, forkJoin} from "rxjs";
import { map, share } from "rxjs/operators";
import { HttpClient } from '@angular/common/http';
import { CacheObservable } from "../utils/cache-util";
import { CURRENT_BASE_HREF } from "../utils/utils";


interface Permission {
    resource: string;
    action: string;
}

/**
 * Get System privilege about current backend server.
 * @abstract
 * class UserPermissionService
 */

export abstract class UserPermissionService {
    /**
     *  Get user privilege information.
     *  @abstract
     *  returns
     */
    abstract getPermission(projectId, resource, action);
    abstract clearPermissionCache();
    abstract hasProjectPermission(projectId: any, permission: Permission): Observable<boolean>;
    abstract hasProjectPermissions(projectId: any, permissions: Array<Permission>): Observable<Array<boolean>>;
}

// @dynamic
@Injectable()
export class UserPermissionDefaultService extends UserPermissionService {
    // to prevent duplicate permissions HTTP requests
    private _sharedPermissionObservableMap: {[key: string]: Observable<Array<Permission>>} = {};
    constructor(
        private http: HttpClient,
    ) {
        super();
    }

    @CacheObservable({ maxAge: 1000 * 60 })
    private getPermissions(scope: string, relative?: boolean): Observable<Array<Permission>> {
        const url = `${ CURRENT_BASE_HREF }/users/current/permissions?scope=${scope}&relative=${relative ? 'true' : 'false'}`;
        if (this._sharedPermissionObservableMap[url]) {
            return this._sharedPermissionObservableMap[url];
        } else {
            this._sharedPermissionObservableMap[url] = this.http.get<Array<Permission>>(url).pipe(share());
            return this._sharedPermissionObservableMap[url];
        }
    }

    private hasPermission(permission: Permission, scope: string, relative?: boolean): Observable<boolean> {
        return this.getPermissions(scope, relative).pipe(map(
            (permissions: Array<Permission>) => {
                return permissions.some((p: Permission) => p.resource === permission.resource && p.action === permission.action);
            }
        ));
    }

    private hasPermissions(permissions: Array<Permission>, scope: string, relative?: boolean): Observable<Array<boolean>> {
        return forkJoin(permissions.map((permission) => this.hasPermission(permission, scope, relative)));
    }

    public hasProjectPermission(projectId: any, permission: Permission): Observable<boolean> {
        return this.hasPermission(permission, `/project/${projectId}`, true);
    }

    public hasProjectPermissions(projectId: any, permissions: Array<Permission>): Observable<Array<boolean>> {
        return this.hasPermissions(permissions, `/project/${projectId}`, true);
    }

    public getPermission(projectId: any, resource: string, action: string): Observable<boolean> {
        return this.hasProjectPermission(projectId, { resource, action });
    }

    public clearPermissionCache() {
    }
}
