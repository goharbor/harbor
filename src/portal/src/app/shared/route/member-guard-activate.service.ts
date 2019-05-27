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
import { SessionService } from '../../shared/session.service';
import { ProjectService } from '@harbor/ui';
import { CommonRoutes } from '@harbor/ui';
import { Observable } from 'rxjs';

@Injectable()
export class MemberGuard implements CanActivate, CanActivateChild {
  constructor(
    private sessionService: SessionService,
    private projectService: ProjectService,
    private router: Router) {}

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    let projectId = route.params['id'];
    this.sessionService.setProjectMembers([]);
    return new Observable((observer) => {
      let user = this.sessionService.getCurrentUser();
      if (user === null) {
        this.sessionService.retrieveUser()
        .subscribe(() => {
          this.checkMemberStatus(state.url, projectId).subscribe((res) => observer.next(res));
        }
        , error => {
          this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
          observer.next(false);
        });
      } else {
        this.checkMemberStatus(state.url, projectId).subscribe((res) => observer.next(res));
      }
    });
  }

  checkMemberStatus(url: string, projectId: number): Observable<boolean> {
    return new Observable<boolean>((observer) => {
      this.projectService.checkProjectMember(projectId)
      .subscribe(res => {
        this.sessionService.setProjectMembers(res);
        return observer.next(true);
      },
      () => {
        // Add exception for repository in project detail router activation.
        this.projectService.getProject(projectId).subscribe(project => {
          if (project.metadata && project.metadata.public === 'true') {
            return observer.next(true);
          }
          this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
          return observer.next(false);
        },
        () => {
          this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
          return observer.next(false);
        });
      });
    });
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    return this.canActivate(route, state);
  }
}
