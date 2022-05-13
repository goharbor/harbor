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
import { Observable } from 'rxjs';
import { map, catchError } from 'rxjs/operators';
import { ReplicationService } from '../../../../ng-swagger-gen/services';
import { ReplicationExecution } from '../../../../ng-swagger-gen/models/replication-execution';

@Injectable({
    providedIn: 'root',
})
export class ReplicationTasksRoutingResolverService
    implements Resolve<ReplicationExecution>
{
    constructor(
        private replicationService: ReplicationService,
        private router: Router
    ) {}

    resolve(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<ReplicationExecution> | any {
        // Support both parameters and query parameters
        let executionId = route.params['id'];
        if (!executionId) {
            executionId = route.queryParams['project_id'];
        }
        return this.replicationService
            .getReplicationExecution({
                id: +executionId,
            })
            .pipe(
                map((res: ReplicationExecution) => {
                    if (!res) {
                        this.router.navigate(['/harbor', 'projects']);
                    }
                    return res;
                }),
                catchError(error => {
                    this.router.navigate(['/harbor', 'projects']);
                    return null;
                })
            );
    }
}
