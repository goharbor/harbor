import { ComponentFixture, TestBed, async } from "@angular/core/testing";
import { By } from "@angular/platform-browser";
import { DebugElement } from "@angular/core";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from "../shared/shared.module";
import { ConfirmationDialogComponent } from "../confirmation-dialog/confirmation-dialog.component";
import { ReplicationComponent } from "../replication/replication.component";

import { ListReplicationRuleComponent } from "../list-replication-rule/list-replication-rule.component";

import { CreateEditRuleComponent } from "./create-edit-rule.component";
import { DatePickerComponent } from "../datetime-picker/datetime-picker.component";
import { FilterComponent } from "../filter/filter.component";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import {
  ReplicationRule,
  ReplicationJob,
  Endpoint,
  ReplicationJobItem
} from "../service/interface";

import { ErrorHandler } from "../error-handler/error-handler";
import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import {
  ReplicationService,
  ReplicationDefaultService,
  JobLogService,
  JobLogDefaultService
} from "../service/index";
import {
  EndpointService,
  EndpointDefaultService
} from "../service/endpoint.service";
import {
  ProjectDefaultService,
  ProjectService
} from "../service/project.service";
import { OperationService } from "../operation/operation.service";
import {FilterLabelComponent} from "./filter-label.component";
import {LabelService} from "../service/label.service";
import {LabelPieceComponent} from "../label-piece/label-piece.component";
import { RouterTestingModule } from '@angular/router/testing';
import { of } from "rxjs";

describe("CreateEditRuleComponent (inline template)", () => {
  let mockRules: ReplicationRule[] = [
    {
      id: 1,
      name: "sync_01",
      description: "",
      src_registry_id: 2,
      src_namespaces: ["name1", "name2"],
      trigger: {
        kind: "Manual",
        schedule_param: null
      },
      filters: [],
      deletion: false
    }
  ];
  let mockJobs: ReplicationJobItem[] = [
    {
      id: 1,
      status: "stopped",
      repository: "library/busybox",
      policy_id: 1,
      operation: "transfer",
      tags: null
    },
    {
      id: 2,
      status: "stopped",
      repository: "library/busybox",
      policy_id: 1,
      operation: "transfer",
      tags: null
    },
    {
      id: 3,
      status: "stopped",
      repository: "library/busybox",
      policy_id: 2,
      operation: "transfer",
      tags: null
    }
  ];

  let mockJob: ReplicationJob = {
    metadata: { xTotalCount: 3 },
    data: mockJobs
  };

  let mockEndpoints: Endpoint[] = [
    {
      id: 1,
      credential: {
        access_key: "admin",
        access_secret: "",
        type: "basic"
      },
      description: "test",
      insecure: false,
      name: "target_01",
      type: "Harbor",
      url: "https://10.117.4.151"
    },
    {
      id: 2,
      credential: {
        access_key: "AAA",
        access_secret: "",
        type: "basic"
      },
      description: "test",
      insecure: false,
      name: "target_02",
      type: "Harbor",
      url: "https://10.117.5.142"
    },
    {
      id: 3,
      credential: {
        access_key: "admin",
        access_secret: "",
        type: "basic"
      },
      description: "test",
      insecure: false,
      name: "target_03",
      type: "Harbor",
      url: "https://101.1.11.111"
    },
    {
      id: 4,
      credential: {
        access_key: "admin",
        access_secret: "",
        type: "basic"
      },
      description: "test",
      insecure: true,
      name: "target_04",
      type: "Harbor",
      url: "https://4.4.4.4"
    }
  ];

  let mockRule: ReplicationRule = {
    id: 1,
    name: "sync_01",
    description: "",
    src_namespaces: ["namespace1", "namespace2"],
    src_registry_id: 10,
    trigger: {
      kind: "Manual",
      schedule_param: null
    },
    filters: [],
    deletion: false
  };

  let fixture: ComponentFixture<ReplicationComponent>;
  let fixtureCreate: ComponentFixture<CreateEditRuleComponent>;

  let comp: ReplicationComponent;
  let compCreate: CreateEditRuleComponent;

  let replicationService: ReplicationService;
  let endpointService: EndpointService;

  let spyRules: jasmine.Spy;
  let spyOneRule: jasmine.Spy;

  let spyJobs: jasmine.Spy;
  let spyEndpoint: jasmine.Spy;

  let config: IServiceConfig = {
    replicationBaseEndpoint: "/api/replication/executions/testing",
    targetBaseEndpoint: "/api/registries/testing"
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule, NoopAnimationsModule, RouterTestingModule],
      declarations: [
        ReplicationComponent,
        ListReplicationRuleComponent,
        CreateEditRuleComponent,
        ConfirmationDialogComponent,
        DatePickerComponent,
        FilterComponent,
        InlineAlertComponent,
        FilterLabelComponent,
        LabelPieceComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ReplicationService, useClass: ReplicationDefaultService },
        { provide: EndpointService, useClass: EndpointDefaultService },
        { provide: ProjectService, useClass: ProjectDefaultService },
        { provide: JobLogService, useClass: JobLogDefaultService },
        { provide: OperationService },
        { provide: LabelService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ReplicationComponent);
    fixtureCreate = TestBed.createComponent(CreateEditRuleComponent);
    comp = fixture.componentInstance;
    compCreate = fixtureCreate.componentInstance;
    comp.projectId = 1;
    comp.search.ruleId = 1;

    replicationService = fixture.debugElement.injector.get(ReplicationService);

    endpointService = fixtureCreate.debugElement.injector.get(EndpointService);

    spyRules = spyOn(
      replicationService,
      "getReplicationRules"
    ).and.returnValues(of(mockRules));
    spyOneRule = spyOn(
      replicationService,
      "getReplicationRule"
    ).and.returnValue(of(mockRule));
    spyJobs = spyOn(replicationService, "getExecutions").and.returnValues(
      of(mockJob));

    spyEndpoint = spyOn(endpointService, "getEndpoints").and.returnValues(
      of(mockEndpoints)
    );

    fixture.detectChanges();
  });

  it("Should open creation modal and load endpoints", async(() => {
    fixture.detectChanges();
    compCreate.openCreateEditRule();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(By.css("input"));
      expect(de).toBeTruthy();
      let deSelect: DebugElement = fixture.debugElement.query(By.css("select"));
      expect(deSelect).toBeTruthy();
      let elSelect: HTMLElement = de.nativeElement;
      expect(elSelect).toBeTruthy();
      expect(elSelect.childNodes.item(0).textContent).toEqual("target_01");
    });
  }));

  it("Should open modal to edit replication rule", async(() => {
    fixture.detectChanges();
    compCreate.openCreateEditRule(mockRule.id);
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(By.css("input"));
      expect(de).toBeTruthy();
      fixture.detectChanges();
      let el: HTMLElement = de.nativeElement;
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual("sync_01");
    });
  }));
});
