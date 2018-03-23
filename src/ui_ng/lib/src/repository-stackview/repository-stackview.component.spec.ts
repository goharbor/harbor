import { ComponentFixture, TestBed, async } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { RepositoryStackviewComponent } from './repository-stackview.component';
import { TagComponent } from '../tag/tag.component';
import { FilterComponent } from '../filter/filter.component';

import { ErrorHandler } from '../error-handler/error-handler';
import { Repository, RepositoryItem, Tag, SystemInfo } from '../service/interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { RepositoryService, RepositoryDefaultService } from '../service/repository.service';
import { TagService, TagDefaultService } from '../service/tag.service';
import { SystemInfoService, SystemInfoDefaultService } from '../service/system-info.service';
import { VULNERABILITY_DIRECTIVES } from '../vulnerability-scanning/index';
import { PUSH_IMAGE_BUTTON_DIRECTIVES } from '../push-image/index';
import { INLINE_ALERT_DIRECTIVES } from '../inline-alert/index';
import { JobLogViewerComponent } from '../job-log-viewer/index';

import { click } from '../utils';

describe('RepositoryComponentStackview (inline template)', () => {

  let compRepo: RepositoryStackviewComponent;
  let fixtureRepo: ComponentFixture<RepositoryStackviewComponent>;
  let repositoryService: RepositoryService;
  let tagService: TagService;
  let systemInfoService: SystemInfoService;

  let spyRepos: jasmine.Spy;
  let spyTags: jasmine.Spy;
  let spySystemInfo: jasmine.Spy;

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
      "description": "",
      "pull_count": 0,
      "star_count": 0,
      "tags_count": 1
    },
    {
      "id": 2,
      "name": "library/nginx",
      "project_id": 1,
      "description": "",
      "pull_count": 0,
      "star_count": 0,
      "tags_count": 1
    }
  ];

  let mockRepo: Repository = {
    metadata: {xTotalCount: 2},
    data: mockRepoData
  };

  let mockTagData: Tag[] = [
    {
      "digest": "sha256:e5c82328a509aeb7c18c1d7fb36633dc638fcf433f651bdcda59c1cc04d3ee55",
      "name": "1.11.5",
      "size": "2049",
      "architecture": "amd64",
      "os": "linux",
      "docker_version": "1.12.3",
      "author": "NGINX Docker Maintainers \"docker-maint@nginx.com\"",
      "created": new Date("2016-11-08T22:41:15.912313785Z"),
      "signature": null
    }
  ];

  let config: IServiceConfig = {
    repositoryBaseEndpoint: '/api/repository/testing',
    systemInfoEndpoint: '/api/systeminfo/testing',
    targetBaseEndpoint: '/api/tag/testing'
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [
        RepositoryStackviewComponent,
        TagComponent,
        ConfirmationDialogComponent,
        FilterComponent,
        VULNERABILITY_DIRECTIVES,
        PUSH_IMAGE_BUTTON_DIRECTIVES,
        INLINE_ALERT_DIRECTIVES,
        JobLogViewerComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: RepositoryService, useClass: RepositoryDefaultService },
        { provide: TagService, useClass: TagDefaultService },
        { provide: SystemInfoService, useClass: SystemInfoDefaultService }
      ]
    });
  }));

  beforeEach(() => {
    fixtureRepo = TestBed.createComponent(RepositoryStackviewComponent);
    compRepo = fixtureRepo.componentInstance;
    compRepo.projectId = 1;
    compRepo.hasProjectAdminRole = true;

    repositoryService = fixtureRepo.debugElement.injector.get(RepositoryService);
    systemInfoService = fixtureRepo.debugElement.injector.get(SystemInfoService);

    spyRepos = spyOn(repositoryService, 'getRepositories').and.returnValues(Promise.resolve(mockRepo));
    spySystemInfo = spyOn(systemInfoService, 'getSystemInfo').and.returnValues(Promise.resolve(mockSystemInfo));
    fixtureRepo.detectChanges();
  });

  it('should create', () => {
    expect(compRepo).toBeTruthy();
  });

  it('should load and render data', async(() => {
    fixtureRepo.detectChanges();

    fixtureRepo.whenStable().then(() => {
      fixtureRepo.detectChanges();

      let deRepo: DebugElement = fixtureRepo.debugElement.query(By.css('datagrid-cell'));
      expect(deRepo).toBeTruthy();
      let elRepo: HTMLElement = deRepo.nativeElement;
      expect(elRepo).toBeTruthy();
      expect(elRepo.textContent).toEqual('library/busybox');
    });
  }));

  it('should filter data by keyword', async(() => {
    fixtureRepo.detectChanges();

    fixtureRepo.whenStable().then(() => {
      fixtureRepo.detectChanges();

      compRepo.doSearchRepoNames('nginx');
      fixtureRepo.detectChanges();
      let de: DebugElement[] = fixtureRepo.debugElement.queryAll(By.css('datagrid-cell'));
      expect(de).toBeTruthy();
      expect(de.length).toEqual(1);
      let el: HTMLElement = de[0].nativeElement;
      expect(el).toBeTruthy();
      expect(el.textContent).toEqual('library/nginx');
    });
  }));

  it('should display embedded tag view when click >', async(() => {
    fixtureRepo.detectChanges();

    fixtureRepo.whenStable().then(() => {
      fixtureRepo.detectChanges();

      let el: HTMLElement = fixtureRepo.nativeElement.querySelector('.datagrid-expandable-caret');
      expect(el).toBeTruthy();
      let button: HTMLButtonElement = el.querySelector('button');
      expect(button).toBeTruthy();
      click(button);

      fixtureRepo.detectChanges();
      let el2: HTMLElement = fixtureRepo.nativeElement.querySelector('.datagrid-row-detail');
      expect(el2).toBeTruthy();
      let el3: Element = el2.querySelector(".datagrid-cell");
      expect(el3).toBeTruthy();
      expect(el3.textContent).toEqual('1.11.5');
    });
  }));

});