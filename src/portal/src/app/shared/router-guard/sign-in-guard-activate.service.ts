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
    Router,
    ActivatedRouteSnapshot,
    RouterStateSnapshot,
} from '@angular/router';
import { SessionService } from '../services/session.service';
import { Observable } from 'rxjs';
import { CommonRoutes } from '../entities/shared.const';
import { AppConfigService } from '../../services/app-config.service';

@Injectable({
    providedIn: 'root',
})
export class SignInGuard {
    constructor(
        private authService: SessionService,
        private router: Router,
        private appCfgService: AppConfigService
    ) {}
    canActivate(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<boolean> | boolean {
        // If user has logged in, should not login again
        return new Observable(observer => {
            // If signout appended
            let queryParams = route.queryParams;
            if (queryParams && queryParams['signout']) {
                this.authService.signOff().subscribe(
                    () => {
                        this.authService.clear(); // Destroy session cache
                        this.appCfgService.resetVersion(); // reset version info for anoymous users

                        return observer.next(true);
                    },
                    error => {
                        console.error(error);
                        return observer.next(false);
                    }
                );
            } else {
                let user = this.authService.getCurrentUser();
                if (user === null) {
                    this.authService.retrieveUser().subscribe(
                        () => {
                            this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
                            return observer.next(false);
                        },
                        error => {
                            return observer.next(true);
                        }
                    );
                } else {
                    this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
                    return observer.next(false);
                }
            }
        });
    }

    canActivateChild(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<boolean> | boolean {
        return this.canActivate(route, state);
    }
}
