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
import { HttpClient } from '@angular/common/http';
import { Observable, of } from 'rxjs';
import { map, catchError, tap } from 'rxjs/operators';

const SETUP_STATUS_URL = '/c/setup/status';
const SETUP_URL = '/c/setup';

export interface SetupStatusResponse {
    setup_required: boolean;
}

@Injectable({
    providedIn: 'root',
})
export class SetupService {
    private cachedStatus: boolean | null = null;

    constructor(private http: HttpClient) {}

    /**
     * Check if one-time admin setup is required.
     * Caches the result to avoid repeated calls.
     */
    isSetupRequired(): Observable<boolean> {
        if (this.cachedStatus !== null) {
            return of(this.cachedStatus);
        }
        return this.http.get<SetupStatusResponse>(SETUP_STATUS_URL).pipe(
            map(res => res.setup_required),
            tap(val => (this.cachedStatus = val)),
            catchError(() => of(false))
        );
    }

    /**
     * Submit the admin password via POST /c/setup.
     */
    setupAdminPassword(password: string): Observable<any> {
        return this.http.post(SETUP_URL, { password }).pipe(
            tap(() => {
                // After successful setup, update cache
                this.cachedStatus = false;
            })
        );
    }

    /**
     * Clear cached status (e.g. after setup completes).
     */
    clearCache(): void {
        this.cachedStatus = null;
    }
}
