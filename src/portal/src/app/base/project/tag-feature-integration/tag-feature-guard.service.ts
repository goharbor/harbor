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
import { ActivatedRouteSnapshot, RouterStateSnapshot } from '@angular/router';
import { Observable } from 'rxjs';
import { UserPrivilegeServeItem } from 'src/app/shared/services/interface';
import { MemberPermissionGuard } from '../../../shared/router-guard/member-permission-guard-activate.service';

@Injectable({
    providedIn: 'root',
})
export class TagFeatureGuardService {
    constructor(private memberPermissionGuard: MemberPermissionGuard) {}

    canActivate(
        route: ActivatedRouteSnapshot,
        state: RouterStateSnapshot
    ): Observable<boolean> | boolean {
        const projectId = route.parent.parent.parent.params['id'];
        const permission = route.data.permissionParam as UserPrivilegeServeItem;
        return this.memberPermissionGuard.checkPermission(
            projectId,
            permission
        );
    }
}
