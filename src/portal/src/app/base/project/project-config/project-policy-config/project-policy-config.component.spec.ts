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
import { SystemInfoService } from '../../../../shared/services';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ConfirmationDialogComponent } from '../../../../shared/components/confirmation-dialog';
import { ProjectPolicyConfigComponent } from './project-policy-config.component';
import { ProjectService } from '../../../../shared/services';
import { SystemCVEAllowlist, SystemInfo } from '../../../../shared/services';
import { Project } from './project';
import { UserPermissionService } from '../../../../shared/services';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { SessionService } from '../../../../shared/services/session.service';
import { Component, ViewChild } from '@angular/core';

const mockSystemInfo: SystemInfo[] = [
    {
        with_trivy: true,
        with_admiral: false,
        admiral_endpoint: 'NA',
        auth_mode: 'db_auth',
        registry_url: '10.112.122.56',
        project_creation_restriction: 'everyone',
        self_registration: true,
        has_ca_root: false,
        harbor_version: 'v1.1.1-rc1-160-g565110d',
    },
    {
        with_trivy: false,
        with_admiral: false,
        admiral_endpoint: 'NA',
        auth_mode: 'db_auth',
        registry_url: '10.112.122.56',
        project_creation_restriction: 'everyone',
        self_registration: true,
        has_ca_root: false,
        harbor_version: 'v1.1.1-rc1-160-g565110d',
    },
];
const mockProjectPolicies: Project[] | any[] = [
    {
        project_id: 1,
        owner_id: 1,
        name: 'library',
        creation_time: '2017-11-03T02:37:24Z',
        update_time: '2017-11-03T02:37:24Z',
        deleted: 0,
        owner_name: '',
        togglable: false,
        current_user_role_id: 0,
        repo_count: 0,
        metadata: {
            public: 'true',
        },
    },
    {
        project_id: 2,
        owner_id: 1,
        name: 'test',
        creation_time: '2017-11-03T02:37:24Z',
        update_time: '2017-11-03T02:37:24Z',
        deleted: 0,
        owner_name: '',
        togglable: false,
        current_user_role_id: 0,
        repo_count: 0,
        metadata: {
            auto_scan: 'true',
            auto_sbom_generation: 'true',
            enable_content_trust: 'true',
            prevent_vul: 'true',
            public: 'true',
            severity: 'low',
        },
    },
];
const mockSystemAllowlist: SystemCVEAllowlist = {
    expires_at: 1561996800,
    id: 1,
    items: [],
    project_id: 0,
};
const projectService = {
    getProject() {
        return of(mockProjectPolicies[1]);
    },
};

const systemInfoService = {
    getSystemInfo() {
        return of(mockSystemInfo[0]);
    },
    getSystemAllowlist() {
        return of(mockSystemAllowlist);
    },
};

const userPermissionService = {
    getPermission() {
        return of(true);
    },
};

const sessionService = {
    getCurrentUser() {
        return of({
            has_admin_role: true,
        });
    },
};
describe('ProjectPolicyConfigComponent', () => {
    let fixture: ComponentFixture<TestHostComponent>,
        component: TestHostComponent;
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [
                TestHostComponent,
                ProjectPolicyConfigComponent,
                ConfirmationDialogComponent,
            ],
            providers: [
                { provide: ErrorHandler, useClass: MessageHandlerService },
                { provide: ProjectService, useValue: projectService },
                { provide: SystemInfoService, useValue: systemInfoService },
                { provide: SessionService, useValue: sessionService },
                {
                    provide: UserPermissionService,
                    useValue: userPermissionService,
                },
            ],
        }).compileComponents();
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(TestHostComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });
    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get systemInfo', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.projectPolicyConfigComponent.systemInfo).toBeTruthy();
    });
    it('should get projectPolicy', () => {
        expect(
            component.projectPolicyConfigComponent.projectPolicy
        ).toBeTruthy();
        expect(
            component.projectPolicyConfigComponent.projectPolicy.ScanImgOnPush
        ).toBeTruthy();
        expect(
            component.projectPolicyConfigComponent.projectPolicy
                .GenerateSbomOnPush
        ).toBeTruthy();
    });
    it('should not allow empty and whitespace CVEs', async () => {
        // set cveIds with mix of empty and whitespace
        component.projectPolicyConfigComponent.cveIds = `
      ,   , \n ,  \t, ,
    `;

        component.projectPolicyConfigComponent.addToProjectAllowlist();

        const finalIds =
            component.projectPolicyConfigComponent.projectAllowlist.items.map(
                i => i.cve_id
            );
        expect(finalIds).not.toContain(' ');
        expect(finalIds).not.toContain('\n');
        expect(finalIds).not.toContain(''); // no empty CVEs

        // modal should be closed
        expect(component.projectPolicyConfigComponent.cveIds).toBeNull();
        expect(component.projectPolicyConfigComponent.showAddModal).toBeFalse();
    });
    it('should add only unique CVEs to the allowlist', () => {
        // set cveIds with duplicates and valid
        component.projectPolicyConfigComponent.cveIds = `
      CVE-2024-0002,
      CVE-2024-0002,
      CVE-2024-0004
    `;
        component.projectPolicyConfigComponent.addToProjectAllowlist();
        const finalIds =
            component.projectPolicyConfigComponent.projectAllowlist.items.map(
                i => i.cve_id
            );
        expect(finalIds).toContain('CVE-2024-0004');
        expect(finalIds).not.toContain(''); // no empty CVEs
        expect(finalIds.filter(id => id === 'CVE-2024-0002').length).toBe(1); // no duplicates

        // modal should be closed
        expect(component.projectPolicyConfigComponent.cveIds).toBeNull();
        expect(component.projectPolicyConfigComponent.showAddModal).toBeFalse();
    });
    it('should get hasChangeConfigRole', () => {
        expect(
            component.projectPolicyConfigComponent.hasChangeConfigRole
        ).toBeTruthy();
    });
});

// mock a TestHostComponent for ProjectPolicyConfigComponent
@Component({
    template: ` <hbr-project-policy-config
        [projectName]="'testing'"
        [projectId]="1">
    </hbr-project-policy-config>`,
})
class TestHostComponent {
    @ViewChild(ProjectPolicyConfigComponent)
    projectPolicyConfigComponent: ProjectPolicyConfigComponent;
}
