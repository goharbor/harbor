import { Injectable } from '@angular/core';
import {
  CanActivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild
} from '@angular/router';
import { CommonRoutes } from '../../shared/shared.const';
import { AppConfigService } from '../../app-config.service';

@Injectable()
export class ModeGuard implements CanActivate, CanActivateChild {
  constructor(
    private router: Router,
    private appConfigService: AppConfigService) { }
  
  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    //Show the right sign-in page for different modes
    return new Promise((resolve, reject) => {
      if (this.appConfigService.isIntegrationMode()) {
        if (state.url.startsWith(CommonRoutes.SIGN_IN)) {
          this.router.navigate([CommonRoutes.EMBEDDED_SIGN_IN], route.queryParams);
          resolve(false);
        } else {
          resolve(true);
        }
      } else {
        if (state.url.startsWith(CommonRoutes.EMBEDDED_SIGN_IN)) {
          this.router.navigate([CommonRoutes.SIGN_IN], route.queryParams);
          resolve(false);
        } else {
          resolve(true);
        }
      }
    });
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
