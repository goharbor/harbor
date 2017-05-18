import { ComponentFixture, TestBed, async } from '@angular/core/testing'; 
import { By } from '@angular/platform-browser';
import { NoopAnimationsModule } from "@angular/platform-browser/animations";

import { DebugElement } from '@angular/core';

import { SharedModule } from '../shared/shared.module';

import { ListReplicationJobComponent } from '../list-replication-job/list-replication-job.component';
import { ReplicationJob } from '../service/interface';

import { ErrorHandler } from '../error-handler/error-handler';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { ReplicationService, ReplicationDefaultService } from '../service/replication.service';


describe('ListReplicationJobComponent (inline template)', ()=>{
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

  let fixture: ComponentFixture<ListReplicationJobComponent>;
  
  let comp: ListReplicationJobComponent;
  
  let replicationService: ReplicationService;
   
  let spyJobs: jasmine.Spy;
  
  let config: IServiceConfig = {
    replicationJobEndpoint: '/api/policies/replication/testing'
  };

  beforeEach(async(()=>{
    TestBed.configureTestingModule({
      imports: [ 
        SharedModule,
        NoopAnimationsModule
      ],
      declarations: [
        ListReplicationJobComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ReplicationService, useClass: ReplicationDefaultService }
      ]
    });
  }));

  beforeEach(()=>{
    fixture = TestBed.createComponent(ListReplicationJobComponent);
    comp = fixture.componentInstance;
    replicationService = fixture.debugElement.injector.get(ReplicationService);
    spyJobs = spyOn(replicationService, 'getJobs').and.returnValues(Promise.resolve(mockJobs));
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
      expect(el.textContent.trim()).toEqual('library/busybox');
    });
  }));

});