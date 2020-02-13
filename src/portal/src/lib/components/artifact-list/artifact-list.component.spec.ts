import { ComponentFixture, TestBed, async, } from '@angular/core/testing';
import { DebugElement, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { RouterTestingModule } from '@angular/router/testing';
import { SharedModule } from '../../utils/shared/shared.module';
import { ArtifactListComponent } from './artifact-list.component';
import { ErrorHandler } from '../../utils/error-handler';
import { Repository, RepositoryItem, Tag, SystemInfo, Label, ArtifactService, ArtifactDefaultService } from '../../services';
import { SERVICE_CONFIG, IServiceConfig } from '../../entities/service.config';
import { RepositoryService, RepositoryDefaultService } from '../../services';
import { SystemInfoService, SystemInfoDefaultService } from '../../services';
import { LabelDefaultService, LabelService } from "../../services";
import { OperationService } from "../operation/operation.service";
import {
  ProjectService,
  RetagDefaultService,
  RetagService,
  ScanningResultService
} from "../../services";
import { UserPermissionDefaultService, UserPermissionService } from "../../services";
import { USERSTATICPERMISSION } from "../../services";
import { of } from "rxjs";
import { delay } from 'rxjs/operators';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { ChannelService } from "../../services/channel.service";
import { HarborLibraryModule } from "../../harbor-library.module";
import { Artifact, Reference } from '../artifact/artifact';
import { ClarityModule } from '@clr/angular';


describe('ArtifactListComponent (inline template)', () => {

  let compRepo: ArtifactListComponent;
  let fixture: ComponentFixture<ArtifactListComponent>;
  let repositoryService: RepositoryService;
  let systemInfoService: SystemInfoService;
  let userPermissionService: UserPermissionService;
  let artifactService: ArtifactService;
  let labelService: LabelService;

  let spyRepos: jasmine.Spy;
  let spyTags: jasmine.Spy;
  let spySystemInfo: jasmine.Spy;
  let spyLabels: jasmine.Spy;
  let spyLabels1: jasmine.Spy;
  let mockPojectService = {
    getProject: () => of({ name: "library" })
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

  let mockRepoData: RepositoryItem[] = [
    {
      'id': 1,
      'name': 'library/busybox',
      'project_id': 1,
      'description': 'asdfsadf',
      'pull_count': 0,
      'star_count': 0,
      'tags_count': 1
    },
    {
      'id': 2,
      'name': 'library/nginx',
      'project_id': 1,
      'description': 'asdf',
      'pull_count': 0,
      'star_count': 0,
      'tags_count': 1
    }
  ];

  let mockRepo: Repository = {
    metadata: { xTotalCount: 2 },
    data: mockRepoData
  };

  let mockArtifactData: Artifact[] = [
    {
      "id": 1,
      type: 'image',
      repository: "goharbor/harbor-portal",
      tags: [{
        id: '1',
        artifact_id: 1,
        name: 'tag1',
        upload_time: '2020-01-06T09:40:08.036866579Z',
      },
      {
        id: '2',
        artifact_id: 2,
        name: 'tag2',
        upload_time: '2020-01-06T09:40:08.036866579Z',
      },],
      references: [new Reference(1), new Reference(2)],
      media_type: 'string',
      "digest": "sha256:4875cda368906fd670c9629b5e416ab3d6c0292015f3c3f12ef37dc9a32fc8d4",
      "size": 20372934,
      "scan_overview": {
        "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0": {
          "report_id": "5e64bc05-3102-11ea-93ae-0242ac140004",
          "scan_status": "Error",
          "severity": "",
          "duration": 118,
          "summary": null,
          "start_time": "2020-01-07T04:01:23.157711Z",
          "end_time": "2020-01-07T04:03:21.662766Z"
        }
      },
      "labels": [
        {
          "id": 3,
          "name": "aaa",
          "description": "",
          "color": "#0095D3",
          "scope": "g",
          "project_id": 0,
          "creation_time": "2020-01-13T05:44:00.580198Z",
          "update_time": "2020-01-13T05:44:00.580198Z",
          "deleted": false
        },
        {
          "id": 6,
          "name": "dbc",
          "description": "",
          "color": "",
          "scope": "g",
          "project_id": 0,
          "creation_time": "2020-01-13T08:27:19.279123Z",
          "update_time": "2020-01-13T08:27:19.279123Z",
          "deleted": false
        }
      ],
      "push_time": "2020-01-07T03:33:41.162319Z",
      "pull_time": "0001-01-01T00:00:00Z",
      hasReferenceArtifactList: [],
      noReferenceArtifactList: []

    }
  ];

  let mockLabels: Label[] = [{
    color: "#9b0d54",
    creation_time: "",
    description: "",
    id: 1,
    name: "label0-g",
    project_id: 1,
    scope: "p",
    update_time: "",
  },
  {
    color: "#9b0d54",
    creation_time: "",
    description: "",
    id: 2,
    name: "label1-g",
    project_id: 0,
    scope: "g",
    update_time: "",
  }];

  let mockLabels1: Label[] = [{
    color: "#9b0d54",
    creation_time: "",
    description: "",
    id: 1,
    name: "label0-g",
    project_id: 1,
    scope: "p",
    update_time: "",
  },
  {
    color: "#9b0d54",
    creation_time: "",
    description: "",
    id: 2,
    name: "label1-g",
    project_id: 1,
    scope: "p",
    update_time: "",
  }];

  let config: IServiceConfig = {
    repositoryBaseEndpoint: '/api/repository/testing',
    systemInfoEndpoint: '/api/systeminfo/testing',
    targetBaseEndpoint: '/api/tag/testing'
  };
  let mockHasAddLabelImagePermission: boolean = true;
  let mockHasRetagImagePermission: boolean = true;
  let mockHasDeleteImagePermission: boolean = true;
  let mockHasScanImagePermission: boolean = true;
  let fakedScanningResultService = {
    getProjectScanner() {
      return of({});
    }
  };
  const permissions = [
    { resource: USERSTATICPERMISSION.REPOSITORY_TAG_LABEL.KEY, action: USERSTATICPERMISSION.REPOSITORY_TAG_LABEL.VALUE.CREATE },
    { resource: USERSTATICPERMISSION.REPOSITORY.KEY, action: USERSTATICPERMISSION.REPOSITORY.VALUE.PULL },
    { resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY, action: USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.DELETE },
    { resource: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.KEY, action: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.VALUE.CREATE },
  ];
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        RouterTestingModule,
        HarborLibraryModule,
        BrowserAnimationsModule,
        ClarityModule
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: RepositoryService, useClass: RepositoryDefaultService },
        { provide: ChannelService, useValue: mockChannelService },
        { provide: SystemInfoService, useClass: SystemInfoDefaultService },
        { provide: ArtifactService, useClass: ArtifactDefaultService },
        { provide: ProjectService, useValue: mockPojectService },
        { provide: RetagService, useClass: RetagDefaultService },
        { provide: LabelService, useClass: LabelDefaultService },
        { provide: UserPermissionService, useClass: UserPermissionDefaultService },
        { provide: OperationService },
        { provide: ScanningResultService, useValue: fakedScanningResultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactListComponent);

    compRepo = fixture.componentInstance;

    compRepo.projectId = 1;
    compRepo.hasProjectAdminRole = true;
    compRepo.repoName = 'library/nginx';
    repositoryService = fixture.debugElement.injector.get(RepositoryService);
    systemInfoService = fixture.debugElement.injector.get(SystemInfoService);
    artifactService = fixture.debugElement.injector.get(ArtifactService);
    userPermissionService = fixture.debugElement.injector.get(UserPermissionService);
    labelService = fixture.debugElement.injector.get(LabelService);

    spyRepos = spyOn(repositoryService, 'getRepositories').and.returnValues(of(mockRepo).pipe(delay(0)));
    spySystemInfo = spyOn(systemInfoService, 'getSystemInfo').and.returnValues(of(mockSystemInfo).pipe(delay(0)));
    spyTags = spyOn(artifactService, 'TriggerArtifactChan$').and.returnValues(of('repoName').pipe(delay(0)));
    spyTags = spyOn(artifactService, 'getArtifactList').and.returnValues(of(mockArtifactData).pipe(delay(0)));
    spyLabels = spyOn(labelService, 'getGLabels').and.returnValues(of(mockLabels).pipe(delay(0)));
    spyLabels1 = spyOn(labelService, 'getPLabels').and.returnValues(of(mockLabels1).pipe(delay(0)));
    spyOn(userPermissionService, "hasProjectPermissions")
      .withArgs(compRepo.projectId, permissions)
      .and.returnValue(of([mockHasAddLabelImagePermission, mockHasRetagImagePermission,
        mockHasDeleteImagePermission, mockHasScanImagePermission]));
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

  it('should load and render data', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(del => del.classes['datagrid-cell']);
      //  de = fixture.debugElement.query(By.css('datagrid-cell'));
      fixture.detectChanges();
      expect(de).toBeTruthy();
      let el: HTMLElement = de.nativeElement;
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('sha256:4875cda3');
    });
  }));
});
