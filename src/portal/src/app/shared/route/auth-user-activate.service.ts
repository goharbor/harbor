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
  CanActivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild,
  NavigationExtras
} from '@angular/router';
import { SessionService } from '../../shared/session.service';
import { AdmiralQueryParamKey } from '../../shared/shared.const';
import { CommonRoutes } from '@harbor/ui';
import { AppConfigService } from '../../app-config.service';
import { maintainUrlQueryParmas } from '../../shared/shared.utils';
import { MessageHandlerService } from '../message-handler/message-handler.service';
import { SearchTriggerService } from '../../base/global-search/search-trigger.service';
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError, observable } from "rxjs";
@Injectable()
export class AuthCheckGuard implements CanActivate, CanActivateChild {
  constructor(
    private authService: SessionService,
    private router: Router,
    private appConfigService: AppConfigService,
    private msgHandler: MessageHandlerService,
    private searchTrigger: SearchTriggerService) { }

  isGuest(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): boolean {
    const proRegExp = /\/harbor\/projects\/[\d]+\/.+/i;
    const libRegExp = /\/harbor\/tags\/[\d]+\/.+/i;
    if (proRegExp.test(state.url) || libRegExp.test(state.url)) {
      return true;
    }

    return false;
  }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    // When routing change, clear
    this.msgHandler.clear();
    if (this.appConfigService.getConfig().read_only.toString() === 'true') {
      this.msgHandler.handleReadOnly();
    }

    this.searchTrigger.closeSearch(true);
    return new Observable(observer => {
      let queryParams = route.queryParams;
      if (queryParams) {
        if (queryParams[AdmiralQueryParamKey]) {
          this.appConfigService.saveAdmiralEndpoint(queryParams[AdmiralQueryParamKey]);
          // Remove the query parameter key pair and redirect
          let keyRemovedUrl = maintainUrlQueryParmas(state.url, AdmiralQueryParamKey, undefined);
          if (!/[?]{1}.+/i.test(keyRemovedUrl)) {
            keyRemovedUrl = keyRemovedUrl.replace('?', '');
          }

          this.router.navigateByUrl(keyRemovedUrl);
          return observer.next(false);
        }
      }
      let user = this.authService.getCurrentUser();
      if (!user) {
        this.authService.retrieveUser()
          .subscribe(() => {
            return observer.next(true);
          }
            , error => {
              // If is guest, skip it
              if (this.isGuest(route, state)) {
                return observer.next(true);
              }
              // Session retrieving failed then redirect to sign-in
              // no matter what status code is.
              // Please pay attention that route 'HARBOR_ROOT' and 'EMBEDDED_SIGN_IN' support anonymous user
              if (!state.url.startsWith(CommonRoutes.EMBEDDED_SIGN_IN)) {
                let navigatorExtra: NavigationExtras = {
                  queryParams: { "redirect_url": state.url }
                };
                this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], navigatorExtra);
                return observer.next(false);
              } else {
                return observer.next(true);
              }
            });
      } else {
        return observer.next(true);
      }
    });
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
