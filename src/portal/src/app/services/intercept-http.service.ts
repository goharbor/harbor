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
    HttpErrorResponse,
} from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import {
    SkipSessionRenewalService,
    HEADER_NO_SESSION_RENEWAL,
} from './skip-session-renewal.service';

@Injectable({
    providedIn: 'root',
})
export class InterceptHttpService implements HttpInterceptor {
    constructor(private skipSessionRenewalService: SkipSessionRenewalService) {}

    intercept(request: HttpRequest<any>, next: HttpHandler): Observable<any> {
        // Check if the current request should skip session renewal.
        // The flag was set synchronously by skipSessionRenewal() operator
        // within the same call stack.
        if (this.skipSessionRenewalService.shouldSkip) {
            this.skipSessionRenewalService.end();
            request = request.clone({
                headers: request.headers.set(HEADER_NO_SESSION_RENEWAL, 'true'),
            });
        }

        return next.handle(request).pipe(
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
                return throwError(error);
            })
        );
    }
}
