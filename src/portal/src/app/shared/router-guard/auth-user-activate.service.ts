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
    NavigationExtras,
} from '@angular/router';
import { SessionService } from '../services/session.service';
import { AppConfigService } from '../../services/app-config.service';
import { MessageHandlerService } from '../services/message-handler.service';
import { SearchTriggerService } from '../components/global-search/search-trigger.service';
import { Observable } from 'rxjs';
import { UN_LOGGED_PARAM, YES } from '../../account/sign-in/sign-in.service';
import { CommonRoutes, CONFIG_AUTH_MODE } from '../entities/shared.const';

@Injectable({
    providedIn: 'root',
})
export class AuthCheckGuard {
    constructor(
        private authService: SessionService,
        private router: Router,
        private appConfigService: AppConfigService,
        private msgHandler: MessageHandlerService,
        private searchTrigger: SearchTriggerService
    ) {}

    canActivate(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<boolean> | boolean {
        // When routing change, clear
        this.msgHandler.clear();
        this.searchTrigger.closeSearch(true);
        return new Observable(observer => {
            // if the url has the queryParam `publicAndNotLogged=yes`, then skip auth check
            const queryParams = route.queryParams;
            if (queryParams && queryParams[UN_LOGGED_PARAM] === YES) {
                return observer.next(true);
            }
            let user = this.authService.getCurrentUser();
            if (!user) {
                this.authService.retrieveUser().subscribe(
                    () => {
                        return observer.next(true);
                    },
                    error => {
                        // Session retrieving failed then redirect to sign-in
                        // no matter what status code is.
                        // no need to check auth for `sign in` page
                        if (
                            !state.url.startsWith(CommonRoutes.EMBEDDED_SIGN_IN)
                        ) {
                            let navigatorExtra: NavigationExtras = {
                                queryParams: { redirect_url: state.url },
                            };
                            // if primary auth mode or auto login enabled, skip the first step
                            if (
                                this.appConfigService.getConfig().auth_mode ==
                                    CONFIG_AUTH_MODE.OIDC_AUTH &&
                                (this.appConfigService.getConfig()
                                    .primary_auth_mode ||
                                 this.appConfigService.getConfig()
                                    .oidc_auto_login)
                            ) {
                                window.location.href =
                                    '/c/oidc/login?redirect_url=' +
                                    encodeURI(state.url);
                                return observer.next(false);
                            }
                            this.router.navigate(
                                [CommonRoutes.EMBEDDED_SIGN_IN],
                                navigatorExtra
                            );
                            return observer.next(false);
                        } else {
                            return observer.next(true);
                        }
                    }
                );
            } else {
                return observer.next(true);
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
