import { ComponentFixture, TestBed, async } from '@angular/core/testing';

import { SharedModule } from '../shared/shared.module';
import { ResultGridComponent } from '../vulnerability-scanning/result-grid.component';
import { TagDetailComponent } from './tag-detail.component';

import { ErrorHandler } from '../error-handler/error-handler';
import { Tag, VulnerabilitySummary } from '../service/interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { TagService, TagDefaultService, ScanningResultService, ScanningResultDefaultService } from '../service/index';

describe('TagDetailComponent (inline template)', () => {

  let comp: TagDetailComponent;
  let fixture: ComponentFixture<TagDetailComponent>;
  let tagService: TagService;
  let spy: jasmine.Spy;
  let mockVulnerability: VulnerabilitySummary = {
    total_package: 124,
    package_with_none: 92,
    package_with_high: 10,
    package_with_medium: 6,
    package_With_low: 13,
    package_with_unknown: 3,
    complete_timestamp: new Date()
  };
  let mockTag: Tag = {
    "digest": "sha256:e5c82328a509aeb7c18c1d7fb36633dc638fcf433f651bdcda59c1cc04d3ee55",
    "name": "nginx",
    "architecture": "amd64",
    "os": "linux",
    "docker_version": "1.12.3",
    "author": "steven",
    "created": new Date("2016-11-08T22:41:15.912313785Z"),
    "signature": null,
    vulnerability: mockVulnerability
  };

  let config: IServiceConfig = {
    repositoryBaseEndpoint: '/api/repositories/testing'
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [
        TagDetailComponent,
        ResultGridComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: TagService, useClass: TagDefaultService },
        { provide: ScanningResultService, useClass: ScanningResultDefaultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TagDetailComponent);
    comp = fixture.componentInstance;

    comp.tagId = "mock_tag";
    comp.repositoryId = "mock_repo";

    tagService = fixture.debugElement.injector.get(TagService);
    spy = spyOn(tagService, 'getTag').and.returnValues(Promise.resolve(mockTag));

    fixture.detectChanges();
  });

  it('should load data', async(() => {
    expect(spy.calls.any).toBeTruthy();
  }));

  it('should rightly display tag name and version', async(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let el: HTMLElement = fixture.nativeElement.querySelector('.tag-name');
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('nginx:v1.12.3');
    });
  }));

  it('should display tag details', async(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let el: HTMLElement = fixture.nativeElement.querySelector('.image-detail-value');
      expect(el).toBeTruthy();
      let el2: HTMLElement = el.querySelector('div');
      expect(el2).toBeTruthy();
      expect(el2.textContent).toEqual("amd64");
    });
  }));

  it('should display vulnerability details', async(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let el: HTMLElement = fixture.nativeElement.querySelector('.second-column');
      expect(el).toBeTruthy();
      let el2: HTMLElement = el.querySelector('div');
      expect(el2).toBeTruthy();
      expect(el2.textContent.trim()).toEqual("10 VULNERABILITY.SEVERITY.HIGH VULNERABILITY.PLURAL");
    });
  }));

});