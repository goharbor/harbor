
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
  CanActivateChild
} from '@angular/router';
import { AppConfigService } from '../../app-config.service';
import { Observable } from 'rxjs';
import { CommonRoutes } from "../../../lib/entities/shared.const";
import { UserPermissionService } from "../../../lib/services";

@Injectable()
export class OidcGuard implements CanActivate, CanActivateChild {
  constructor(private appConfigService: AppConfigService, private router: Router, private userPermission: UserPermissionService) { }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    // If user has logged in, should not login again
    return new Observable((observer) => {
      // If signout appended
      let queryParams = route.queryParams;
      this.appConfigService.load()
        .subscribe(updatedConfig => {
          if (updatedConfig.auth_mode === 'oidc_auth') {
            return observer.next(true);
          } else {
            this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
            return observer.next(false);
          }
        }
          , error => {
            // Catch the error
            this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
            console.error("Failed to load bootstrap options with error: ", error);
            return observer.next(false);

          });
    });
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
