import { ComponentFixture, TestBed, async, fakeAsync, tick } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';
import { Router } from '@angular/router';

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { TagComponent } from './tag.component';

import { ErrorHandler } from '../error-handler/error-handler';
import { Tag } from '../service/interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { TagService, TagDefaultService, ScanningResultService, ScanningResultDefaultService } from '../service/index';
import { VULNERABILITY_DIRECTIVES } from '../vulnerability-scanning/index';
import { FILTER_DIRECTIVES } from '../filter/index'

import { Observable, Subscription } from 'rxjs/Rx';
import { ChannelService } from '../channel/index';

import { JobLogViewerComponent } from '../job-log-viewer/index';

describe('TagComponent (inline template)', () => {

  let comp: TagComponent;
  let fixture: ComponentFixture<TagComponent>;
  let tagService: TagService;
  let spy: jasmine.Spy;
  let mockTags: Tag[] = [
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
    repositoryBaseEndpoint: '/api/repositories/testing'
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [
        TagComponent,
        ConfirmationDialogComponent,
        VULNERABILITY_DIRECTIVES,
        FILTER_DIRECTIVES,
        JobLogViewerComponent
      ],
      providers: [
        ErrorHandler,
        ChannelService,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: TagService, useClass: TagDefaultService },
        { provide: ScanningResultService, useClass: ScanningResultDefaultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TagComponent);
    comp = fixture.componentInstance;

    comp.projectId = 1;
    comp.repoName = 'library/nginx';
    comp.hasProjectAdminRole = true;
    comp.hasSignedIn = true;
    comp.registryUrl = 'http://registry.testing.com';
    comp.withNotary = false;

    tagService = fixture.debugElement.injector.get(TagService);
    spy = spyOn(tagService, 'getTags').and.returnValues(Promise.resolve(mockTags));
    fixture.detectChanges();
  });

  it('should load data', async(() => {
    expect(spy.calls.any).toBeTruthy();
  }));

  it('should load and render data', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(del => del.classes['datagrid-cell']);
      fixture.detectChanges();
      expect(de).toBeTruthy();
      let el: HTMLElement = de.nativeElement;
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('1.11.5');
    });
  }));

});