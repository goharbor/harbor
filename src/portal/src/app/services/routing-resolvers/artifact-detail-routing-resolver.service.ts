// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
    Resolve,
    RouterStateSnapshot,
    ActivatedRouteSnapshot,
} from '@angular/router';
import { forkJoin, Observable, of } from 'rxjs';
import { map, catchError, mergeMap } from 'rxjs/operators';
import { Artifact } from '../../../../ng-swagger-gen/models/artifact';
import { ArtifactService } from '../../../../ng-swagger-gen/services/artifact.service';
import { Project } from '../../base/project/project';
import { ProjectService } from '../../shared/services';
import { dbEncodeURIComponent } from '../../shared/units/utils';

@Injectable({
    providedIn: 'root',
})
export class ArtifactDetailRoutingResolverService implements Resolve<Artifact> {
    constructor(
        private projectService: ProjectService,
        private artifactService: ArtifactService,
        private router: Router
    ) {}

    resolve(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<Artifact> | any {
        const projectId: string = route.params['id'];
        const repositoryName: string = route.params['repo'];
        const artifactDigest: string = route.params['digest'];
        return this.projectService.getProject(projectId).pipe(
            mergeMap((project: Project) => {
                return forkJoin([
                    this.artifactService.getArtifact({
                        repositoryName: dbEncodeURIComponent(repositoryName),
                        reference: artifactDigest,
                        projectName: project.name,
                        withLabel: true,
                        withScanOverview: true,
                        withTag: false,
                        withImmutableStatus: true,
                    }),
                    of(project),
                ]);
            }),
            catchError(error => {
                this.router.navigate(['/harbor', 'projects']);
                return null;
            })
        );
    }
}
