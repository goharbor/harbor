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


describe('ListReplicationRuleComponent (inline template)', ()=>{

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

  let fixture: ComponentFixture<ListReplicationRuleComponent>;
  
  let comp: ListReplicationRuleComponent;
  
  let replicationService: ReplicationService;
   
  let spyRules: jasmine.Spy;
  
  let config: IServiceConfig = {
    replicationRuleEndpoint: '/api/policies/replication/testing'
  };

  beforeEach(async(()=>{
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
        { provide: ReplicationService, useClass: ReplicationDefaultService }
      ]
    });
  }));

  beforeEach(()=>{
    fixture = TestBed.createComponent(ListReplicationRuleComponent);
    comp = fixture.componentInstance;
    replicationService = fixture.debugElement.injector.get(ReplicationService);
    spyRules = spyOn(replicationService, 'getReplicationRules').and.returnValues(Promise.resolve(mockRules));
    fixture.detectChanges();
  });

  it('Should load and render data', async(()=>{
    fixture.detectChanges();
    fixture.whenStable().then(()=>{
      fixture.detectChanges();
      let de: DebugElement = fixture.debugElement.query(By.css('datagrid-cell'));
      expect(de).toBeTruthy();
      fixture.detectChanges();
      let el: HTMLElement = de.nativeElement;
      expect(el.textContent.trim()).toEqual('sync_01');
    });
  }));

});