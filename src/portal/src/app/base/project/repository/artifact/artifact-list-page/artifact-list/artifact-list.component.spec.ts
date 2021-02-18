import { ComponentFixture, TestBed, waitForAsync, } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ArtifactListComponent } from './artifact-list.component';
import { of } from "rxjs";
import { delay } from 'rxjs/operators';
import { ActivatedRoute } from '@angular/router';
import {
  SystemInfo, SystemInfoDefaultService,
  SystemInfoService,
} from "../../../../../../shared/services";
import { ArtifactDefaultService, ArtifactService } from "../../artifact.service";
import { ChannelService } from "../../../../../../shared/services/channel.service";
import { ErrorHandler } from "../../../../../../shared/units/error-handler";
import {
  RepositoryService as NewRepositoryService
} from "../../../../../../../../ng-swagger-gen/services/repository.service";
import { SharedTestingModule } from "../../../../../../shared/shared.module";
import { HttpClientTestingModule } from "@angular/common/http/testing";

describe('ArtifactListComponent (inline template)', () => {

  let compRepo: ArtifactListComponent;
  let fixture: ComponentFixture<ArtifactListComponent>;
  let systemInfoService: SystemInfoService;
  let artifactService: ArtifactService;
  let spyRepos: jasmine.Spy;
  let spySystemInfo: jasmine.Spy;
  let mockActivatedRoute = {
    data: of(
      {
        projectResolver: {
          name: 'library'
        }
      }
    ),
    params: {
      subscribe: () => {
        return of(null);
      }
    },
    snapshot: { data: null }
  };
  let mockChannelService = {
    scanCommand$: of(1)
  };
  let mockSystemInfo: SystemInfo = {
    'with_notary': true,
    'with_admiral': false,
    'admiral_endpoint': 'NA',
    'auth_mode': 'db_auth',
    'registry_url': '10.112.122.56',
    'project_creation_restriction': 'everyone',
    'self_registration': true,
    'has_ca_root': false,
    'harbor_version': 'v1.1.1-rc1-160-g565110d'
  };

  let newRepositoryService = {
    updateRepository: () => of(null),
    getRepository: () => of({description: ''})
  };

  const fakedErrorHandler = {
    error: () => {}
  };
  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule,
        SharedTestingModule,
      ],
      schemas: [
        NO_ERRORS_SCHEMA
      ],
      declarations: [
        ArtifactListComponent
      ],
      providers: [
        { provide: ErrorHandler, useValue: fakedErrorHandler },
        { provide: ChannelService, useValue: mockChannelService },
        { provide: SystemInfoService, useClass: SystemInfoDefaultService },
        { provide: ArtifactService, useClass: ArtifactDefaultService },
        { provide: ActivatedRoute, useValue: mockActivatedRoute },
        { provide: NewRepositoryService, useValue: newRepositoryService},
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactListComponent);
    compRepo = fixture.componentInstance;
    compRepo.projectId = 1;
    compRepo.hasProjectAdminRole = true;
    compRepo.repoName = 'library/nginx';
    systemInfoService = fixture.debugElement.injector.get(SystemInfoService);
    artifactService = fixture.debugElement.injector.get(ArtifactService);
    spySystemInfo = spyOn(systemInfoService, 'getSystemInfo').and.returnValues(of(mockSystemInfo).pipe(delay(0)));
    fixture.detectChanges();
  });
  let originalTimeout;

  beforeEach(function () {
    originalTimeout = jasmine.DEFAULT_TIMEOUT_INTERVAL;
    jasmine.DEFAULT_TIMEOUT_INTERVAL = 100000;
  });

  afterEach(function () {
    jasmine.DEFAULT_TIMEOUT_INTERVAL = originalTimeout;
  });
  it('should create', () => {
    expect(compRepo).toBeTruthy();
  });
});
