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
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { of } from 'rxjs';
import { Role } from '../../../../ng-swagger-gen/models/role';
import { RoleService } from '../../../../ng-swagger-gen/services/role.service';
import { getAllRoles, roleDisplayName } from './role-util';

function role(id: number): Role {
    return { id, name: `role-${id}`, is_builtin: false };
}

/**
 * A role service holding `total` roles, numbered 1..total and served in pages of
 * 100, that records the parameters of every call it receives.
 */
function fakeRoleService(total: number) {
    const all: Role[] = [];
    for (let id = 1; id <= total; id++) {
        all.push(role(id));
    }
    const calls: { page: number; pageSize: number }[] = [];
    const page = (params: { page?: number; pageSize?: number }): Role[] => {
        calls.push({ page: params.page, pageSize: params.pageSize });
        const from = (params.page - 1) * params.pageSize;
        return all.slice(from, from + params.pageSize);
    };
    return {
        calls,
        ListRoleResponse(params) {
            return of(
                new HttpResponse<Array<Role>>({
                    headers: new HttpHeaders({
                        'x-total-count': `${total}`,
                    }),
                    body: page(params),
                })
            );
        },
        ListRole(params) {
            return of(page(params));
        },
    };
}

describe('role-util', () => {
    describe('getAllRoles', () => {
        it('returns the first page when there is only one', () => {
            const service = fakeRoleService(7);
            let roles: Role[];
            getAllRoles(service as unknown as RoleService).subscribe(
                res => (roles = res)
            );
            expect(roles.length).toBe(7);
            expect(service.calls).toEqual([{ page: 1, pageSize: 100 }]);
        });

        it('aggregates every page, in order, when the total exceeds one page', () => {
            const service = fakeRoleService(250);
            let roles: Role[];
            getAllRoles(service as unknown as RoleService).subscribe(
                res => (roles = res)
            );
            expect(roles.length).toBe(250);
            expect(roles.map(r => r.id)).toEqual(
                Array.from({ length: 250 }, (_, i) => i + 1)
            );
            expect(service.calls).toEqual([
                { page: 1, pageSize: 100 },
                { page: 2, pageSize: 100 },
                { page: 3, pageSize: 100 },
            ]);
        });

        it('treats a missing total count as a single page', () => {
            const service = {
                ListRoleResponse() {
                    return of(
                        new HttpResponse<Array<Role>>({
                            headers: new HttpHeaders(),
                            body: [role(1)],
                        })
                    );
                },
                ListRole() {
                    fail('should not ask for a second page');
                    return of([]);
                },
            };
            let roles: Role[];
            getAllRoles(service as unknown as RoleService).subscribe(
                res => (roles = res)
            );
            expect(roles).toEqual([role(1)]);
        });
    });

    describe('roleDisplayName', () => {
        it('maps a built-in role to its i18n key', () => {
            expect(
                roleDisplayName({
                    id: 1,
                    name: 'projectAdmin',
                    is_builtin: true,
                })
            ).toBe('MEMBER.PROJECT_ADMIN');
            expect(
                roleDisplayName({ id: 4, name: 'maintainer', is_builtin: true })
            ).toBe('MEMBER.PROJECT_MAINTAINER');
        });

        it('returns a custom role name verbatim', () => {
            expect(
                roleDisplayName({
                    id: 9,
                    name: 'BUTTON.DELETE',
                    is_builtin: false,
                })
            ).toBe('BUTTON.DELETE');
        });
    });
});
