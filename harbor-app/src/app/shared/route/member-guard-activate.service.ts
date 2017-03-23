import { Injectable } from '@angular/core';
import {
  CanActivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild
} from '@angular/router';
import { SessionService } from '../../shared/session.service';
import { ProjectService } from '../../project/project.service';
import { CommonRoutes } from '../../shared/shared.const';

@Injectable()
export class MemberGuard implements CanActivate, CanActivateChild {
  constructor(
    private sessionService: SessionService,
    private projectService: ProjectService, 
    private router: Router) {}

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    let projectId: number = route.params['id'];
    return new Promise((resolve, reject) => {
        this.projectService.checkProjectMember(projectId)
          .subscribe(
            res=>{
              this.sessionService.setProjectMembers(res);
              return resolve(true)
            },
            error => {
              this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
              return resolve(false);
            });
    });
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
