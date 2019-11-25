import { ComponentFixture, TestBed, async, } from '@angular/core/testing';
import { DebugElement} from '@angular/core';
import { RouterTestingModule } from '@angular/router/testing';
import { SharedModule } from '../../utils/shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog';
import { ImageNameInputComponent } from "../image-name-input/image-name-input.component";
import { RepositoryComponent } from './repository.component';
import { GridViewComponent } from '../gridview/grid-view.component';
import { FilterComponent } from '../filter/filter.component';
import { TagComponent } from '../tag/tag.component';
import { ErrorHandler } from '../../utils/error-handler';
import { Repository, RepositoryItem, Tag, SystemInfo, Label } from '../../services';
import { SERVICE_CONFIG, IServiceConfig } from '../../entities/service.config';
import { RepositoryService, RepositoryDefaultService } from '../../services';
import { SystemInfoService, SystemInfoDefaultService } from '../../services';
import { TagService, TagDefaultService } from '../../services';
import { LabelPieceComponent } from "../label-piece/label-piece.component";
import { LabelDefaultService, LabelService } from "../../services";
import { OperationService } from "../operation/operation.service";
import {
  ProjectDefaultService,
  ProjectService,
  RetagDefaultService,
  RetagService, ScanningResultDefaultService,
  ScanningResultService
} from "../../services";
import { UserPermissionDefaultService, UserPermissionService } from "../../services";
import { USERSTATICPERMISSION } from "../../services";
import { of } from "rxjs";
import { delay } from 'rxjs/operators';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { ChannelService } from "../../services/channel.service";
import { HarborLibraryModule } from "../../harbor-library.module";


class RouterStub {
  navigateByUrl(url: string) { return url; }
}

describe('RepositoryComponent (inline template)', () => {

  let compRepo: RepositoryComponent;
  let fixture: ComponentFixture<RepositoryComponent>;
  let repositoryService: RepositoryService;
  let systemInfoService: SystemInfoService;
  let userPermissionService: UserPermissionService;
  let tagService: TagService;
  let labelService: LabelService;

  let spyRepos: jasmine.Spy;
  let spyTags: jasmine.Spy;
  let spySystemInfo: jasmine.Spy;
  let spyLabels: jasmine.Spy;
  let spyLabels1: jasmine.Spy;

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
    metadata: {xTotalCount: 2},
    data: mockRepoData
  };

  let mockTagData: Tag[] = [
    {
      'digest': 'sha256:e5c82328a509aeb7c18c1d7fb36633dc638fcf433f651bdcda59c1cc04d3ee55',
      'name': '1.11.5',
      'size': '2049',
      'architecture': 'amd64',
      'os': 'linux',
      'os.version': '',
      'docker_version': '1.12.3',
      'author': 'NGINX Docker Maintainers \"docker-maint@nginx.com\"',
      'created': new Date('2016-11-08T22:41:15.912313785Z'),
      'signature': null,
      'labels': []
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
    {resource: USERSTATICPERMISSION.REPOSITORY_TAG_LABEL.KEY, action:  USERSTATICPERMISSION.REPOSITORY_TAG_LABEL.VALUE.CREATE},
    {resource: USERSTATICPERMISSION.REPOSITORY.KEY, action:  USERSTATICPERMISSION.REPOSITORY.VALUE.PULL},
    {resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY, action:  USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.DELETE},
    {resource: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.KEY, action:  USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.VALUE.CREATE},
  ];
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        RouterTestingModule,
        HarborLibraryModule,
        BrowserAnimationsModule
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: RepositoryService, useClass: RepositoryDefaultService },
        { provide: SystemInfoService, useClass: SystemInfoDefaultService },
        { provide: TagService, useClass: TagDefaultService },
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: RetagService, useClass: RetagDefaultService },
        { provide: LabelService, useClass: LabelDefaultService},
        { provide: UserPermissionService, useClass: UserPermissionDefaultService},
        { provide: ChannelService},
        { provide: OperationService },
        { provide: ScanningResultService, useValue: fakedScanningResultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RepositoryComponent);

    compRepo = fixture.componentInstance;

    compRepo.projectId = 1;
    compRepo.hasProjectAdminRole = true;
    compRepo.repoName = 'library/nginx';
    repositoryService = fixture.debugElement.injector.get(RepositoryService);
    systemInfoService = fixture.debugElement.injector.get(SystemInfoService);
    tagService = fixture.debugElement.injector.get(TagService);
    userPermissionService = fixture.debugElement.injector.get(UserPermissionService);
    labelService = fixture.debugElement.injector.get(LabelService);

    spyRepos = spyOn(repositoryService, 'getRepositories').and.returnValues(of(mockRepo).pipe(delay(0)));
    spySystemInfo = spyOn(systemInfoService, 'getSystemInfo').and.returnValues(of(mockSystemInfo).pipe(delay(0)));
    spyTags = spyOn(tagService, 'getTags').and.returnValues(of(mockTagData).pipe(delay(0)));

    spyLabels = spyOn(labelService, 'getGLabels').and.returnValues(of(mockLabels).pipe(delay(0)));
    spyLabels1 = spyOn(labelService, 'getPLabels').and.returnValues(of(mockLabels1).pipe(delay(0)));
    spyOn(userPermissionService, "hasProjectPermissions")
    .withArgs(compRepo.projectId, permissions )
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
      fixture.detectChanges();
      expect(de).toBeTruthy();
      let el: HTMLElement = de.nativeElement;
      expect(el).toBeTruthy();
      expect(el.textContent).toEqual('1.11.5');
    });
  }));
});
