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
import { CommonRoutes } from '@harbor/ui';
import { AppConfigService } from '../../app-config.service';
import { Observable } from 'rxjs';

@Injectable()
export class SystemAdminGuard implements CanActivate, CanActivateChild {
  constructor(
    private authService: SessionService,
    private router: Router,
    private appConfigService: AppConfigService) { }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    return new Observable((observer) => {
      let user = this.authService.getCurrentUser();
      if (!user) {
        this.authService.retrieveUser()
          .subscribe(() => {
            // updated user
            user = this.authService.getCurrentUser();
            if (user.has_admin_role) {
              return observer.next(true);
            } else {
              this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
              return observer.next(false);
            }
          }, error => {
            // Session retrieving failed.pipe(map redirect to sign-in
            // no matter what status code is.
            // Please pay attention that route 'harborRootRoute' support anonymous user
            if (state.url !== CommonRoutes.HARBOR_ROOT && !state.url.startsWith(CommonRoutes.EMBEDDED_SIGN_IN)) {
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
        if (user.has_admin_role) {
          return observer.next(true);
        } else {
          this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
          return observer.next(false);
        }
      }
    });
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
