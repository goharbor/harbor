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
import { Injectable } from '@angular/core';
import {
    HttpInterceptor,
    HttpRequest,
    HttpHandler,
    HttpResponse,
    HttpErrorResponse,
} from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError, tap } from 'rxjs/operators';
import { errorHandler } from '../shared/units/shared.utils';

export const SAFE_METHODS: string[] = ['GET', 'HEAD', 'OPTIONS', 'TRACE'];

enum INVALID_CSRF_TOKEN {
    CODE = 403,
    MESSAGE = 'CSRF token invalid',
}

@Injectable({
    providedIn: 'root',
})
export class InterceptHttpService implements HttpInterceptor {
    constructor() {}

    intercept(request: HttpRequest<any>, next: HttpHandler): Observable<any> {
        // Get the csrf token from localstorage
        const token = localStorage.getItem('__csrf');
        if (token) {
            // Clone the request and replace the original headers with
            // cloned headers, updated with the csrf token.
            // not for requests using safe methods
            if (
                request.method &&
                SAFE_METHODS.indexOf(request.method.toUpperCase()) === -1
            ) {
                request = request.clone({
                    headers: request.headers.set('X-Harbor-CSRF-Token', token),
                });
            }
        }
        return next
            .handle(request)
            .pipe(
                tap(
                    response => {
                        if (
                            response &&
                            response instanceof HttpResponse &&
                            response.headers
                        ) {
                            const responseToken: string = response.headers.get(
                                'X-Harbor-CSRF-Token'
                            );
                            if (responseToken) {
                                localStorage.setItem('__csrf', responseToken);
                            }
                        }
                    },
                    error => {
                        if (error && error.headers) {
                            const responseToken: string = error.headers.get(
                                'X-Harbor-CSRF-Token'
                            );
                            if (responseToken) {
                                localStorage.setItem('__csrf', responseToken);
                            }
                        }
                    }
                )
            )
            .pipe(
                catchError(error => {
                    // handle 504 error in document format from backend
                    if (error && error.status === 504) {
                        // throw 504 error in json format
                        return throwError(
                            new HttpErrorResponse({
                                error: '504 gateway timeout',
                                status: 504,
                            })
                        );
                    }
                    if (
                        error.status === INVALID_CSRF_TOKEN.CODE &&
                        errorHandler(error) === INVALID_CSRF_TOKEN.MESSAGE
                    ) {
                        const csrfToken = localStorage.getItem('__csrf');
                        if (csrfToken) {
                            request = request.clone({
                                headers: request.headers.set(
                                    'X-Harbor-CSRF-Token',
                                    csrfToken
                                ),
                            });
                            return next.handle(request);
                        }
                    }
                    return throwError(error);
                })
            );
    }
}
