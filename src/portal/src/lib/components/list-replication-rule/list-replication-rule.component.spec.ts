import { ComponentFixture, TestBed, async } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { DebugElement } from '@angular/core';

import { SharedModule } from '../../utils/shared/shared.module';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';

import { ListReplicationRuleComponent } from './list-replication-rule.component';
import { ReplicationRule } from '../../services/interface';

import { ErrorHandler } from '../../utils/error-handler/error-handler';
import { SERVICE_CONFIG, IServiceConfig } from '../../entities/service.config';
import { ReplicationService, ReplicationDefaultService } from '../../services/replication.service';
import { OperationService } from "../operation/operation.service";
import { of } from 'rxjs';
import { CURRENT_BASE_HREF } from "../../utils/utils";

describe('ListReplicationRuleComponent (inline template)', () => {

  let mockRules: ReplicationRule[] = [
    {
        "id": 1,
        "name": "sync_01",
        "description": "",
        "filters": null,
        "trigger": {"type": "Manual", "trigger_settings": null},
        "error_job_count": 2,
        "deletion": false,
        "src_namespaces": ["name1", "name2"],
        "src_registry": {id: 3},
        "enabled": true,
        "override": true
    },
    {
          "id": 2,
          "name": "sync_02",
          "description": "",
          "filters": null,
          "trigger": {"type": "Manual", "trigger_settings": null},
          "error_job_count": 2,
          "deletion": false,
          "src_namespaces": ["name1", "name2"],
          "dest_registry": {id: 3},
          "enabled": true,
          "override": true
      },
  ];

  let fixture: ComponentFixture<ListReplicationRuleComponent>;

  let comp: ListReplicationRuleComponent;

  let replicationService: ReplicationService;

  let spyRules: jasmine.Spy;

  let config: IServiceConfig = {
    replicationRuleEndpoint: CURRENT_BASE_HREF + '/policies/replication/testing'
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
    spyRules = spyOn(replicationService, 'getReplicationRules').and.returnValues(of(mockRules));

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
