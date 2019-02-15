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
import { Observable, throwError as observableThrowError } from "rxjs";
import { map, catchError, shareReplay } from "rxjs/operators";
import { UserPrivilegeServeItem } from './interface';
import { HttpClient } from '@angular/common/http';



const CACHE_SIZE = 1;
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
}

@Injectable()
export class UserPermissionDefaultService extends UserPermissionService {
    constructor(
        private http: HttpClient,
    ) {
        super();
    }
    private permissionCache: Observable<object>;
    private projectId: number;
    private getPermissionFromBackend(projectId): Observable<object> {
        const userPermissionUrl = `/api/users/current/permissions?scope=/project/${projectId}&relative=true`;
        return this.http.get(userPermissionUrl);
    }
    private processingPermissionResult(responsePermission, resource, action): boolean {
        const permissionList = responsePermission as UserPrivilegeServeItem[];
                for (const privilegeItem of permissionList) {
                    if (privilegeItem.resource === resource && privilegeItem.action === action) {
                        return true;
                    }
                }
                return false;
    }
    public getPermission(projectId, resource, action): Observable<boolean> {

        if (!this.permissionCache || this.projectId !== +projectId) {
            this.projectId = +projectId;
            this.permissionCache = this.getPermissionFromBackend(projectId).pipe(
                shareReplay(CACHE_SIZE));
        }
        return this.permissionCache.pipe(map(response => {
            return this.processingPermissionResult(response, resource, action);
        }))
        .pipe(catchError(error => observableThrowError(error)
        ));
    }
    public clearPermissionCache() {
        this.permissionCache = null;
        this.projectId = null;
    }
}
