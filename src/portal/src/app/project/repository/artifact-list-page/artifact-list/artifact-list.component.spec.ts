import { ComponentFixture, TestBed, async, } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ArtifactListComponent } from './artifact-list.component';
import { of } from "rxjs";
import { delay } from 'rxjs/operators';
import { ClarityModule } from '@clr/angular';
import { ActivatedRoute } from '@angular/router';
import {
  SystemInfo, SystemInfoDefaultService,
  SystemInfoService,
} from "../../../../../lib/services";
import { ArtifactDefaultService, ArtifactService } from "../../artifact/artifact.service";
import { ChannelService } from "../../../../../lib/services/channel.service";

import { TranslateFakeLoader, TranslateLoader, TranslateModule, TranslateService } from "@ngx-translate/core";
import { ErrorHandler } from "../../../../../lib/utils/error-handler";
import { IServiceConfig, SERVICE_CONFIG } from "../../../../../lib/entities/service.config";
import { SharedModule } from "../../../../../lib/utils/shared/shared.module";
import {
  RepositoryService as NewRepositoryService
} from "../../../../../../ng-swagger-gen/services/repository.service";

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
  const config: IServiceConfig = {
    repositoryBaseEndpoint: "/api/repositories/testing"
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        ClarityModule,
        SharedModule,
        TranslateModule.forRoot({
          loader: {
            provide: TranslateLoader,
            useClass: TranslateFakeLoader,
          }
        })
      ],
      schemas: [
        NO_ERRORS_SCHEMA
      ],
      declarations: [
        ArtifactListComponent
      ],
      providers: [
        TranslateService,
        { provide: ErrorHandler, useValue: fakedErrorHandler },
        { provide: ChannelService, useValue: mockChannelService },
        { provide: SystemInfoService, useClass: SystemInfoDefaultService },
        { provide: ArtifactService, useClass: ArtifactDefaultService },
        { provide: ActivatedRoute, useValue: mockActivatedRoute },
        { provide: SERVICE_CONFIG, useValue: config },
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
