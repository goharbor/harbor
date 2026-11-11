// Copyright Project Harbor Authors
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
import { ElementRef, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { map, catchError } from 'rxjs/operators';
import { Observable, throwError as observableThrowError } from 'rxjs';
import { SystemCVEAllowlist, SystemInfo } from './interface';
import {
    CURRENT_BASE_HREF,
    HTTP_GET_OPTIONS,
    HTTP_JSON_OPTIONS,
} from '../units/utils';

/**
 * Get System information about current backend server.
 * @abstract
 * class
 */
export abstract class SystemInfoService {
    /**
     *  Get global system information.
     *  @abstract
     *  returns
     */
    abstract getSystemInfo(): Observable<SystemInfo>;
    /**
     *  get system CEVAllowlist
     */
    abstract getSystemAllowlist(): Observable<SystemCVEAllowlist>;
    /**
     *  update systemCVEAllowlist
     * @param systemCVEAllowlist
     */
    abstract updateSystemAllowlist(
        systemCVEAllowlist: SystemCVEAllowlist
    ): Observable<any>;
    /**
     *  set null to the date type input
     * @param ref
     */
    abstract resetDateInput(ref: ElementRef);
}

@Injectable()
export class SystemInfoDefaultService extends SystemInfoService {
    constructor(private http: HttpClient) {
        super();
    }
    getSystemInfo(): Observable<SystemInfo> {
        let url = CURRENT_BASE_HREF + '/systeminfo';
        return this.http.get(url, HTTP_GET_OPTIONS).pipe(
            map(systemInfo => systemInfo as SystemInfo),
            catchError(error => observableThrowError(error))
        );
    }
    public getSystemAllowlist(): Observable<SystemCVEAllowlist> {
        return this.http
            .get(CURRENT_BASE_HREF + '/system/CVEAllowlist', HTTP_GET_OPTIONS)
            .pipe(
                map(
                    systemCVEAllowlist =>
                        systemCVEAllowlist as SystemCVEAllowlist
                ),
                catchError(error => observableThrowError(error))
            );
    }
    public updateSystemAllowlist(
        systemCVEAllowlist: SystemCVEAllowlist
    ): Observable<any> {
        return this.http
            .put(
                CURRENT_BASE_HREF + '/system/CVEAllowlist',
                JSON.stringify(systemCVEAllowlist),
                HTTP_JSON_OPTIONS
            )
            .pipe(
                map(response => response),
                catchError(error => observableThrowError(error))
            );
    }
    public resetDateInput(ref: ElementRef) {
        if (ref) {
            ref.nativeElement.value = null;
        }
    }
}
