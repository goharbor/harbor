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
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
import { Repository } from "../../../lib/services";
import { HTTP_GET_OPTIONS } from "../../../lib/utils/utils";

export const topRepoEndpoint = "/api/repositories/top";
/**
 * Declare service to handle the top repositories
 *
 *
 **
 * class GlobalSearchService
 */
@Injectable()
export class TopRepoService {

    constructor(private http: HttpClient) { }

    /**
     * Get top popular repositories
     *
     *  ** deprecated param {string} keyword
     * returns {Observable<TopRepo>}
     *
     * @memberOf GlobalSearchService
     */
    getTopRepos(): Observable<Repository[]> {
        return this.http.get(topRepoEndpoint, HTTP_GET_OPTIONS)
            .pipe(map(response => response as Repository[])
            , catchError(error => observableThrowError(error)));
    }
}
