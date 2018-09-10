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

describe('ListReplicationRuleComponent (inline template)', () => {

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
            "endpoint": "",
            "id": 0,
            "insecure": false,
            "name": "khans3",
            "username": "",
            "password": "",
            "type": 0,
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
              "endpoint": "",
              "id": 0,
              "insecure": false,
              "name": "khans3",
              "username": "",
              "password": "",
              "type": 0,
          }],
          "name": "sync_02",
          "description": "",
          "filters": null,
          "trigger": {"kind": "Manual", "schedule_param": null},
          "error_job_count": 2,
          "replicate_deletion": false,
          "replicate_existing_image_now": false,
      },
  ];

  let fixture: ComponentFixture<ListReplicationRuleComponent>;

  let comp: ListReplicationRuleComponent;

  let replicationService: ReplicationService;

  let spyRules: jasmine.Spy;

  let config: IServiceConfig = {
    replicationRuleEndpoint: '/api/policies/replication/testing'
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
        { provide: OperationService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(ListReplicationRuleComponent);
    comp = fixture.componentInstance;
    replicationService = fixture.debugElement.injector.get(ReplicationService);
    spyRules = spyOn(replicationService, 'getReplicationRules').and.returnValues(Promise.resolve(mockRules));
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
