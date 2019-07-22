import { SystemInfoService, SystemInfoDefaultService } from './../service/system-info.service';
import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { ErrorHandler } from '../error-handler/error-handler';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ProjectPolicyConfigComponent } from './project-policy-config.component';
import { SharedModule } from '../shared/shared.module';
import { ProjectService, ProjectDefaultService} from '../service/project.service';
import { SERVICE_CONFIG, IServiceConfig} from '../service.config';
import {SystemCVEWhitelist, SystemInfo} from '../service/interface';
import { Project } from './project';
import { UserPermissionService, UserPermissionDefaultService } from '../service/permission.service';
import { USERSTATICPERMISSION } from '../service/permission-static';
import { of } from 'rxjs';
describe('ProjectPolicyConfigComponent', () => {

  let systemInfoService: SystemInfoService;
  let projectPolicyService: ProjectService;
  let userPermissionService: UserPermissionService;

  let spySystemInfo: jasmine.Spy;
  let spyProjectPolicies: jasmine.Spy;
  let mockHasChangeConfigRole: boolean = true;
  let mockSystemInfo: SystemInfo[] = [
    {
      'with_clair': true,
      'with_notary': true,
      'with_admiral': false,
      'admiral_endpoint': 'NA',
      'auth_mode': 'db_auth',
      'registry_url': '10.112.122.56',
      'project_creation_restriction': 'everyone',
      'self_registration': true,
      'has_ca_root': false,
      'harbor_version': 'v1.1.1-rc1-160-g565110d'
    },
    {
      'with_clair': false,
      'with_notary': false,
      'with_admiral': false,
      'admiral_endpoint': 'NA',
      'auth_mode': 'db_auth',
      'registry_url': '10.112.122.56',
      'project_creation_restriction': 'everyone',
      'self_registration': true,
      'has_ca_root': false,
      'harbor_version': 'v1.1.1-rc1-160-g565110d'
    }
  ];

  let mockProjectPolicies: Project[] | any[] = [
    {
      'project_id': 1,
      'owner_id': 1,
      'name': 'library',
      'creation_time': '2017-11-03T02:37:24Z',
      'update_time': '2017-11-03T02:37:24Z',
      'deleted': 0,
      'owner_name': '',
      'togglable': false,
      'current_user_role_id': 0,
      'repo_count': 0,
      'metadata': {
        'public': 'true'
      }
    },
    {
      'project_id': 2,
      'owner_id': 1,
      'name': 'test',
      'creation_time': '2017-11-03T02:37:24Z',
      'update_time': '2017-11-03T02:37:24Z',
      'deleted': 0,
      'owner_name': '',
      'togglable': false,
      'current_user_role_id': 0,
      'repo_count': 0,
      'metadata': {
        'auto_scan': 'true',
        'enable_content_trust': 'true',
        'prevent_vul': 'true',
        'public': 'true',
        'severity': 'low'
      }
    }
  ];
  let mockSystemWhitelist: SystemCVEWhitelist = {
    "expires_at": 1561996800,
    "id": 1,
    "items": [],
    "project_id": 0
  };
  let component: ProjectPolicyConfigComponent;
  let fixture: ComponentFixture<ProjectPolicyConfigComponent>;

  let config: IServiceConfig = {
    projectPolicyEndpoint: '/api/projects/testing',
    systemInfoEndpoint: '/api/systeminfo/testing',
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule],
      declarations: [
        ProjectPolicyConfigComponent,
        ConfirmationDialogComponent,
        ConfirmationDialogComponent,
       ],
       providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: SystemInfoService, useClass: SystemInfoDefaultService},
        { provide: UserPermissionService, useClass: UserPermissionDefaultService},
      ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ProjectPolicyConfigComponent);
    component = fixture.componentInstance;
    component.projectId = 1;
    component.hasProjectAdminRole = true;

    systemInfoService = fixture.debugElement.injector.get(SystemInfoService);
    projectPolicyService = fixture.debugElement.injector.get(ProjectService);

    spySystemInfo = spyOn(systemInfoService, 'getSystemInfo').and.returnValues(of(mockSystemInfo[0]));
    spySystemInfo = spyOn(systemInfoService, 'getSystemWhitelist').and.returnValue(of(mockSystemWhitelist));
    spyProjectPolicies = spyOn(projectPolicyService, 'getProject').and.returnValues(of(mockProjectPolicies[0]));

    userPermissionService = fixture.debugElement.injector.get(UserPermissionService);
    spyOn(userPermissionService, "getPermission")
    .withArgs(component.projectId,
      USERSTATICPERMISSION.CONFIGURATION.KEY, USERSTATICPERMISSION.CONFIGURATION.VALUE.UPDATE )
    .and.returnValue(of(mockHasChangeConfigRole));
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
