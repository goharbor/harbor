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
import { Observable, defer } from 'rxjs';

/**
 * Header name sent to the backend to skip session TTL renewal.
 */
export const HEADER_NO_SESSION_RENEWAL = 'X-Harbor-No-Session-Renewal';

/**
 * Service that coordinates between the skipSessionRenewal() operator
 * and the HTTP interceptor to mark background polling requests so that
 * the backend does not renew the session TTL.
 *
 * Uses a synchronous counter that is safe in JavaScript's single-threaded
 * execution model: the counter is incremented in defer() and decremented
 * in the interceptor — both within the same synchronous call stack.
 */
@Injectable({
    providedIn: 'root',
})
export class SkipSessionRenewalService {
    private _counter = 0;

    /**
     * Called by the operator right before the inner HTTP observable is subscribed.
     */
    begin(): void {
        this._counter++;
    }

    /**
     * Called by the interceptor after reading the flag.
     */
    end(): void {
        if (this._counter > 0) {
            this._counter--;
        }
    }

    /**
     * Returns true if the current request should skip session renewal.
     */
    get shouldSkip(): boolean {
        return this._counter > 0;
    }
}

/**
 * RxJS pipeable operator. Wrap any Observable returned by an ng-swagger-gen
 * service to mark the underlying HTTP request as "no session renewal".
 *
 * Usage:
 *   this.someSwaggerService.someMethod(params)
 *       .pipe(skipSessionRenewal(this.sessionRenewalSkipService))
 *       .subscribe(...);
 */
export function skipSessionRenewal<T>(
    service: SkipSessionRenewalService
): (source: Observable<T>) => Observable<T> {
    return (source: Observable<T>) =>
        defer(() => {
            service.begin();
            return source;
        });
}
