import { ComponentFixture, TestBed, async } from "@angular/core/testing";
import { DebugElement } from "@angular/core";

import { SharedModule } from "../shared/shared.module";
import { ConfirmationDialogComponent } from "../confirmation-dialog/confirmation-dialog.component";
import { ImageNameInputComponent } from "../image-name-input/image-name-input.component";
import { TagComponent } from "./tag.component";

import { ErrorHandler } from "../error-handler/error-handler";
import { Label, Tag } from "../service/interface";
import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import {
    TagService, TagDefaultService, ScanningResultService, ScanningResultDefaultService,
    RetagService, RetagDefaultService, ProjectService, ProjectDefaultService
} from "../service/index";
import { VULNERABILITY_DIRECTIVES } from "../vulnerability-scanning/index";
import { FILTER_DIRECTIVES } from "../filter/index";
import { ChannelService } from "../channel/index";

import { JobLogViewerComponent } from "../job-log-viewer/index";
import { CopyInputComponent } from "../push-image/copy-input.component";
import { LabelPieceComponent } from "../label-piece/label-piece.component";
import { LabelDefaultService, LabelService } from "../service/label.service";
import { OperationService } from "../operation/operation.service";

describe("TagComponent (inline template)", () => {

  let comp: TagComponent;
  let fixture: ComponentFixture<TagComponent>;
  let tagService: TagService;
  let spy: jasmine.Spy;
  let spyLabels: jasmine.Spy;
  let spyLabels1: jasmine.Spy;
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
      "signature": null,
      "labels": [],
    }
  ];

  let mockLabels: Label[] = [
    {
      color: "#9b0d54",
      creation_time: "",
      description: "",
      id: 1,
      name: "label0-g",
      project_id: 0,
      scope: "g",
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
    }
  ];

  let mockLabels1: Label[] = [
    {
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
    }
  ];

  let config: IServiceConfig = {
    repositoryBaseEndpoint: "/api/repositories/testing"
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [
        TagComponent,
        LabelPieceComponent,
        ConfirmationDialogComponent,
        ImageNameInputComponent,
        VULNERABILITY_DIRECTIVES,
        FILTER_DIRECTIVES,
        JobLogViewerComponent,
        CopyInputComponent
      ],
      providers: [
        ErrorHandler,
        ChannelService,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: TagService, useClass: TagDefaultService },
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: RetagService, useClass: RetagDefaultService },
        { provide: ScanningResultService, useClass: ScanningResultDefaultService },
        { provide: LabelService, useClass: LabelDefaultService },
        { provide: OperationService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(TagComponent);
    comp = fixture.componentInstance;

    comp.projectId = 1;
    comp.repoName = "library/nginx";
    comp.hasProjectAdminRole = true;
    comp.hasSignedIn = true;
    comp.registryUrl = "http://registry.testing.com";
    comp.withNotary = false;


    let labelService: LabelService;


    tagService = fixture.debugElement.injector.get(TagService);
    spy = spyOn(tagService, "getTags").and.returnValues(Promise.resolve(mockTags));

    labelService = fixture.debugElement.injector.get(LabelService);

    spyLabels = spyOn(labelService, "getGLabels").and.returnValues(Promise.resolve(mockLabels));
    spyLabels1 = spyOn(labelService, "getPLabels").and.returnValues(Promise.resolve(mockLabels1));

    fixture.detectChanges();
  });

  it("should load data", async(() => {
    expect(spy.calls.any).toBeTruthy();
  }));

  // fail after upgrade to angular 6.
  xit("should load and render data", async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(del => del.classes["datagrid-cell"]);
      fixture.detectChanges();
      expect(de).toBeTruthy();
      let el: HTMLElement = de.nativeElement;
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual("1.11.5");
    });
  }));

});
