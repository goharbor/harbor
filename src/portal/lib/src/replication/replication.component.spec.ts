import { ComponentFixture, TestBed, async } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { SharedModule } from '../shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ReplicationComponent } from './replication.component';
import { ListReplicationRuleComponent } from '../list-replication-rule/list-replication-rule.component';
import { CreateEditRuleComponent } from '../create-edit-rule/create-edit-rule.component';
import { DatePickerComponent } from '../datetime-picker/datetime-picker.component';
import { FilterComponent } from '../filter/filter.component';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';
import {ReplicationRule, ReplicationJob, Endpoint} from '../service/interface';

import { ErrorHandler } from '../error-handler/error-handler';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { ReplicationService, ReplicationDefaultService } from '../service/replication.service';
import { EndpointService, EndpointDefaultService } from '../service/endpoint.service';
import { JobLogService, JobLogDefaultService, ReplicationJobItem } from '../service/index';
import {ProjectDefaultService, ProjectService} from "../service/project.service";
import {OperationService} from "../operation/operation.service";
import {FilterLabelComponent} from "../create-edit-rule/filter-label.component";
import {LabelPieceComponent} from "../label-piece/label-piece.component";

describe('Replication Component (inline template)', () => {

  let mockRules: ReplicationRule[] = [
      {
          "id": 1,
          "projects": [{
              "project_id": 33,
              "owner_id": 1,
              "name": "aeas",
              "deleted": 0,
              "togglable": false,
              "current_user_role_id": 0,
              "repo_count": 0,
              "metadata": {
                  "public": false,
                  "enable_content_trust": "",
                  "prevent_vul": "",
                  "severity": "",
                  "auto_scan": ""},
              "owner_name": "",
              "creation_time": null,
              "update_time": null,
              "has_project_admin_role": true,
              "is_member": true,
              "role_name": ""
          }],
          "targets": [{
              "id": 1,
              "endpoint": "https://10.117.4.151",
              "name": "target_01",
              "username": "admin",
              "password": "",
              "insecure": false,
              "type": 0
          }],
          "name": "sync_01",
          "description": "",
          "filters": null,
          "trigger": {"kind": "Manual", "schedule_param": null},
          "error_job_count": 2,
          "replicate_deletion": false,
          "replicate_existing_image_now": false,
      },
      {
          "id": 2,
          "projects": [{
              "project_id": 33,
              "owner_id": 1,
              "name": "aeas",
              "deleted": 0,
              "togglable": false,
              "current_user_role_id": 0,
              "repo_count": 0,
              "metadata": {
                  "public": false,
                  "enable_content_trust": "",
                  "prevent_vul": "",
                  "severity": "",
                  "auto_scan": ""},
              "owner_name": "",
              "creation_time": null,
              "update_time": null,
              "has_project_admin_role": true,
              "is_member": true,
              "role_name": ""
          }],
          "targets": [{
              "id": 1,
              "endpoint": "https://10.117.4.151",
              "name": "target_01",
              "username": "admin",
              "password": "",
              "insecure": false,
              "type": 0
          }],
          "name": "sync_02",
          "description": "",
          "filters": null,
          "trigger": {"kind": "Manual", "schedule_param": null},
          "error_job_count": 2,
          "replicate_deletion": false,
          "replicate_existing_image_now": false,
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

    let mockEndpoints: Endpoint[] = [
        {
            "id": 1,
            "endpoint": "https://10.117.4.151",
            "name": "target_01",
            "username": "admin",
            "password": "",
            "insecure": false,
            "type": 0
        },
        {
            "id": 2,
            "endpoint": "https://10.117.5.142",
            "name": "target_02",
            "username": "AAA",
            "password": "",
            "insecure": false,
            "type": 0
        },
    ];

    // let mockProjects: Project[] = [
    //     { "project_id": 1,
    //         "owner_id": 0,
    //         "name": 'project_01',
    //         "creation_time": '',
    //         "deleted": 0,
    //         "owner_name": '',
    //         "togglable": false,
    //         "update_time": '',
    //         "current_user_role_id": 0,
    //         "repo_count": 0,
    //         "has_project_admin_role": false,
    //         "is_member": false,
    //         "role_name": '',
    //         "metadata": {
    //             "public": '',
    //             "enable_content_trust": '',
    //             "prevent_vul": '',
    //             "severity": '',
    //             "auto_scan": '',
    //         }
    //     }];

  let mockJob: ReplicationJob = {
    metadata: {xTotalCount: 3},
    data: mockJobs
  };

  let fixture: ComponentFixture<ReplicationComponent>;
  let fixtureCreate: ComponentFixture<CreateEditRuleComponent>;
  let comp: ReplicationComponent;
  let compCreate: CreateEditRuleComponent;

  let replicationService: ReplicationService;
  let endpointService: EndpointService;

  let spyRules: jasmine.Spy;
  let spyJobs: jasmine.Spy;
  let spyEndpoint: jasmine.Spy;

  let deGrids: DebugElement[];
  let deRules: DebugElement;
  let deJobs: DebugElement;

  let elRule: HTMLElement;
  let elJob: HTMLElement;

  let config: IServiceConfig = {
    replicationRuleEndpoint: '/api/policies/replication/testing',
    replicationJobEndpoint: '/api/jobs/replication/testing'
  };

  beforeEach(async(() => {
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
        { provide: OperationService }
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

    spyRules = spyOn(replicationService, 'getReplicationRules').and.returnValues(Promise.resolve(mockRules));
    spyJobs = spyOn(replicationService, 'getJobs').and.returnValues(Promise.resolve(mockJob));

    spyEndpoint = spyOn(endpointService, 'getEndpoints').and.returnValues(Promise.resolve(mockEndpoints));

    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      deGrids = fixture.debugElement.queryAll(del => del.classes['datagrid']);
      fixture.detectChanges();
      expect(deGrids).toBeTruthy();
      expect(deGrids.length).toEqual(2);
    });
  });


  it('Should load replication rules', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      deRules = deGrids[0].query(By.css('datagrid-cell'));
      expect(deRules).toBeTruthy();
      fixture.detectChanges();
      elRule = deRules.nativeElement;
      expect(elRule).toBeTruthy();
      expect(elRule.textContent).toEqual('sync_01');
    });
  }));

  it('Should load replication jobs', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
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

  it('Should filter replication rules by keywords', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      comp.doSearchRules('sync_01');
      fixture.detectChanges();
      let el: HTMLElement = deRules.nativeElement;
      fixture.detectChanges();
      expect(el.textContent.trim()).toEqual('sync_01');
    });
  }));

  it('Should filter replication jobs by keywords', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      comp.doSearchJobs('nginx');
      fixture.detectChanges();
      let el: HTMLElement = deJobs.nativeElement;
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('library/nginx');
    });
  }));

  it('Should filter replication jobs by status', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      comp.doFilterJobStatus('finished');
      let el: HTMLElement = deJobs.nativeElement;
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('library/mysql');
    });
  }));

  it('Should filter replication jobs by date range', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();
      comp.doJobSearchByStartTime('2017-05-01');
      comp.doJobSearchByEndTime('2015-05-25');
      let el: HTMLElement = deJobs.nativeElement;
      fixture.detectChanges();
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('library/nginx');
    });
  }));
});
