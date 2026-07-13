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
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { MemberComponent } from './member.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { SessionService } from '../../../shared/services/session.service';
import { MemberService } from '../../../../../ng-swagger-gen/services/member.service';
import { AppConfigService } from '../../../services/app-config.service';
import { of } from 'rxjs';
import { OperationService } from '../../../shared/components/operation/operation.service';
import { UserPermissionService } from '../../../shared/services';
import { ErrorHandler } from '../../../shared/units/error-handler';
import { ConfirmationDialogService } from '../../global-confirmation-dialog/confirmation-dialog.service';
import { SharedTestingModule } from '../../../shared/shared.module';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../ng-swagger-gen/models/registry';
import { ProjectMemberEntity } from '../../../../../ng-swagger-gen/models/project-member-entity';
import { delay } from 'rxjs/operators';
import { Role } from '../../../../../ng-swagger-gen/models/role';
import { RoleService } from '../../../../../ng-swagger-gen/services/role.service';
import { SessionUser } from '../../../shared/entities/session-user';

describe('MemberComponent', () => {
    let component: MemberComponent;
    let fixture: ComponentFixture<MemberComponent>;
    const mockedMembers: ProjectMemberEntity[] = [
        {
            id: 1,
            project_id: 1,
            entity_name: 'test1',
            role_name: 'projectAdmin',
            role_id: 1,
            entity_id: 1,
            entity_type: 'u',
        },
        {
            id: 2,
            project_id: 1,
            entity_name: 'test2',
            role_name: 'projectAdmin',
            role_id: 1,
            entity_id: 2,
            entity_type: 'u',
        },
    ];
    const mockMemberService = {
        getUsersNameList: () => {
            return of([]);
        },
        listProjectMembersResponse: () => {
            const response: HttpResponse<Array<Registry>> = new HttpResponse<
                Array<Registry>
            >({
                headers: new HttpHeaders({ 'x-total-count': '2' }),
                body: mockedMembers,
            });
            return of(response).pipe(delay(0));
        },
        updateProjectMember: () => {
            return of(null);
        },
        deleteProjectMember: () => {
            return of(null);
        },
    };
    const mockSessionService = {
        getCurrentUser: () => {
            return of({
                user_id: 1,
            });
        },
    };
    const mockAppConfigService = {
        isLdapMode: () => {
            return false;
        },
        isHttpAuthMode: () => {
            return false;
        },
        isOidcMode: () => {
            return true;
        },
    };
    const mockOperationService = {
        publishInfo: () => {},
    };
    const mockMessageHandlerService = {
        handleError: () => {},
    };
    const mockConfirmationDialogService = {
        openComfirmDialog: () => {},
        confirmationConfirm$: of({
            state: 1,
            source: 2,
        }),
    };
    const mockUserPermissionService = {
        getPermission() {
            return of(true);
        },
    };
    const mockErrorHandler = {
        error() {},
    };
    const mockRoleService = {
        ListRole: () => of([]),
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [MemberComponent],
            providers: [
                { provide: MemberService, useValue: mockMemberService },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
                {
                    provide: ConfirmationDialogService,
                    useValue: mockConfirmationDialogService,
                },
                { provide: SessionService, useValue: mockSessionService },
                { provide: OperationService, useValue: mockOperationService },
                { provide: AppConfigService, useValue: mockAppConfigService },
                {
                    provide: UserPermissionService,
                    useValue: mockUserPermissionService,
                },
                { provide: ErrorHandler, useValue: mockErrorHandler },
                { provide: RoleService, useValue: mockRoleService },
                {
                    provide: ActivatedRoute,
                    useValue: {
                        RouterparamMap: of({ get: key => 'value' }),
                        snapshot: {
                            parent: {
                                parent: {
                                    params: { id: 1 },
                                },
                            },
                        },
                    },
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(MemberComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should render member list', async () => {
        fixture.autoDetectChanges(true);
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(2);
    });
});

// ---------------------------------------------------------------------------
// Fixtures shared across custom-role unit tests
// ---------------------------------------------------------------------------

const pullOnlyAccess = [{ resource: 'repository', action: 'pull' }];

function makeRole(
    id: number,
    name: string,
    isBuiltin: boolean,
    access: { resource: string; action: string }[] = []
): Role {
    return {
        id,
        name,
        is_builtin: isBuiltin,
        permissions: access.length
            ? [{ kind: 'project', namespace: '*', access }]
            : [],
    } as Role;
}

const BUILTIN_ROLES: Role[] = [
    makeRole(1, 'projectAdmin', true, [
        { resource: 'member', action: 'create' },
        { resource: 'member', action: 'delete' },
        { resource: 'repository', action: 'pull' },
        { resource: 'repository', action: 'push' },
    ]),
    makeRole(2, 'maintainer', true, [
        { resource: 'repository', action: 'pull' },
        { resource: 'repository', action: 'push' },
    ]),
    makeRole(3, 'developer', true, [
        { resource: 'repository', action: 'pull' },
        { resource: 'repository', action: 'push' },
    ]),
    makeRole(4, 'guest', true, [{ resource: 'repository', action: 'pull' }]),
    makeRole(5, 'limitedGuest', true, [
        { resource: 'repository', action: 'pull' },
    ]),
];

const CUSTOM_PULL_ONLY = makeRole(100, 'pull-only', false, pullOnlyAccess);
const CUSTOM_PUSH_PULL = makeRole(101, 'push-pull', false, [
    { resource: 'repository', action: 'pull' },
    { resource: 'repository', action: 'push' },
]);

// ---------------------------------------------------------------------------
// Helper: build a bare MemberComponent without TestBed (pure logic tests)
// ---------------------------------------------------------------------------

function buildComponent(currentUser: Partial<SessionUser>): MemberComponent {
    // Cast to any so we can construct without all injected services.
    const c = Object.create(MemberComponent.prototype) as MemberComponent;
    (c as any).currentUser = currentUser as SessionUser;
    (c as any).roles = [];
    (c as any).currentUserRoleId = null;
    (c as any).assignableRoleIds = null;
    return c;
}

// ---------------------------------------------------------------------------
// computeAssignableRoles
// ---------------------------------------------------------------------------

describe('MemberComponent — computeAssignableRoles', () => {
    it('returns null (unrestricted) when has_admin_role is true', () => {
        const c = buildComponent({ has_admin_role: true });
        (c as any).roles = BUILTIN_ROLES;
        (c as any).currentUserRoleId = 2; // maintainer
        (c as any).computeAssignableRoles();
        expect((c as any).assignableRoleIds).toBeNull();
    });

    it('returns null when caller is the built-in projectAdmin', () => {
        const c = buildComponent({ has_admin_role: false });
        (c as any).roles = BUILTIN_ROLES;
        (c as any).currentUserRoleId = 1; // projectAdmin
        (c as any).computeAssignableRoles();
        expect((c as any).assignableRoleIds).toBeNull();
    });

    it('returns null when caller role is not found in list', () => {
        const c = buildComponent({ has_admin_role: false });
        (c as any).roles = BUILTIN_ROLES;
        (c as any).currentUserRoleId = 999; // unknown
        (c as any).computeAssignableRoles();
        expect((c as any).assignableRoleIds).toBeNull();
    });

    it('limits to roles whose perms are a subset of caller perms (custom pull-only)', () => {
        const allRoles = [...BUILTIN_ROLES, CUSTOM_PULL_ONLY, CUSTOM_PUSH_PULL];
        const c = buildComponent({ has_admin_role: false });
        (c as any).roles = allRoles;
        (c as any).currentUserRoleId = CUSTOM_PULL_ONLY.id;
        (c as any).computeAssignableRoles();

        const assignable: Set<number> = (c as any).assignableRoleIds;
        expect(assignable).not.toBeNull();
        // pull-only role is assignable (same permissions)
        expect(assignable!.has(CUSTOM_PULL_ONLY.id)).toBeTrue();
        // push-pull requires push which caller lacks → not assignable
        expect(assignable!.has(CUSTOM_PUSH_PULL.id)).toBeFalse();
        // guest (pull only) → assignable
        expect(assignable!.has(4)).toBeTrue();
        // maintainer (push+pull) → not assignable (caller lacks push)
        expect(assignable!.has(2)).toBeFalse();
        // projectAdmin (member:create etc.) → not assignable
        expect(assignable!.has(1)).toBeFalse();
    });

    it('includes roles with zero permissions (always a subset)', () => {
        const emptyRole = makeRole(200, 'empty', false, []);
        const c = buildComponent({ has_admin_role: false });
        (c as any).roles = [...BUILTIN_ROLES, CUSTOM_PULL_ONLY, emptyRole];
        (c as any).currentUserRoleId = CUSTOM_PULL_ONLY.id;
        (c as any).computeAssignableRoles();

        const assignable: Set<number> = (c as any).assignableRoleIds;
        expect(assignable!.has(emptyRole.id)).toBeTrue();
    });
});

// ---------------------------------------------------------------------------
// isRoleAssignable
// ---------------------------------------------------------------------------

describe('MemberComponent — isRoleAssignable', () => {
    it('returns true for any role when assignableRoleIds is null', () => {
        const c = buildComponent({});
        (c as any).assignableRoleIds = null;
        expect(c.isRoleAssignable(BUILTIN_ROLES[0])).toBeTrue();
        expect(c.isRoleAssignable(CUSTOM_PULL_ONLY)).toBeTrue();
    });

    it('returns true only for roles in the assignable set', () => {
        const c = buildComponent({});
        (c as any).assignableRoleIds = new Set<number>([
            4,
            5,
            CUSTOM_PULL_ONLY.id,
        ]);
        expect(c.isRoleAssignable(makeRole(4, 'guest', true))).toBeTrue();
        expect(c.isRoleAssignable(CUSTOM_PULL_ONLY)).toBeTrue();
        expect(c.isRoleAssignable(BUILTIN_ROLES[0])).toBeFalse(); // projectAdmin id=1
        expect(c.isRoleAssignable(BUILTIN_ROLES[1])).toBeFalse(); // maintainer id=2
    });
});

// ---------------------------------------------------------------------------
// getRoleDisplayName
// ---------------------------------------------------------------------------

describe('MemberComponent — getRoleDisplayName', () => {
    let c: MemberComponent;
    beforeEach(() => {
        c = buildComponent({});
    });

    it('returns translation key for built-in projectAdmin', () => {
        expect(c.getRoleDisplayName(makeRole(1, 'projectAdmin', true))).toBe(
            'MEMBER.PROJECT_ADMIN'
        );
    });

    it('returns translation key for built-in maintainer', () => {
        expect(c.getRoleDisplayName(makeRole(4, 'maintainer', true))).toBe(
            'MEMBER.PROJECT_MAINTAINER'
        );
    });

    it('returns translation key for built-in developer', () => {
        expect(c.getRoleDisplayName(makeRole(2, 'developer', true))).toBe(
            'MEMBER.DEVELOPER'
        );
    });

    it('returns translation key for built-in guest', () => {
        expect(c.getRoleDisplayName(makeRole(3, 'guest', true))).toBe(
            'MEMBER.GUEST'
        );
    });

    it('returns translation key for built-in limitedGuest', () => {
        expect(c.getRoleDisplayName(makeRole(5, 'limitedGuest', true))).toBe(
            'MEMBER.LIMITED_GUEST'
        );
    });

    it('returns raw name for unknown built-in name', () => {
        expect(c.getRoleDisplayName(makeRole(99, 'unknownBuiltin', true))).toBe(
            'unknownBuiltin'
        );
    });

    it('returns raw name for custom role regardless of name', () => {
        expect(c.getRoleDisplayName(CUSTOM_PULL_ONLY)).toBe('pull-only');
        expect(c.getRoleDisplayName(makeRole(200, 'projectAdmin', false))).toBe(
            'projectAdmin'
        );
    });
});

// ---------------------------------------------------------------------------
// getMemberRoleDisplayName
// ---------------------------------------------------------------------------

describe('MemberComponent — getMemberRoleDisplayName', () => {
    let c: MemberComponent;
    beforeEach(() => {
        c = buildComponent({});
        (c as any).roles = [...BUILTIN_ROLES, CUSTOM_PULL_ONLY];
    });

    function member(roleId: number, roleName: string): ProjectMemberEntity {
        return {
            id: 1,
            project_id: 1,
            entity_name: 'u',
            role_id: roleId,
            role_name: roleName,
            entity_id: 1,
            entity_type: 'u',
        };
    }

    it('returns translation key for built-in role found in roles list', () => {
        expect(c.getMemberRoleDisplayName(member(2, 'maintainer'))).toBe(
            'MEMBER.PROJECT_MAINTAINER'
        );
    });

    it('returns raw name for custom role found in roles list', () => {
        expect(
            c.getMemberRoleDisplayName(member(CUSTOM_PULL_ONLY.id, 'pull-only'))
        ).toBe('pull-only');
    });

    it('falls back to RoleInfo translation key when roles list is empty (race condition)', () => {
        (c as any).roles = [];
        // maintainer role_id = 4 → RoleInfo[4] = 'MEMBER.PROJECT_MAINTAINER'
        const result = c.getMemberRoleDisplayName(member(4, 'maintainer'));
        expect(result).toBe('MEMBER.PROJECT_MAINTAINER');
    });

    it('falls back to raw role_name when id is not in RoleInfo', () => {
        (c as any).roles = [];
        expect(c.getMemberRoleDisplayName(member(999, 'my-custom-role'))).toBe(
            'my-custom-role'
        );
    });
});

// ---------------------------------------------------------------------------
// Anti-escalation: valid (non-escalating) assignment cases
// These tests confirm that roles within the caller's permission set ARE
// assignable — i.e. the guard does not over-restrict.
// ---------------------------------------------------------------------------

describe('MemberComponent — anti-escalation: valid assignments allowed', () => {
    it('maintainer can assign developer (developer perms ⊆ maintainer perms)', () => {
        // maintainer has push+pull; developer has push+pull → equal subset → OK
        const maintainer = BUILTIN_ROLES.find(r => r.name === 'maintainer')!;
        const developer = BUILTIN_ROLES.find(r => r.name === 'developer')!;
        const c = buildComponent({ has_admin_role: false });
        (c as any).roles = BUILTIN_ROLES;
        (c as any).currentUserRoleId = maintainer.id;
        (c as any).computeAssignableRoles();

        expect(c.isRoleAssignable(developer)).toBeTrue();
    });

    it('maintainer can assign guest (guest perms ⊂ maintainer perms)', () => {
        const maintainer = BUILTIN_ROLES.find(r => r.name === 'maintainer')!;
        const guest = BUILTIN_ROLES.find(r => r.name === 'guest')!;
        const c = buildComponent({ has_admin_role: false });
        (c as any).roles = BUILTIN_ROLES;
        (c as any).currentUserRoleId = maintainer.id;
        (c as any).computeAssignableRoles();

        expect(c.isRoleAssignable(guest)).toBeTrue();
    });

    it('maintainer cannot assign projectAdmin (projectAdmin perms ⊄ maintainer perms)', () => {
        const maintainer = BUILTIN_ROLES.find(r => r.name === 'maintainer')!;
        const projectAdmin = BUILTIN_ROLES.find(
            r => r.name === 'projectAdmin'
        )!;
        const c = buildComponent({ has_admin_role: false });
        (c as any).roles = BUILTIN_ROLES;
        (c as any).currentUserRoleId = maintainer.id;
        (c as any).computeAssignableRoles();

        expect(c.isRoleAssignable(projectAdmin)).toBeFalse();
    });

    it('push-pull custom role can assign pull-only custom role', () => {
        const allRoles = [...BUILTIN_ROLES, CUSTOM_PULL_ONLY, CUSTOM_PUSH_PULL];
        const c = buildComponent({ has_admin_role: false });
        (c as any).roles = allRoles;
        (c as any).currentUserRoleId = CUSTOM_PUSH_PULL.id;
        (c as any).computeAssignableRoles();

        expect(c.isRoleAssignable(CUSTOM_PULL_ONLY)).toBeTrue();
    });

    it('pull-only custom role cannot assign push-pull custom role', () => {
        const allRoles = [...BUILTIN_ROLES, CUSTOM_PULL_ONLY, CUSTOM_PUSH_PULL];
        const c = buildComponent({ has_admin_role: false });
        (c as any).roles = allRoles;
        (c as any).currentUserRoleId = CUSTOM_PULL_ONLY.id;
        (c as any).computeAssignableRoles();

        expect(c.isRoleAssignable(CUSTOM_PUSH_PULL)).toBeFalse();
    });

    it('sysadmin (has_admin_role) can assign any role including projectAdmin', () => {
        const c = buildComponent({ has_admin_role: true });
        (c as any).roles = [...BUILTIN_ROLES, CUSTOM_PUSH_PULL];
        (c as any).currentUserRoleId = 3; // developer
        (c as any).computeAssignableRoles();

        // assignableRoleIds is null → isRoleAssignable always true
        expect((c as any).assignableRoleIds).toBeNull();
        expect(c.isRoleAssignable(BUILTIN_ROLES[0])).toBeTrue(); // projectAdmin
        expect(c.isRoleAssignable(CUSTOM_PUSH_PULL)).toBeTrue();
    });
});
