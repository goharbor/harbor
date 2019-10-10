import { ComponentFixture, TestBed, async } from "@angular/core/testing";

import { SharedModule } from "../shared/shared.module";
import { ResultGridComponent } from "../vulnerability-scanning/result-grid.component";
import { TagDetailComponent } from "./tag-detail.component";
import { TagHistoryComponent } from "./tag-history.component";

import { ErrorHandler } from "../error-handler/error-handler";
import {
  Tag,
  Manifest,
  VulnerabilitySummary,
  VulnerabilityItem,
  VulnerabilitySeverity
} from "../service/interface";
import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import {
  TagService,
  TagDefaultService,
  ScanningResultService,
  ScanningResultDefaultService
} from "../service/index";
import { FilterComponent } from "../filter/index";
import { VULNERABILITY_SCAN_STATUS, VULNERABILITY_SEVERITY } from "../utils";
import { VULNERABILITY_DIRECTIVES } from "../vulnerability-scanning/index";
import { LabelPieceComponent } from "../label-piece/label-piece.component";
import { ChannelService } from "../channel/channel.service";
import { of } from "rxjs";
import {
  JobLogService,
  JobLogDefaultService
} from "../service/job-log.service";
import { UserPermissionService, UserPermissionDefaultService } from "../service/permission.service";
import { USERSTATICPERMISSION } from "../service/permission-static";

describe("TagDetailComponent (inline template)", () => {
  let comp: TagDetailComponent;
  let fixture: ComponentFixture<TagDetailComponent>;
  let tagService: TagService;
  let userPermissionService: UserPermissionService;
  let scanningService: ScanningResultService;
  let spy: jasmine.Spy;
  let vulSpy: jasmine.Spy;
  let manifestSpy: jasmine.Spy;
  let mockVulnerability: VulnerabilitySummary = {
    scan_status: VULNERABILITY_SCAN_STATUS.SUCCESS,
    severity: "High",
    end_time: new Date(),
    summary: {
      total: 124,
      summary: {
        "High": 5,
        "Low": 5
      }
    }
  };
  let mockTag: Tag = {
    digest:
      "sha256:e5c82328a509aeb7c18c1d7fb36633dc638fcf433f651bdcda59c1cc04d3ee55",
    name: "nginx",
    size: "2049",
    architecture: "amd64",
    os: "linux",
    'os.version': "",
    docker_version: "1.12.3",
    author: "steven",
    created: new Date("2016-11-08T22:41:15.912313785Z"),
    signature: null,
    scan_overview: {
      "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0": mockVulnerability
    },
    labels: []
  };

  let config: IServiceConfig = {
    repositoryBaseEndpoint: "/api/repositories/testing"
  };
  let mockHasVulnerabilitiesListPermission: boolean = false;
  let mockHasBuildHistoryPermission: boolean = true;
  let mockManifest: Manifest = {
    manifset: {},
    // tslint:disable-next-line:max-line-length
    config: `{"architecture":"amd64","config":{"Hostname":"","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"Cmd":["/bin/sh"],"ArgsEscaped":true,"Image":"sha256:fbef17698ac8605733924d5662f0cbfc0b27a51e83ab7d7a4b8d8a9a9fe0d1c2","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":null,"Labels":null},"container":"30e1a2427aa2325727a092488d304505780501585a6ccf5a6a53c4d83a826101","container_config":{"Hostname":"30e1a2427aa2","Domainname":"","User":"","AttachStdin":false,"AttachStdout":false,"AttachStderr":false,"Tty":false,"OpenStdin":false,"StdinOnce":false,"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"Cmd":["/bin/sh","-c","#(nop) ","CMD [\\"/bin/sh\\"]"],"ArgsEscaped":true,"Image":"sha256:fbef17698ac8605733924d5662f0cbfc0b27a51e83ab7d7a4b8d8a9a9fe0d1c2","Volumes":null,"WorkingDir":"","Entrypoint":null,"OnBuild":null,"Labels":{}},"created":"2018-01-09T21:10:58.579708634Z","docker_version":"17.06.2-ce","history":[{"created":"2018-01-09T21:10:58.365737589Z","created_by":"/bin/sh -c #(nop) ADD file:093f0723fa46f6cdbd6f7bd146448bb70ecce54254c35701feeceb956414622f in / "},{"created":"2018-01-09T21:10:58.579708634Z","created_by":"/bin/sh -c #(nop)  CMD [\\"/bin/sh\\"]","empty_layer":true}],"os":"linux","rootfs":{"type":"layers","diff_ids":["sha256:cd7100a72410606589a54b932cabd804a17f9ae5b42a1882bd56d263e02b6215"]}}`
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule],
      declarations: [
        TagDetailComponent,
        TagHistoryComponent,
        ResultGridComponent,
        VULNERABILITY_DIRECTIVES,
        LabelPieceComponent,
        FilterComponent
      ],
      providers: [
        ErrorHandler,
        ChannelService,
        JobLogDefaultService,
        { provide: JobLogService, useClass: JobLogDefaultService },
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: TagService, useClass: TagDefaultService },
        { provide: UserPermissionService, useClass: UserPermissionDefaultService },
        {
          provide: ScanningResultService,
          useClass: ScanningResultDefaultService
        }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TagDetailComponent);
    comp = fixture.componentInstance;

    comp.tagId = "mock_tag";
    comp.repositoryId = "mock_repo";
    comp.projectId = 1;


    tagService = fixture.debugElement.injector.get(TagService);
    spy = spyOn(tagService, "getTag").and.returnValues(
      of(mockTag)
    );

    let mockData: VulnerabilityItem[] = [];
    for (let i = 0; i < 30; i++) {
      let res: VulnerabilityItem = {
        id: "CVE-2016-" + (8859 + i),
        severity:
          i % 2 === 0
            ? VULNERABILITY_SEVERITY.HIGH
            : VULNERABILITY_SEVERITY.MEDIUM,
        package: "package_" + i,
        links: ["https://security-tracker.debian.org/tracker/CVE-2016-4484"],
        layer: "layer_" + i,
        version: "4." + i + ".0",
        fix_version: "4." + i + ".11",
        description: "Mock data"
      };
      mockData.push(res);
    }
    scanningService = fixture.debugElement.injector.get(ScanningResultService);
    vulSpy = spyOn(
      scanningService,
      "getVulnerabilityScanningResults"
    ).and.returnValue(of(mockData));
    manifestSpy = spyOn(tagService, "getManifest").and.returnValues(
      of(mockManifest)
    );
    userPermissionService = fixture.debugElement.injector.get(UserPermissionService);

    spyOn(userPermissionService, "getPermission")
    .withArgs(comp.projectId,
      USERSTATICPERMISSION.REPOSITORY_TAG_VULNERABILITY.KEY, USERSTATICPERMISSION.REPOSITORY_TAG_VULNERABILITY.VALUE.LIST )
    .and.returnValue(of(mockHasVulnerabilitiesListPermission))
     .withArgs(comp.projectId, USERSTATICPERMISSION.REPOSITORY_TAG_MANIFEST.KEY, USERSTATICPERMISSION.REPOSITORY_TAG_MANIFEST.VALUE.READ )
     .and.returnValue(of(mockHasBuildHistoryPermission));
    fixture.detectChanges();
  });

  it("should load data", async(() => {
    expect(spy.calls.any).toBeTruthy();
  }));

  it("should load history data", async(() => {
    expect(manifestSpy.calls.any).toBeTruthy();
  }));

  it("should rightly display tag name and version", async(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let el: HTMLElement = fixture.nativeElement.querySelector(".custom-h2");
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual("mock_repo:nginx");
    });
  }));

  it("should display tag details", async(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let el: HTMLElement = fixture.nativeElement.querySelector(
        ".image-detail-label .image-details"
      );
      expect(el).toBeTruthy();
      expect(el.textContent).toEqual("steven");
    });
  }));

  it("should render history info", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let els: HTMLElement[] = fixture.nativeElement.querySelectorAll(
        ".history-item"
      );
      expect(els).toBeTruthy();
      expect(els.length).toBe(2);
    });
  }));
});
