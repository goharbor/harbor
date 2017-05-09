import { TestBed, inject, async } from '@angular/core/testing';

import { TagService, TagDefaultService } from './tag.service';
import { SharedModule } from '../shared/shared.module';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';

import { Tag, TagCompatibility, TagManifest } from './interface';

import { VerifiedSignature } from './tag.service';
import { toPromise } from '../utils';

describe('TagService', () => {
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

  let mockSignatures: VerifiedSignature[] = [{
    tag: '1.11.5',
    hashes: {
      sha256: 'fake'
    }
  }];

  let mockSignatures2: VerifiedSignature[] = [{
    tag: '1.11.15',
    hashes: {
      sha256: 'fake2'
    }
  }];

  beforeEach(() => {
    const mockConfig: IServiceConfig = {
      repositoryBaseEndpoint: "/api/repositories/testing"
    };

    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        TagDefaultService,
        {
          provide: TagService,
          useClass: TagDefaultService
        }, {
          provide: SERVICE_CONFIG,
          useValue: mockConfig
        }]
    });
  });

  it('should be initialized', inject([TagDefaultService], (service: TagService) => {
    expect(service).toBeTruthy();
  }));

  it('should get tags with signed status[1] if signatures exists', async(inject([TagDefaultService], (service: TagService) => {
    expect(service).toBeTruthy();
    let spy1: jasmine.Spy = spyOn(service, '_getTags')
      .and.returnValue(Promise.resolve(mockTags));
    let spy2: jasmine.Spy = spyOn(service, '_getSignatures')
      .and.returnValue(Promise.resolve(mockSignatures));

    toPromise<Tag[]>(service.getTags('library/nginx'))
      .then(tags => {
        expect(tags).toBeTruthy();
        expect(tags.length).toBe(1);
        expect(tags[0].signed).toBe(1);
      });
  })));

  it('should get tags with not-signed status[0] if signatures exists', async(inject([TagDefaultService], (service: TagService) => {
    expect(service).toBeTruthy();
    let spy1: jasmine.Spy = spyOn(service, '_getTags')
      .and.returnValue(Promise.resolve(mockTags));
    let spy2: jasmine.Spy = spyOn(service, '_getSignatures')
      .and.returnValue(Promise.resolve(mockSignatures2));

    toPromise<Tag[]>(service.getTags('library/nginx'))
      .then(tags => {
        expect(tags).toBeTruthy();
        expect(tags.length).toBe(1);
        expect(tags[0].signed).toBe(0);
      });
  })));

  it('should get tags with default signed status[-1] if signatures not exist', async(inject([TagDefaultService], (service: TagService) => {
    expect(service).toBeTruthy();
    let spy1: jasmine.Spy = spyOn(service, '_getTags')
      .and.returnValue(Promise.resolve(mockTags));
    let spy2: jasmine.Spy = spyOn(service, '_getSignatures')
      .and.returnValue(Promise.reject("Error"));

    toPromise<Tag[]>(service.getTags('library/nginx'))
      .then(tags => {
        expect(tags).toBeTruthy();
        expect(tags.length).toBe(1);
        expect(tags[0].signed).toBe(-1);
      });
  })));


});
