import { Injectable } from '@angular/core';
import {
  CanActivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild
} from '@angular/router';
import { SessionService } from '../../shared/session.service';
import { CommonRoutes } from '../../shared/shared.const';

@Injectable()
export class StartGuard implements CanActivate, CanActivateChild {
  constructor(private authService: SessionService, private router: Router) { }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    //Authentidacted user should not see the start page now
    return new Promise((resolve, reject) => {
      let user = this.authService.getCurrentUser();
      if (!user) {
        this.authService.retrieveUser()
          .then(() => {
            this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
            return resolve(false);
          })
          .catch(error => {
            return resolve(true);
          });
      } else {
        this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
        return resolve(false);
      }
    });
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
