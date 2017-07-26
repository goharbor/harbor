import { ComponentFixture, TestBed, async } from '@angular/core/testing'; 
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ReplicationComponent } from '../replication/replication.component';

import { ListReplicationRuleComponent } from '../list-replication-rule/list-replication-rule.component';

import { CreateEditRuleComponent } from './create-edit-rule.component';
import { DatePickerComponent } from '../datetime-picker/datetime-picker.component';
import { DateValidatorDirective } from '../datetime-picker/date-validator.directive';
import { FilterComponent } from '../filter/filter.component';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';
import { ReplicationRule, ReplicationJob, Endpoint } from '../service/interface';

import { ErrorHandler } from '../error-handler/error-handler';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { 
  ReplicationService, 
  ReplicationDefaultService,
  JobLogService,
  JobLogDefaultService
 } from '../service/index';
import { EndpointService, EndpointDefaultService } from '../service/endpoint.service';
import { JobLogViewerComponent } from '../job-log-viewer/job-log-viewer.component';

describe('CreateEditRuleComponent (inline template)', ()=>{

  let mockRules: ReplicationRule[] = [
    {
        "id": 1,
        "project_id": 1,
        "project_name": "library",
        "target_id": 1,
        "target_name": "target_01",
        "name": "sync_01",
        "enabled": 0,
        "description": "",
        "cron_str": "",    
        "error_job_count": 2,
        "deleted": 0
    },
    {
        "id": 2,
        "project_id": 1,
        "project_name": "library",
        "target_id": 3,
        "target_name": "target_02",
        "name": "sync_02",
        "enabled": 1,
        "description": "",
        "cron_str": "",
        "error_job_count": 1,
        "deleted": 0
    },
    {
        "id": 3,
        "project_id": 1,
        "project_name": "library",
        "target_id": 2,
        "target_name": "target_03",
        "name": "sync_03",
        "enabled": 0,
        "description": "",
        "cron_str": "",
        "error_job_count": 0,
        "deleted": 0
    }
  ];

  let mockJobs: ReplicationJob[] = [
    {
        "id": 1,
        "status": "stopped",
        "repository": "library/busybox",
        "policy_id": 1,
        "operation": "transfer",
        "tags": null
    },
    {
        "id": 2,
        "status": "stopped",
        "repository": "library/busybox",
        "policy_id": 1,
        "operation": "transfer",
        "tags": null
    },
    {
        "id": 3,
        "status": "stopped",
        "repository": "library/busybox",
        "policy_id": 2,
        "operation": "transfer",
        "tags": null  
    }
  ];

  let mockEndpoints: Endpoint[] = [
    {
        "id": 1,
        "endpoint": "https://10.117.4.151",
        "name": "target_01",
        "username": "admin",
        "password": "",
        "type": 0
    },
    {
        "id": 2,
        "endpoint": "https://10.117.5.142",
        "name": "target_02",
        "username": "AAA",
        "password": "",
        "type": 0
    },
    {
        "id": 3,
        "endpoint": "https://101.1.11.111",
        "name": "target_03",
        "username": "admin",
        "password": "",
        "type": 0
    },
    {
        "id": 4,
        "endpoint": "http://4.4.4.4",
        "name": "target_04",
        "username": "",
        "password": "",
        "type": 0
    }
  ];

  let mockRule: ReplicationRule = {
      "id": 1,
      "project_id": 1,
      "project_name": "library",
      "target_id": 1,
      "target_name": "target_01",
      "name": "sync_01",
      "enabled": 0,
      "description": "",
      "cron_str": "",    
      "error_job_count": 2,
      "deleted": 0
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
    replicationRuleEndpoint: '/api/policies/replication/testing',
    replicationJobEndpoint: '/api/jobs/replication/testing',
    targetBaseEndpoint: '/api/targets/testing'
  };

  beforeEach(async(()=>{
    TestBed.configureTestingModule({
      imports: [ 
        SharedModule,
        NoopAnimationsModule
      ],
      declarations: [
        ReplicationComponent,
        ListReplicationRuleComponent,
        CreateEditRuleComponent,
        ConfirmationDialogComponent,
        DatePickerComponent,
        FilterComponent,
        InlineAlertComponent,
        JobLogViewerComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ReplicationService, useClass: ReplicationDefaultService },
        { provide: EndpointService, useClass: EndpointDefaultService },
        { provide: JobLogService, useClass: JobLogDefaultService }
      ]
    });
  }));

  beforeEach(()=>{
    fixture = TestBed.createComponent(ReplicationComponent);

    comp = fixture.componentInstance;
    comp.projectId = 1;
    comp.search.ruleId = 1;

    replicationService = fixture.debugElement.injector.get(ReplicationService);
   
    spyRules = spyOn(replicationService, 'getReplicationRules').and.returnValues(Promise.resolve(mockRules));
    spyOneRule = spyOn(replicationService, 'getReplicationRule').and.returnValue(Promise.resolve(mockRule));
    spyJobs = spyOn(replicationService, 'getJobs').and.returnValues(Promise.resolve(mockJobs));
    fixture.detectChanges();
  });

  beforeEach(()=>{
    fixtureCreate = TestBed.createComponent(CreateEditRuleComponent);
    
    compCreate = fixtureCreate.componentInstance;
    compCreate.projectId = 1;

    endpointService = fixtureCreate.debugElement.injector.get(EndpointService);
    spyEndpoint = spyOn(endpointService, 'getEndpoints').and.returnValues(Promise.resolve(mockEndpoints));
    fixture.detectChanges();
  });

  it('Should open creation modal and load endpoints', async(()=>{
    fixture.detectChanges();
    comp.openModal();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(By.css('input'));
      expect(de).toBeTruthy();
      let deSelect: DebugElement = fixture.debugElement.query(By.css('select'));
      expect(deSelect).toBeTruthy();
      let elSelect: HTMLElement = de.nativeElement;
      expect(elSelect).toBeTruthy();
      expect(elSelect.childNodes.item(0).textContent).toEqual('target_01');
    });
  }));

  it('Should open modal to edit replication rule', async(()=>{
    fixture.detectChanges();
    comp.openEditRule(mockRule);
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(By.css('input'));
      expect(de).toBeTruthy();
      fixture.detectChanges();
      let el: HTMLElement = de.nativeElement;
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('sync_01');
    });
  }));
});