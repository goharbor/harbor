import { ComponentFixture, TestBed, async } from '@angular/core/testing'; 
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { RepositoryStackviewComponent } from './repository-stackview.component';
import { TagComponent } from '../tag/tag.component';
import { FilterComponent } from '../filter/filter.component';

import { ErrorHandler } from '../error-handler/error-handler';
import { Repository, Tag } from '../service/interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { RepositoryService, RepositoryDefaultService } from '../service/repository.service';
import { TagService, TagDefaultService } from '../service/tag.service';

import { click } from '../utils';

describe('RepositoryComponentStackview (inline template)', ()=> {
  
  let compRepo: RepositoryStackviewComponent;
  let fixtureRepo: ComponentFixture<RepositoryStackviewComponent>;
  let repositoryService: RepositoryService;
  let spyRepos: jasmine.Spy;

  let compTag: TagComponent;
  let fixtureTag: ComponentFixture<TagComponent>;
  let tagService: TagService;
  let spyTags: jasmine.Spy;

  let mockRepoData: Repository[] = [
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

  let mockTagData: Tag[] = [
    {
      "digest": "sha256:e5c82328a509aeb7c18c1d7fb36633dc638fcf433f651bdcda59c1cc04d3ee55",
      "name": "1.11.5",
      "architecture": "amd64",
      "os": "linux",
      "docker_version": "1.12.3",
      "author": "NGINX Docker Maintainers \"docker-maint@nginx.com\"",
      "created": new Date("2016-11-08T22:41:15.912313785Z"),
      "signature": null
    }
  ];

  let config: IServiceConfig = {
      repositoryBaseEndpoint: '/api/repository/testing'
  };

  beforeEach(async(()=>{
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [
        RepositoryStackviewComponent,
        TagComponent,
        ConfirmationDialogComponent,
        FilterComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue : config },
        { provide: RepositoryService, useClass: RepositoryDefaultService },
        { provide: TagService, useClass: TagDefaultService }
      ]
    });
  }));

  beforeEach(()=>{
    fixtureRepo = TestBed.createComponent(RepositoryStackviewComponent);
    compRepo = fixtureRepo.componentInstance;
    compRepo.projectId = 1;
    compRepo.sessionInfo = {
      hasProjectAdminRole: true
    };
    repositoryService = fixtureRepo.debugElement.injector.get(RepositoryService);

    spyRepos = spyOn(repositoryService, 'getRepositories').and.returnValues(Promise.resolve(mockRepoData));
    fixtureRepo.detectChanges();
   });

   beforeEach(()=>{
    fixtureTag = TestBed.createComponent(TagComponent);
    compTag = fixtureTag.componentInstance;
    compTag.projectId = compRepo.projectId;
    compTag.repoName = 'library/busybox';
    compTag.sessionInfo = compRepo.sessionInfo;
    tagService = fixtureTag.debugElement.injector.get(TagService);
    spyTags = spyOn(tagService, 'getTags').and.returnValues(Promise.resolve(mockTagData));
    fixtureTag.detectChanges();
  });

  it('should load and render data', async(()=>{
    fixtureRepo.detectChanges();
    fixtureRepo.whenStable().then(()=>{
      fixtureRepo.detectChanges();
      let deRepo: DebugElement = fixtureRepo.debugElement.query(By.css('datagrid-cell'));
      fixtureRepo.detectChanges();
      expect(deRepo).toBeTruthy();
      let elRepo: HTMLElement = deRepo.nativeElement;
      fixtureRepo.detectChanges();
      expect(elRepo).toBeTruthy();
      fixtureRepo.detectChanges();
      expect(elRepo.textContent).toEqual('library/busybox');
      click(deRepo);  
      fixtureTag.detectChanges();
      let deTag: DebugElement = fixtureTag.debugElement.query(By.css('datagrid-cell'));
      expect(deTag).toBeTruthy();
      let elTag: HTMLElement = deTag.nativeElement;
      expect(elTag).toBeTruthy();
      expect(elTag.textContent).toEqual('1.12.5');
    });
  }));

  it('should filter data by keyword', async(()=>{
    fixtureRepo.detectChanges();
    fixtureRepo.whenStable().then(()=>{
      fixtureRepo.detectChanges();
      compRepo.doSearchRepoNames('nginx');
      fixtureRepo.detectChanges();
      let de: DebugElement[] = fixtureRepo.debugElement.queryAll(By.css('datagrid-cell'));
      fixtureRepo.detectChanges();
      expect(de).toBeTruthy();
      expect(de.length).toEqual(1);
      let el: HTMLElement = de[0].nativeElement;
      fixtureRepo.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent).toEqual('library/nginx');
    });
  }));

});