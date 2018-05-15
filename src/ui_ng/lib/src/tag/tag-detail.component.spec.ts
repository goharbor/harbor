import { ComponentFixture, TestBed, async } from '@angular/core/testing';

import { SharedModule } from '../shared/shared.module';
import { ResultGridComponent } from '../vulnerability-scanning/result-grid.component';
import { TagDetailComponent } from './tag-detail.component';

import { ErrorHandler } from '../error-handler/error-handler';
import { Tag, VulnerabilitySummary, VulnerabilityItem, VulnerabilitySeverity } from '../service/interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { TagService, TagDefaultService, ScanningResultService, ScanningResultDefaultService } from '../service/index';
import { FilterComponent } from '../filter/index';
import { VULNERABILITY_SCAN_STATUS } from '../utils';
import {VULNERABILITY_DIRECTIVES} from "../vulnerability-scanning/index";
import {LabelPieceComponent} from "../label-piece/label-piece.component";
import {JobLogViewerComponent} from "../job-log-viewer/job-log-viewer.component";
import {ChannelService} from "../channel/channel.service";
import {JobLogService, JobLogDefaultService} from "../service/job-log.service";

describe('TagDetailComponent (inline template)', () => {

  let comp: TagDetailComponent;
  let fixture: ComponentFixture<TagDetailComponent>;
  let tagService: TagService;
  let scanningService: ScanningResultService;
  let spy: jasmine.Spy;
  let vulSpy: jasmine.Spy;
  let mockVulnerability: VulnerabilitySummary = {
    scan_status: VULNERABILITY_SCAN_STATUS.finished,
    severity: 5,
    update_time: new Date(),
    components: {
      total: 124,
      summary: [{
        severity: 1,
        count: 90
      }, {
        severity: 3,
        count: 10
      }, {
        severity: 4,
        count: 10
      }, {
        severity: 5,
        count: 13
      }]
    }
  };
  let mockTag: Tag = {
    "digest": "sha256:e5c82328a509aeb7c18c1d7fb36633dc638fcf433f651bdcda59c1cc04d3ee55",
    "name": "nginx",
    "size": "2049",
    "architecture": "amd64",
    "os": "linux",
    "docker_version": "1.12.3",
    "author": "steven",
    "created": new Date("2016-11-08T22:41:15.912313785Z"),
    "signature": null,
    "scan_overview": mockVulnerability,
    "labels": [],
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
        ResultGridComponent,
        VULNERABILITY_DIRECTIVES,
          LabelPieceComponent,
          JobLogViewerComponent,
        FilterComponent
      ],
      providers: [
        ErrorHandler,
        ChannelService,
        JobLogDefaultService,
        {provide: JobLogService, useClass: JobLogDefaultService},
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

    let mockData: VulnerabilityItem[] = [];
    for (let i = 0; i < 30; i++) {
      let res: VulnerabilityItem = {
        id: "CVE-2016-" + (8859 + i),
        severity: i % 2 === 0 ? VulnerabilitySeverity.HIGH : VulnerabilitySeverity.MEDIUM,
        package: "package_" + i,
        link: "https://security-tracker.debian.org/tracker/CVE-2016-4484",
        layer: "layer_" + i,
        version: '4.' + i + ".0",
        fixedVersion: '4.' + i + '.11',
        description: "Mock data"
      };
      mockData.push(res);
    }
    scanningService = fixture.debugElement.injector.get(ScanningResultService);
    vulSpy = spyOn(scanningService, 'getVulnerabilityScanningResults').and.returnValue(Promise.resolve(mockData));

    fixture.detectChanges();
  });

  it('should load data', async(() => {
    expect(spy.calls.any).toBeTruthy();
  }));

  it('should rightly display tag name and version', async(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let el: HTMLElement = fixture.nativeElement.querySelector('.custom-h2');
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('mock_repo:nginx');
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
      expect(el2.textContent).toEqual("steven");
    });
  }));

});
