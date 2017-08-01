import { ComponentFixture, TestBed, async } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';
import { Router } from '@angular/router';

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { RepositoryComponent } from './repository.component';
import { ListRepositoryComponent } from '../list-repository/list-repository.component';
import { FilterComponent } from '../filter/filter.component';

import { ErrorHandler } from '../error-handler/error-handler';
import { Repository, RepositoryItem } from '../service/interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { RepositoryService, RepositoryDefaultService } from '../service/repository.service';
import { SystemInfoService, SystemInfoDefaultService } from '../service/system-info.service';

class RouterStub {
  navigateByUrl(url: string) { return url; }
}

describe('RepositoryComponent (inline template)', () => {

  let comp: RepositoryComponent;
  let fixture: ComponentFixture<RepositoryComponent>;
  let repositoryService: RepositoryService;
  let spy: jasmine.Spy;

  let mockData: RepositoryItem[] = [
    {
      "id": 11,
      "name": "library/busybox",
      "project_id": 1,
      "description": "",
      "pull_count": 0,
      "star_count": 0,
      "tags_count": 1
    },
    {
      "id": 12,
      "name": "library/nginx",
      "project_id": 1,
      "description": "",
      "pull_count": 0,
      "star_count": 0,
      "tags_count": 1
    }
  ];
  let mockRepo: Repository = {
    metadata: { xTotalCount: 2 },
    data: mockData
  };

  let config: IServiceConfig = {
    repositoryBaseEndpoint: '/api/repository/testing'
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [
        RepositoryComponent,
        ListRepositoryComponent,
        ConfirmationDialogComponent,
        FilterComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: RepositoryService, useClass: RepositoryDefaultService },
        { provide: SystemInfoService, useClass: SystemInfoDefaultService },
        { provide: Router, useClass: RouterStub }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RepositoryComponent);
    comp = fixture.componentInstance;
    comp.projectId = 1;
    comp.hasProjectAdminRole = true;
    repositoryService = fixture.debugElement.injector.get(RepositoryService);

    spy = spyOn(repositoryService, 'getRepositories').and.returnValues(Promise.resolve(mockRepo));
    fixture.detectChanges();
  });

  it('should load and render data', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(By.css('datagrid-cell'));
      fixture.detectChanges();
      expect(de).toBeTruthy();
      let el: HTMLElement = de.nativeElement;
      expect(el).toBeTruthy();
      expect(el.textContent).toEqual('library/busybox');
    });
  }));

  it('should filter data by keyword', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      comp.doSearchRepoNames('nginx');
      fixture.detectChanges();
      let de: DebugElement[] = fixture.debugElement.queryAll(By.css('datagrid-cell'));
      fixture.detectChanges();
      expect(de).toBeTruthy();
      expect(de.length).toEqual(1);
      let el: HTMLElement = de[0].nativeElement;
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent).toEqual('library/nginx');
    });
  }));

});