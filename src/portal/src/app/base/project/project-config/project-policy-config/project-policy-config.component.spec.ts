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
import { Component, ViewChild } from '@angular/core';

const mockSystemInfo: SystemInfo[] = [
    {
        with_trivy: true,
        with_notary: true,
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
        with_notary: false,
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
