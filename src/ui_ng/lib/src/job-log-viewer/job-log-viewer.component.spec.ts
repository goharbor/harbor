import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { DebugElement } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { ReplicationService, ReplicationDefaultService } from '../service/index';

import { JobLogViewerComponent } from './job-log-viewer.component';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { ErrorHandler } from '../error-handler/index';
import { SharedModule } from '../shared/shared.module';

describe('JobLogViewerComponent (inline template)', () => {
  let component: JobLogViewerComponent;
  let fixture: ComponentFixture<JobLogViewerComponent>;
  let serviceConfig: IServiceConfig;
  let replicationService: ReplicationService;
  let spy: jasmine.Spy;
  let testConfig: IServiceConfig = {
    replicationJobEndpoint: "/api/jobs/replication/testing"
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        BrowserAnimationsModule
      ],
      declarations: [JobLogViewerComponent],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: testConfig },
        { provide: ReplicationService, useClass: ReplicationDefaultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(JobLogViewerComponent);
    component = fixture.componentInstance;

    serviceConfig = TestBed.get(SERVICE_CONFIG);
    replicationService = fixture.debugElement.injector.get(ReplicationService);
    spy = spyOn(replicationService, 'getJobLog')
      .and.returnValues(Promise.resolve("job log text"));
    fixture.detectChanges();
  });

  it('should be created', () => {
    fixture.detectChanges();

    expect(component).toBeTruthy();
    expect(serviceConfig).toBeTruthy();
    expect(serviceConfig.replicationJobEndpoint).toEqual("/api/jobs/replication/testing");

    component.open(16);
  });

});
