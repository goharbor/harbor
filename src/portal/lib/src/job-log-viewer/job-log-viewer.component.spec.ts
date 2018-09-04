import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
// tslint:disable-next-line:no-unused-variable
import { Observable } from 'rxjs/Observable';
import { JobLogService, JobLogDefaultService } from '../service/index';

import { JobLogViewerComponent } from './job-log-viewer.component';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { ErrorHandler } from '../error-handler/index';
import { SharedModule } from '../shared/shared.module';

describe('JobLogViewerComponent (inline template)', () => {
  let component: JobLogViewerComponent;
  let fixture: ComponentFixture<JobLogViewerComponent>;
  let serviceConfig: IServiceConfig;
  let jobLogService: JobLogDefaultService;
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
        { provide: JobLogService, useClass: JobLogDefaultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(JobLogViewerComponent);
    component = fixture.componentInstance;

    serviceConfig = TestBed.get(SERVICE_CONFIG);
    jobLogService = fixture.debugElement.injector.get(JobLogService);
    spy = spyOn(jobLogService, 'getJobLog')
      .and.returnValue(Promise.resolve("job log text"));
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
