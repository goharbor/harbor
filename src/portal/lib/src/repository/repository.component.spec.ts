import { ComponentFixture, TestBed, async, } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement} from '@angular/core';
import { RouterTestingModule } from '@angular/router/testing';

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ImageNameInputComponent } from "../image-name-input/image-name-input.component";
import { RepositoryComponent } from './repository.component';
import { GridViewComponent } from '../gridview/grid-view.component';
import { FilterComponent } from '../filter/filter.component';
import { TagComponent } from '../tag/tag.component';
import { VULNERABILITY_DIRECTIVES } from '../vulnerability-scanning/index';
import { PUSH_IMAGE_BUTTON_DIRECTIVES } from '../push-image/index';
import { INLINE_ALERT_DIRECTIVES } from '../inline-alert/index';


import { ErrorHandler } from '../error-handler/error-handler';
import { Repository, RepositoryItem, Tag, SystemInfo, Label } from '../service/interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { RepositoryService, RepositoryDefaultService } from '../service/repository.service';
import { SystemInfoService, SystemInfoDefaultService } from '../service/system-info.service';
import { TagService, TagDefaultService } from '../service/tag.service';
import { ChannelService } from '../channel/index';
import { LabelPieceComponent } from "../label-piece/label-piece.component";
import { LabelDefaultService, LabelService } from "../service/label.service";
import { OperationService } from "../operation/operation.service";
import { ProjectDefaultService, ProjectService, RetagDefaultService, RetagService } from "../service";
import { UserPermissionDefaultService, UserPermissionService } from "../service/permission.service";
import { USERSTATICPERMISSION } from "../service/permission-static";
import { of } from "rxjs";
import { delay } from 'rxjs/operators';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";


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
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        RouterTestingModule,
        BrowserAnimationsModule
      ],
      declarations: [
        RepositoryComponent,
        GridViewComponent,
        ConfirmationDialogComponent,
        ImageNameInputComponent,
        FilterComponent,
        TagComponent,
        LabelPieceComponent,
        VULNERABILITY_DIRECTIVES,
        PUSH_IMAGE_BUTTON_DIRECTIVES,
        INLINE_ALERT_DIRECTIVES,
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
        { provide: OperationService }
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
    spyOn(userPermissionService, "getPermission")
    .withArgs(compRepo.projectId, USERSTATICPERMISSION.REPOSITORY_TAG_LABEL.KEY, USERSTATICPERMISSION.REPOSITORY_TAG_LABEL.VALUE.CREATE )
    .and.returnValue(of(mockHasAddLabelImagePermission))
     .withArgs(compRepo.projectId, USERSTATICPERMISSION.REPOSITORY.KEY, USERSTATICPERMISSION.REPOSITORY.VALUE.PULL )
     .and.returnValue(of(mockHasRetagImagePermission))
     .withArgs(compRepo.projectId, USERSTATICPERMISSION.REPOSITORY_TAG.KEY, USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.DELETE )
     .and.returnValue(of(mockHasDeleteImagePermission))
     .withArgs(compRepo.projectId, USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.KEY
      , USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.VALUE.CREATE)
     .and.returnValue(of(mockHasScanImagePermission));
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
