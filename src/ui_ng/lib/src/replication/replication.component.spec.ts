import { ComponentFixture, TestBed, async, inject } from '@angular/core/testing'; 
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ReplicationComponent } from './replication.component';
import { ListReplicationRuleComponent } from '../list-replication-rule/list-replication-rule.component';
import { CreateEditRuleComponent } from '../create-edit-rule/create-edit-rule.component';
import { DatePickerComponent } from '../datetime-picker/datetime-picker.component';
import { DateValidatorDirective } from '../datetime-picker/date-validator.directive';
import { FilterComponent } from '../filter/filter.component';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';
import { ReplicationRule, ReplicationJob, Endpoint } from '../service/interface';

import { ErrorHandler } from '../error-handler/error-handler';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { ReplicationService, ReplicationDefaultService } from '../service/replication.service';
import { EndpointService, EndpointDefaultService } from '../service/endpoint.service';
import { JobLogViewerComponent } from '../job-log-viewer/job-log-viewer.component';
import { JobLogService, JobLogDefaultService, ReplicationJobItem } from '../service/index';

describe('Replication Component (inline template)', ()=>{

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

  let mockJobs: ReplicationJobItem[] = [
    {
        "id": 1,
        "status": "error",
        "repository": "library/nginx",
        "policy_id": 1,
        "operation": "transfer",
        "update_time": new Date("2017-05-23 12:20:33"),
        "tags": null
    },
    {
        "id": 2,
        "status": "finished",
        "repository": "library/mysql",
        "policy_id": 1,
        "operation": "transfer",
        "update_time": new Date("2017-05-27 12:20:33"),        
        "tags": null
    },
    {
        "id": 3,
        "status": "stopped",
        "repository": "library/busybox",
        "policy_id": 2,
        "operation": "transfer",
        "update_time": new Date("2017-04-23 12:20:33"),        
        "tags": null
    }
  ];

  let mockJob: ReplicationJob = {
    metadata: {xTotalCount: 3},
    data: mockJobs
  };

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
  let comp: ReplicationComponent;
  
  let replicationService: ReplicationService;
  
  let spyRules: jasmine.Spy;
  let spyJobs: jasmine.Spy;
  
  let deGrids: DebugElement[];
  let deRules: DebugElement;
  let deJobs: DebugElement;

  let elRule: HTMLElement;
  let elJob: HTMLElement;

  let config: IServiceConfig = {
    replicationRuleEndpoint: '/api/policies/replication/testing',
    replicationJobEndpoint: '/api/jobs/replication/testing'
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
    spyJobs = spyOn(replicationService, 'getJobs').and.returnValues(Promise.resolve(mockJob));
    
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      deGrids = fixture.debugElement.queryAll(del=>del.classes['datagrid']);
      fixture.detectChanges();
      expect(deGrids).toBeTruthy();
      expect(deGrids.length).toEqual(2);
    });
  });

  it('Should load replication rules', async(()=>{    
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      deRules = deGrids[0].query(By.css('datagrid-cell'));
      expect(deRules).toBeTruthy();
      fixture.detectChanges();
      elRule = deRules.nativeElement;
      expect(elRule).toBeTruthy();
      expect(elRule.textContent).toEqual('sync_01');
    });
  }));

  it('Should load replication jobs', async(()=>{    
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      deJobs = deGrids[1].query(By.css('datagrid-cell'));
      expect(deJobs).toBeTruthy();
      fixture.detectChanges();
      elJob = deJobs.nativeElement;
      fixture.detectChanges();
      expect(elJob).toBeTruthy();
      expect(elJob.textContent).toEqual('library/nginx');
    });
  }));

  it('Should filter replication rules by keywords', async(()=>{
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();     
      comp.doSearchRules('sync_01');
      fixture.detectChanges();     
      let el: HTMLElement = deRules.nativeElement;
      fixture.detectChanges();
      expect(el.textContent.trim()).toEqual('sync_01');
    });
  }));

  it('Should filter replication rules by status', async(()=>{
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      comp.doFilterRuleStatus('1' /*Enabled*/);
      fixture.detectChanges();
      let el: HTMLElement = deRules.nativeElement;
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('sync_02');
    });
  }));

  it('Should filter replication jobs by keywords', async(()=>{
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      comp.doSearchJobs('nginx');
      fixture.detectChanges();
      let el: HTMLElement = deJobs.nativeElement;
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('library/nginx');
    });
  }));

  it('Should filter replication jobs by status', async(()=>{
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      comp.doFilterJobStatus('finished');
      let el: HTMLElement = deJobs.nativeElement;
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('library/mysql');
    });
  }));

  it('Should filter replication jobs by date range', async(()=>{
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      comp.doJobSearchByStartTime('2017-05-01');
      comp.doJobSearchByEndTime('2015-05-25');
      let el: HTMLElement = deJobs.nativeElement; 
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('library/nginx');
    });
  }))
});