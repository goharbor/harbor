import { ComponentFixture, TestBed, async } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';
import { RouterTestingModule } from '@angular/router/testing';
import { of } from "rxjs";
import { RepositoryService as NewRepositoryService } from "../../../../ng-swagger-gen/services/repository.service";
import { RepositoryGridviewComponent } from "./repository-gridview.component";
import {
  ProjectDefaultService,
  ProjectService,
  Repository,
  RepositoryItem,
  RequestQueryParams, RetagDefaultService, RetagService,
  SystemInfo, SystemInfoService,
  TagDefaultService,
  TagService, UserPermissionService
} from "../../../lib/services";
import { IServiceConfig, SERVICE_CONFIG } from "../../../lib/entities/service.config";
import { delay } from 'rxjs/operators';
import { SharedModule } from "../../../lib/utils/shared/shared.module";
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { RepositoryService } from "./repository.service";
import { OperationService } from "../../../lib/components/operation/operation.service";
import { ProjectModule } from "../project.module";
import { ActivatedRoute } from "@angular/router";
import { Repository as NewRepository } from "../../../../ng-swagger-gen/models/repository";
import { StrictHttpResponse as __StrictHttpResponse } from '../../../../ng-swagger-gen/strict-http-response';

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
  let mockRepoData: NewRepository[] = [
    {
      "id": 1,
      "name": "library/busybox",
      "project_id": 1,
      "description": "asdfsadf",
      "pull_count": 0,
      "artifact_count": 1
    },
    {
      "id": 2,
      "name": "library/nginx",
      "project_id": 1,
      "description": "asdf",
      "pull_count": 0,
      "artifact_count": 1
    }
  ];
  let mockRepoNginxData: NewRepository[] = [
    {
      "id": 2,
      "name": "library/nginx",
      "project_id": 1,
      "description": "asdf",
      "pull_count": 0,
      "artifact_count": 1
    }
  ];

  let mockRepo: NewRepository[] = mockRepoData;
  let mockNginxRepo: NewRepository[] = mockRepoNginxData;
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
    listRepositoriesResponse(params: NewRepositoryService.ListRepositoriesParams) {
      if (params.name === 'nginx') {
        return of({headers: new Map(), body: mockNginxRepo});
        }
      return of({headers: new Map(), body: mockRepo}).pipe(delay(0));
    }
  };
  const fakedUserPermissionService = {
    getPermission() {
      return of(true);
    }
  };
  const fakedActivatedRoute = {
    snapshot: {
      parent: {
        params: {
          id: "1"
        }
      }
    }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        RouterTestingModule,
        ProjectModule
      ],
      providers: [
        { provide: ActivatedRoute, useValue: fakedActivatedRoute },
        { provide: ErrorHandler, useValue: fakedErrorHandler },
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: RepositoryService, useValue: fakedRepositoryService },
        { provide: NewRepositoryService, useValue: fakedRepositoryService },
        { provide: TagService, useClass: TagDefaultService },
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: RetagService, useClass: RetagDefaultService },
        { provide: SystemInfoService, useValue: fakedSystemInfoService },
        { provide: UserPermissionService, useValue: fakedUserPermissionService },
        { provide: OperationService },
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
});
