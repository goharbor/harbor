import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class ArtifactGuardActivateService {

  constructor() { }
}
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
import {
  CanActivate, Router,
  ActivatedRouteSnapshot,
  RouterStateSnapshot,
  CanActivateChild
} from '@angular/router';
import { SessionService } from '../../shared/session.service';
import { Observable, of } from 'rxjs';
import { map, catchError, switchMap } from 'rxjs/operators';
import { ProjectService, ArtifactService } from "../../../lib/services";
import { CommonRoutes } from "../../../lib/entities/shared.const";

@Injectable()
export class ArtifactGuard implements CanActivate, CanActivateChild {
  constructor(
    private sessionService: SessionService,
    private artifactService: ArtifactService,
    private projectService: ProjectService,
    private router: Router) { }

  canActivate(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    const projectId = route.params['id'];
    const repoName = route.params['repo'];
    const digest = route.params['digest'];
    return this.projectService.getProject(projectId).pipe(
      switchMap((project) => {
        return this.hasArtifactPerm(project.name, repoName, digest);
      }),
      catchError(err => {
        this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
        return of(false);
      })
    );
  }

  canActivateChild(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Observable<boolean> | boolean {
    return this.canActivate(route, state);
  }

  hasArtifactPerm(projectName: string, repoName: string, digest): Observable<boolean> {
    // Note: current user will have the permission to visit the project when the user can get response from GET /projects/:id API.
    return this.artifactService.getArtifactFromDigest(projectName, repoName, digest).pipe(
      () => {
        return of(true);
      },
      catchError(err => {
        this.router.navigate([CommonRoutes.HARBOR_DEFAULT]);
        return of(false);
      })
    );
  }
}
