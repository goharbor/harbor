import { ComponentFixture, TestBed, async } from "@angular/core/testing";
import { By } from "@angular/platform-browser";
import { DebugElement } from "@angular/core";
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from "../../utils/shared/shared.module";
import { ConfirmationDialogComponent } from "../confirmation-dialog/confirmation-dialog.component";
import { ReplicationComponent } from "../replication/replication.component";

import { ListReplicationRuleComponent } from "../list-replication-rule/list-replication-rule.component";
import { CronTooltipComponent } from "../cron-schedule/cron-tooltip/cron-tooltip.component";
import { CreateEditRuleComponent } from "./create-edit-rule.component";
import { DatePickerComponent } from "../datetime-picker/datetime-picker.component";
import { FilterComponent } from "../filter/filter.component";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import {
  ReplicationRule,
  ReplicationJob,
  Endpoint,
  ReplicationJobItem
} from "../../services/interface";

import { ErrorHandler } from "../../utils/error-handler/error-handler";
import { SERVICE_CONFIG, IServiceConfig } from "../../entities/service.config";
import {
  ReplicationService,
  ReplicationDefaultService,
  JobLogService,
  JobLogDefaultService
} from "../../services";
import {
  EndpointService,
  EndpointDefaultService
} from "../../services/endpoint.service";

import { OperationService } from "../operation/operation.service";
import {FilterLabelComponent} from "./filter-label.component";
import {LabelService} from "../../services/label.service";
import {LabelPieceComponent} from "../label-piece/label-piece.component";
import { RouterTestingModule } from '@angular/router/testing';
import { of } from "rxjs";

describe("CreateEditRuleComponent (inline template)", () => {
  let mockRules: ReplicationRule[] = [
    {
      id: 1,
      name: "sync_01",
      description: "",
      src_registry: {id: 2},
      src_namespaces: ["name1", "name2"],
      trigger: {
        type: "Manual",
        trigger_settings: {}
      },
      filters: [],
      deletion: false,
      enabled: true,
      override: true
    }
  ];
  let mockJobs: ReplicationJobItem[] = [
    {
      id: 1,
      status: "stopped",
      policy_id: 1,
      trigger: "Manual",
      total: 0,
      failed: 0,
      succeed: 0,
      in_progress: 0,
      stopped: 0
    },
    {
      id: 2,
      status: "stopped",
      policy_id: 1,
      trigger: "Manual",
      total: 1,
      failed: 0,
      succeed: 1,
      in_progress: 0,
      stopped: 0
    },
    {
      id: 3,
      status: "stopped",
      policy_id: 2,
      trigger: "Manual",
      total: 1,
      failed: 1,
      succeed: 0,
      in_progress: 0,
      stopped: 0
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
    src_registry: {id: 10 },
    dest_registry: {id: 0 },
    trigger: {
      type: "Manual",
      trigger_settings: {}
    },
    filters: [],
    deletion: false,
    enabled: true,
    override: true
  };

  let mockRegistryInfo = {
    "type": "harbor",
    "description": "",
    "supported_resource_filters": [
      {
        "type": "Name",
        "style": "input"
      },
      {
        "type": "Version",
        "style": "input"
      },
      {
        "type": "Label",
        "style": "input"
      },
      {
        "type": "Resource",
        "style": "radio",
        "values": [
          "repository",
          "chart"
        ]
      }
    ],
    "supported_triggers": [
      "manual",
      "scheduled",
      "event_based"
    ]
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
  let spyAdapter: jasmine.Spy;
  let spyEndpoint: jasmine.Spy;


  let config: IServiceConfig = {
    replicationBaseEndpoint: "/api/replication/testing",
    targetBaseEndpoint: "/api/registries/testing"
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [SharedModule, NoopAnimationsModule, RouterTestingModule],
      declarations: [
        ReplicationComponent,
        ListReplicationRuleComponent,
        CreateEditRuleComponent,
        CronTooltipComponent,
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

    spyAdapter = spyOn(replicationService, "getRegistryInfo").and.returnValues(
        of(mockRegistryInfo));
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
