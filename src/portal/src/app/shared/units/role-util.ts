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
import { forkJoin, Observable, of } from 'rxjs';
import { map, switchMap } from 'rxjs/operators';
import { RoleService } from '../../../../ng-swagger-gen/services/role.service';
import { Role } from '../../../../ng-swagger-gen/models/role';

const ROLE_PAGE_SIZE = 100;

/**
 * Fetches every page of roles. Custom-role creation is not capped at 100, so the
 * member/group role pickers must aggregate all pages rather than treating the
 * first page as exhaustive (otherwise roles past the 100th silently disappear).
 */
export function getAllRoles(roleService: RoleService): Observable<Role[]> {
    return roleService
        .ListRoleResponse({ page: 1, pageSize: ROLE_PAGE_SIZE })
        .pipe(
            switchMap(resp => {
                const first = (resp.body as Role[]) ?? [];
                const total = Number.parseInt(
                    resp.headers.get('x-total-count') ?? '0',
                    10
                );
                const totalPages = Math.ceil(total / ROLE_PAGE_SIZE);
                if (totalPages <= 1) {
                    return of(first);
                }
                const rest: Observable<Role[]>[] = [];
                for (let page = 2; page <= totalPages; page++) {
                    rest.push(
                        roleService.ListRole({ page, pageSize: ROLE_PAGE_SIZE })
                    );
                }
                return forkJoin(rest).pipe(map(pages => first.concat(...pages)));
            })
        );
}
