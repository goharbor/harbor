import { ComponentFixture, TestBed, async } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { DebugElement } from '@angular/core';

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';

import { ListReplicationRuleComponent } from '../list-replication-rule/list-replication-rule.component';
import { ReplicationRule } from '../service/interface';

import { ErrorHandler } from '../error-handler/error-handler';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { ReplicationService, ReplicationDefaultService } from '../service/replication.service';
import { OperationService } from "../operation/operation.service";
import { of } from 'rxjs';
import { EndpointService, EndpointDefaultService } from "../service/endpoint.service";
import { Endpoint } from "../service/interface";

describe('ListReplicationRuleComponent (inline template)', () => {
  let mockEndpoint: Endpoint = {
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
  };

  let mockRules: ReplicationRule[] = [
    {
        "id": 1,
        "name": "sync_01",
        "description": "",
        "filters": null,
        "trigger": {"kind": "Manual", "schedule_param": null},
        "error_job_count": 2,
        "deletion": false,
        "src_namespaces": ["name1", "name2"],
        "src_registry_id": 3
    },
    {
          "id": 2,
          "name": "sync_02",
          "description": "",
          "filters": null,
          "trigger": {"kind": "Manual", "schedule_param": null},
          "error_job_count": 2,
          "deletion": false,
          "src_namespaces": ["name1", "name2"],
          "dest_registry_id": 3
      },
  ];

  let fixture: ComponentFixture<ListReplicationRuleComponent>;

  let comp: ListReplicationRuleComponent;

  let replicationService: ReplicationService;

  let endpointService: EndpointService;


  let spyRules: jasmine.Spy;

  let spyEndpoint: jasmine.Spy;

  let config: IServiceConfig = {
    replicationRuleEndpoint: '/api/policies/replication/testing',
    systemInfoEndpoint: "/api/endpoints/testing"
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        NoopAnimationsModule
      ],
      declarations: [
        ListReplicationRuleComponent,
        ConfirmationDialogComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ReplicationService, useClass: ReplicationDefaultService },
        { provide: OperationService },
        { provide: EndpointService, useClass: EndpointDefaultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ListReplicationRuleComponent);
    comp = fixture.componentInstance;
    replicationService = fixture.debugElement.injector.get(ReplicationService);
    spyRules = spyOn(replicationService, 'getReplicationRules').and.returnValues(of(mockRules));

    endpointService = fixture.debugElement.injector.get(EndpointService);
    spyEndpoint = spyOn(endpointService, "getEndpoint").and.returnValue(
      of(mockEndpoint)
    );
    fixture.detectChanges();
  });

  it('Should load and render data', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(By.css('datagrid-cell'));
      expect(de).toBeTruthy();
      fixture.detectChanges();
      let el: HTMLElement = de.nativeElement;
      expect(el.textContent.trim()).toEqual('sync_01');
    });
  }));

});
