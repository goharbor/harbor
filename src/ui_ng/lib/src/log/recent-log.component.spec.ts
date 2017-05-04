import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpModule } from '@angular/http';
import { DebugElement } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { AccessLog, RequestQueryParams } from '../service/index';

import { RecentLogComponent } from './recent-log.component';
import { AccessLogService } from '../service/access-log.service';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { ErrorHandler } from '../error-handler/index';
import { SharedModule } from '../shared/shared.module';
import { FilterComponent } from '../filter/filter.component';

describe('RecentLogComponent', () => {
  let component: RecentLogComponent;
  let fixture: ComponentFixture<RecentLogComponent>;
  let serviceConfig: IServiceConfig;
  let logService: AccessLogService;

  beforeEach(async(() => {
    const testConfig: IServiceConfig = {
      logBaseEndpoint: "/api/logs/testing"
    };

    class MockLogService extends AccessLogService {
      public getAuditLogs(projectId: number | string, queryParams?: RequestQueryParams): Observable<AccessLog[]> | Promise<AccessLog[]> | AccessLog[] {
        return Observable.of([]);
      }

      public getRecentLogs(lines: number): Observable<AccessLog[]> | Promise<AccessLog[]> | AccessLog[] {
        let mockData: AccessLog[] = [{
          log_id: 23,
          user_id: 45,
          project_id: 11,
          repo_name: "myproject/",
          repo_tag: "N/A",
          operation: "create",
          op_time: "2017-04-11T10:26:22Z",
          username: "user91"
        }, {
          log_id: 18,
          user_id: 1,
          project_id: 5,
          repo_name: "demo2/vmware/harbor-ui",
          repo_tag: "0.6",
          operation: "push",
          op_time: "2017-03-09T02:29:59Z",
          username: "admin"
        }];
        return Observable.of(mockData);
      }
    }

    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [FilterComponent, RecentLogComponent],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: testConfig },
        { provide: AccessLogService, useClass: MockLogService }
      ]
    })
      .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RecentLogComponent);
    component = fixture.componentInstance;
    serviceConfig = TestBed.get(SERVICE_CONFIG);
    logService = TestBed.get(AccessLogService);

    fixture.detectChanges();
  });

  it('should be created', () => {
    expect(component).toBeTruthy();
  });

  it('should inject the SERVICE_CONFIG', () => {
    expect(serviceConfig).toBeTruthy();
    expect(serviceConfig.logBaseEndpoint).toEqual("/api/logs/testing");
  });

  it('should inject the AccessLogService', () => {
    expect(logService).toBeTruthy();
  });
});
