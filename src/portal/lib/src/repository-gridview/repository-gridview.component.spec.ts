import { ComponentFixture, TestBed, async } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';

import { RouterTestingModule } from '@angular/router/testing';

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ImageNameInputComponent } from "../image-name-input/image-name-input.component";
import { RepositoryGridviewComponent } from './repository-gridview.component';
import { TagComponent } from '../tag/tag.component';
import { FilterComponent } from '../filter/filter.component';

import { ErrorHandler } from '../error-handler/error-handler';
import { Repository, RepositoryItem, SystemInfo } from '../service/interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { RepositoryService, RepositoryDefaultService } from '../service/repository.service';
import { TagService, TagDefaultService } from '../service/tag.service';
import { SystemInfoService, SystemInfoDefaultService } from '../service/system-info.service';
import { VULNERABILITY_DIRECTIVES } from '../vulnerability-scanning/index';
import { HBR_GRIDVIEW_DIRECTIVES } from '../gridview/index';
import { PUSH_IMAGE_BUTTON_DIRECTIVES } from '../push-image/index';
import { INLINE_ALERT_DIRECTIVES } from '../inline-alert/index';
import { LabelPieceComponent } from "../label-piece/label-piece.component";
import { OperationService } from "../operation/operation.service";
import {
    ProjectDefaultService,
    ProjectService,
    RequestQueryParams,
    RetagDefaultService,
    RetagService
} from "../service";
import { UserPermissionService, UserPermissionDefaultService } from "../service/permission.service";
import { USERSTATICPERMISSION } from "../service/permission-static";
import { of } from "rxjs";
import { delay } from 'rxjs/operators';
describe('RepositoryComponentGridview (inline template)', () => {

  let compRepo: RepositoryGridviewComponent;
  let fixtureRepo: ComponentFixture<RepositoryGridviewComponent>;
  let mockSystemInfo: SystemInfo = {
    "with_notary": true,
    "with_admiral": false,
    "admiral_endpoint": "NA",
    "auth_mode": "db_auth",
    "registry_url": "10.112.122.56",
    "project_creation_restriction": "everyone",
    "self_registration": true,
    "has_ca_root": false,
    "harbor_version": "v1.1.1-rc1-160-g565110d"
  };

  let mockRepoData: RepositoryItem[] = [
    {
      "id": 1,
      "name": "library/busybox",
      "project_id": 1,
      "description": "asdfsadf",
      "pull_count": 0,
      "star_count": 0,
      "tags_count": 1
    },
    {
      "id": 2,
      "name": "library/nginx",
      "project_id": 1,
      "description": "asdf",
      "pull_count": 0,
      "star_count": 0,
      "tags_count": 1
    }
  ];
  let mockRepoNginxData: RepositoryItem[] = [
    {
      "id": 2,
      "name": "library/nginx",
      "project_id": 1,
      "description": "asdf",
      "pull_count": 0,
      "star_count": 0,
      "tags_count": 1
    }
  ];

  let mockRepo: Repository = {
    metadata: { xTotalCount: 2 },
    data: mockRepoData
  };
  let mockNginxRepo: Repository = {
    metadata: { xTotalCount: 2 },
    data: mockRepoNginxData
  };
  let config: IServiceConfig = {
    repositoryBaseEndpoint: '/api/repository/testing',
    systemInfoEndpoint: '/api/systeminfo/testing',
    targetBaseEndpoint: '/api/tag/testing'
  };
   const fakedErrorHandler = {
    error() {
      return undefined;
    }
  };
  const fakedSystemInfoService = {
    getSystemInfo() {
      return of(mockSystemInfo);
    }
  };
  const fakedRepositoryService = {
    getRepositories(projectId: number, name: string, param?: RequestQueryParams) {
      if (name === 'nginx') {
          return of(mockNginxRepo);
        }
        return of(mockRepo).pipe(delay(0));
    }
  };
  const fakedUserPermissionService = {
    getPermission() {
      return of(true);
    }
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        RouterTestingModule
      ],
      declarations: [
        RepositoryGridviewComponent,
        TagComponent,
        LabelPieceComponent,
        ConfirmationDialogComponent,
        ImageNameInputComponent,
        FilterComponent,
        VULNERABILITY_DIRECTIVES,
        PUSH_IMAGE_BUTTON_DIRECTIVES,
        INLINE_ALERT_DIRECTIVES,
        HBR_GRIDVIEW_DIRECTIVES,
      ],
      providers: [
         { provide: ErrorHandler, useValue: fakedErrorHandler },
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: RepositoryService, useValue: fakedRepositoryService },
        { provide: TagService, useClass: TagDefaultService },
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: RetagService, useClass: RetagDefaultService },
        { provide: SystemInfoService, useValue: fakedSystemInfoService },
        { provide: UserPermissionService, useValue: fakedUserPermissionService },
        { provide: OperationService }
      ]
    });
  }));

  beforeEach(async () => {
    fixtureRepo = TestBed.createComponent(RepositoryGridviewComponent);
    compRepo = fixtureRepo.componentInstance;
    compRepo.projectId = 1;
    compRepo.mode = '';
    compRepo.hasProjectAdminRole = true;
    compRepo.isCardView = false;
    fixtureRepo.detectChanges();
  });
  let originalTimeout;

  beforeEach(function () {
    originalTimeout = jasmine.DEFAULT_TIMEOUT_INTERVAL;
    jasmine.DEFAULT_TIMEOUT_INTERVAL = 100000;
  });

  afterEach(function () {
    jasmine.DEFAULT_TIMEOUT_INTERVAL = originalTimeout;
  });

  it('should create', async(() => {
    expect(compRepo).toBeTruthy();
  }));

  it('should load and render data', async(() => {
    fixtureRepo.whenStable().then(() => {
      fixtureRepo.detectChanges();
      let deRepo: DebugElement = fixtureRepo.debugElement.query(del => del.classes['datagrid-cell']);
      expect(deRepo).toBeTruthy();
      let elRepo: HTMLElement = deRepo.nativeElement;
      expect(elRepo).toBeTruthy();
      expect(elRepo.textContent).toEqual('library/busybox');
    });
  }));
  // Will fail after upgrade to angular 6. todo: need to fix it.
  it('should filter data by keyword', async () => {
       fixtureRepo.detectChanges();
       await fixtureRepo.whenStable();
       compRepo.doSearchRepoNames('nginx');
       fixtureRepo.detectChanges();
       await fixtureRepo.whenStable();
       let de: DebugElement[] = fixtureRepo.debugElement.queryAll(By.css('.datagrid-cell'));
       expect(de).toBeTruthy();
       expect(compRepo.repositories.length).toEqual(1);
       expect(de.length).toEqual(4);
       let el: HTMLElement = de[1].nativeElement;
       expect(el).toBeTruthy();
       expect(el.textContent).toEqual('library/nginx');
  });
});
