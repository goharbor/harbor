import { ComponentFixture, TestBed, async } from "@angular/core/testing";
import { CUSTOM_ELEMENTS_SCHEMA, DebugElement } from "@angular/core";
import { ArtifactListTabComponent } from "./artifact-list-tab.component";
import { of } from "rxjs";
import { delay } from "rxjs/operators";
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { HttpClientTestingModule } from "@angular/common/http/testing";
import { HttpClient } from "@angular/common/http";
import { ActivatedRoute, Router } from "@angular/router";
import { ArtifactDefaultService, ArtifactService } from "../../../artifact/artifact.service";
import {
  Label,
  LabelDefaultService,
  LabelService,
  ProjectDefaultService,
  ProjectService,
  RetagDefaultService,
  RetagService,
  ScanningResultDefaultService,
  ScanningResultService,
  UserPermissionDefaultService,
  UserPermissionService,
  USERSTATICPERMISSION
} from "../../../../../../lib/services";
import { Artifact, Reference } from "../../../artifact/artifact";
import { IServiceConfig, SERVICE_CONFIG } from "../../../../../../lib/entities/service.config";
import { SharedModule } from "../../../../../../lib/utils/shared/shared.module";
import { LabelPieceComponent } from "../../../../../../lib/components/label-piece/label-piece.component";
import { ConfirmationDialogComponent } from "../../../../../../lib/components/confirmation-dialog";
import { ImageNameInputComponent } from "../../../../../../lib/components/image-name-input/image-name-input.component";
import { CopyInputComponent } from "../../../../../../lib/components/push-image/copy-input.component";
import { ErrorHandler } from "../../../../../../lib/utils/error-handler";
import { ChannelService } from "../../../../../../lib/services/channel.service";
import { OperationService } from "../../../../../../lib/components/operation/operation.service";

describe("ArtifactListTabComponent (inline template)", () => {

  let comp: ArtifactListTabComponent;
  let fixture: ComponentFixture<ArtifactListTabComponent>;
  let artifactService: ArtifactService;
  let userPermissionService: UserPermissionService;
  let spy: jasmine.Spy;
  let spyLabels: jasmine.Spy;
  let spyLabels1: jasmine.Spy;
  let spyScanner: jasmine.Spy;
  let scannerMock = {
    disabled: false,
    name: "Clair"
  };
  let mockActivatedRoute = {
    snapshot: {
      params: {
        id: 1,
        repo: "test",
        digest: "ABC",
        subscribe: () => {
          return of(null);
        }
      },
      data: {
        projectResolver: {
          has_project_admin_role: true,
          current_user_role_id: 3,
          name: "demo"
        }
      }
    },
    data: of(
      {
        projectResolver: {
          name: 'library'
        }
      }
    ),
    params: {
      subscribe: () => {
        return of(null);
      }
    }
  };
  let mockArtifacts: Artifact[] = [
    {
      "id": 1,
      type: 'image',
      repository: "goharbor/harbor-portal",
      tags: [{
        id: '1',
        name: 'tag1',
        artifact_id: 1,
        upload_time: '2020-01-06T09:40:08.036866579Z',
    },
    {
        id: '2',
        name: 'tag2',
        artifact_id: 2,
        pull_time: '2020-01-06T09:40:08.036866579Z',
        push_time: '2020-01-06T09:40:08.036866579Z',
    }
    ],
      references: [new Reference(1), new Reference(2)],
      media_type: 'string',
      "digest": "sha256:4875cda368906fd670c9629b5e416ab3d6c0292015f3c3f12ef37dc9a32fc8d4",
      "size": 20372934,
      "scan_overview": {
        "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0": {
          "report_id": "5e64bc05-3102-11ea-93ae-0242ac140004",
          "scan_status": "Error",
          "severity": "",
          "duration": 118,
          "summary": null,
          "start_time": "2020-01-07T04:01:23.157711Z",
          "end_time": "2020-01-07T04:03:21.662766Z"
        }
      },
      "labels": [
        {
          "id": 3,
          "name": "aaa",
          "description": "",
          "color": "#0095D3",
          "scope": "g",
          "project_id": 0,
          "creation_time": "2020-01-13T05:44:00.580198Z",
          "update_time": "2020-01-13T05:44:00.580198Z",
          "deleted": false
        },
        {
          "id": 6,
          "name": "dbc",
          "description": "",
          "color": "",
          "scope": "g",
          "project_id": 0,
          "creation_time": "2020-01-13T08:27:19.279123Z",
          "update_time": "2020-01-13T08:27:19.279123Z",
          "deleted": false
        }
      ],
      "push_time": "2020-01-07T03:33:41.162319Z",
      "pull_time": "0001-01-01T00:00:00Z",
      hasReferenceArtifactList: [],
      noReferenceArtifactList: []

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
  let mockHasAddLabelImagePermission: boolean = true;
  let mockHasRetagImagePermission: boolean = true;
  let mockHasDeleteImagePermission: boolean = true;
  let mockHasScanImagePermission: boolean = true;
  const mockErrorHandler = {
    error: () => {}
  };
  const permissions = [
    {resource: USERSTATICPERMISSION.REPOSITORY_TAG_LABEL.KEY, action:  USERSTATICPERMISSION.REPOSITORY_TAG_LABEL.VALUE.CREATE},
    {resource: USERSTATICPERMISSION.REPOSITORY.KEY, action:  USERSTATICPERMISSION.REPOSITORY.VALUE.PULL},
    {resource: USERSTATICPERMISSION.REPOSITORY_TAG.KEY, action:  USERSTATICPERMISSION.REPOSITORY_TAG.VALUE.DELETE},
    {resource: USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.KEY, action:  USERSTATICPERMISSION.REPOSITORY_TAG_SCAN_JOB.VALUE.CREATE},
  ];
  const mockRouter = {
    navigate: () => { }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        BrowserAnimationsModule,
        HttpClientTestingModule,
      ],
      schemas: [
        CUSTOM_ELEMENTS_SCHEMA
      ],
      declarations: [
        ArtifactListTabComponent,
        LabelPieceComponent,
        ConfirmationDialogComponent,
        ImageNameInputComponent,
        CopyInputComponent
      ],
      providers: [
        ErrorHandler,
        ChannelService,
        ArtifactDefaultService,
        { provide: Router, useValue: mockRouter },
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ArtifactService, useClass: ArtifactDefaultService },
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: RetagService, useClass: RetagDefaultService },
        { provide: ScanningResultService, useClass: ScanningResultDefaultService },
        { provide: LabelService, useClass: LabelDefaultService },
        { provide: UserPermissionService, useClass: UserPermissionDefaultService },
        { provide: ErrorHandler, useValue: mockErrorHandler },
        { provide: ActivatedRoute, useValue: mockActivatedRoute },
        { provide: OperationService },
      ]
    }).compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ArtifactListTabComponent);
    comp = fixture.componentInstance;
    comp.projectId = 1;
    comp.repoName = "library/nginx";
    comp.hasDeleteImagePermission = true;
    comp.hasScanImagePermission = true;
    comp.hasSignedIn = true;
    comp.registryUrl = "http://registry.testing.com";
    comp.withNotary = false;
    comp.withAdmiral = false;
    let labelService: LabelService;
    artifactService = fixture.debugElement.injector.get(ArtifactService);
    spy = spyOn(artifactService, "getArtifactList").and.returnValues(of(
      {
        body: mockArtifacts
      }
      ).pipe(delay(0)));
    userPermissionService = fixture.debugElement.injector.get(UserPermissionService);
    let http: HttpClient;
    http = fixture.debugElement.injector.get(HttpClient);
    spyScanner = spyOn(http, "get").and.returnValue(of(scannerMock));
    spyOn(userPermissionService, "hasProjectPermissions")
    .withArgs(comp.projectId, permissions )
    .and.returnValue(of([mockHasAddLabelImagePermission, mockHasRetagImagePermission,
       mockHasDeleteImagePermission, mockHasScanImagePermission]));
    labelService = fixture.debugElement.injector.get(LabelService);
    spyLabels = spyOn(labelService, "getGLabels").and.returnValues(of(mockLabels).pipe(delay(0)));
    spyLabels1 = spyOn(labelService, "getPLabels").withArgs(comp.projectId).and.returnValues(of(mockLabels1).pipe(delay(0)));
    fixture.detectChanges();
  });
  it("should load data", async(() => {
    expect(spy.calls.any).toBeTruthy();
  }));
});


