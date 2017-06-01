import { ComponentFixture, TestBed, async, fakeAsync, tick } from '@angular/core/testing'; 
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';
import { Router } from '@angular/router';

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { TagComponent } from './tag.component';

import { ErrorHandler } from '../error-handler/error-handler';
import { Tag, TagCompatibility, TagManifest, TagView } from '../service/interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { TagService, TagDefaultService } from '../service/tag.service';

describe('TagComponent (inline template)', ()=> {
  
  let comp: TagComponent;
  let fixture: ComponentFixture<TagComponent>;
  let tagService: TagService;
  let spy: jasmine.Spy;

  let mockComp: TagCompatibility[] = [{
    v1Compatibility: '{"architecture":"amd64","author":"NGINX Docker Maintainers \\"docker-maint@nginx.com\\"","config":{"Hostname":"6b3797ab1e90","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"ExposedPorts":{"443/tcp":{},"80/tcp":{}},"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","NGINX_VERSION=1.11.5-1~jessie"],"Cmd":["nginx","-g","daemon off;"],"ArgsEscaped":true,"Image":"sha256:47a33f0928217b307cf9f20920a0c6445b34ae974a60c1b4fe73b809379ad928","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":[],"Labels":{}},"container":"f1883a3fb44b0756a2a3b1e990736a44b1387183125351370042ce7bd9ffc338","container_config":{"Hostname":"6b3797ab1e90","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"ExposedPorts":{"443/tcp":{},"80/tcp":{}},"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin","NGINX_VERSION=1.11.5-1~jessie"],"Cmd":["/bin/sh","-c","#(nop) ","CMD [\\"nginx\\" \\"-g\\" \\"daemon off;\\"]"],"ArgsEscaped":true,"Image":"sha256:47a33f0928217b307cf9f20920a0c6445b34ae974a60c1b4fe73b809379ad928","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":[],"Labels":{}},"created":"2016-11-08T22:41:15.912313785Z","docker_version":"1.12.3","id":"db3700426e6d7c1402667f42917109b2467dd49daa85d38ac99854449edc20b3","os":"linux","parent":"f3ef5f96caf99a18c6821487102c136b00e0275b1da0c7558d7090351f9d447e","throwaway":true}'
  }];
  let mockManifest: TagManifest = {
    schemaVersion: 1,
    name: 'library/nginx',
    tag: '1.11.5',
    architecture: 'amd64',
    history: mockComp
  };

  let mockTags: Tag[] = [{
    tag: '1.11.5',
    manifest: mockManifest
  }];

  let config: IServiceConfig = {
      repositoryBaseEndpoint: '/api/repositories/testing'
  };

  beforeEach(async(()=>{
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [
        TagComponent,
        ConfirmationDialogComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: TagService, useClass: TagDefaultService }
      ]
    });
  }));

  beforeEach(()=>{
    fixture = TestBed.createComponent(TagComponent);
    comp = fixture.componentInstance;

    comp.projectId = 1;
    comp.repoName = 'library/nginx';
    comp.sessionInfo = {
      hasProjectAdminRole: true,
      hasSignedIn: true,
      withNotary: true
    };
    
    tagService = fixture.debugElement.injector.get(TagService);

    spy = spyOn(tagService, 'getTags').and.returnValues(Promise.resolve(mockTags));
    fixture.detectChanges();
  });

  it('Should load data', async(()=>{
    expect(spy.calls.any).toBeTruthy();
  }));

  it('should load and render data', async(()=>{
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(del=>del.classes['datagrid-cell']);
      fixture.detectChanges();
      expect(de).toBeTruthy();
      let el: HTMLElement = de.nativeElement;
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('1.11.5');
    });
  }));

});