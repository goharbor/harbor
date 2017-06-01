import { async, ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpModule } from '@angular/http';
import { DebugElement } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { AccessLog, RequestQueryParams } from '../service/index';

import { RecentLogComponent } from './recent-log.component';
import { AccessLogService, AccessLogDefaultService } from '../service/access-log.service';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { ErrorHandler } from '../error-handler/index';
import { SharedModule } from '../shared/shared.module';
import { FilterComponent } from '../filter/filter.component';

describe('RecentLogComponent (inline template)', () => {
  let component: RecentLogComponent;
  let fixture: ComponentFixture<RecentLogComponent>;
  let serviceConfig: IServiceConfig;
  let logService: AccessLogService;
  let spy: jasmine.Spy;
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
  let testConfig: IServiceConfig = {
    logBaseEndpoint: "/api/logs/testing"
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [FilterComponent, RecentLogComponent],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: testConfig },
        { provide: AccessLogService, useClass: AccessLogDefaultService }
      ]
    });

  }));

  beforeEach(()=>{
    fixture = TestBed.createComponent(RecentLogComponent);
    component = fixture.componentInstance;
    serviceConfig = TestBed.get(SERVICE_CONFIG);
    logService = fixture.debugElement.injector.get(AccessLogService);

    spy = spyOn(logService, 'getRecentLogs')
      .and.returnValue(Promise.resolve(mockData));

    fixture.detectChanges();
  });

  it('should be created', () => {
    expect(component).toBeTruthy();
  });

  it('should inject the SERVICE_CONFIG', () => {
    expect(serviceConfig).toBeTruthy();
    expect(serviceConfig.logBaseEndpoint).toEqual("/api/logs/testing");
  });

  it('should inject and call the AccessLogService', () => {
    expect(logService).toBeTruthy();
    expect(spy.calls.any()).toBe(true, 'getRecentLogs called');
  });

  it('should get data from AccessLogService', async(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => { // wait for async getRecentLogs
      fixture.detectChanges();
      expect(component.recentLogs).toBeTruthy();
      expect(component.logsCache).toBeTruthy();
      expect(component.recentLogs.length).toEqual(2);
    });
  }));

  it('should support filtering list by keywords', async(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      component.doFilter('push');
      fixture.detectChanges();
      expect(component.recentLogs.length).toEqual(1);
      let log: AccessLog = component.recentLogs[0];
      expect(log).toBeTruthy();
      expect(log.username).toEqual('admin');
    });
  }));

  it('should support refreshing', async(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      component.doFilter('push');
      fixture.detectChanges();
      expect(component.recentLogs.length).toEqual(1);
    });

    component.refresh();
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();

      expect(component.recentLogs.length).toEqual(1);
    });
  }));

  it('should render data to view', async(() => {
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let de: DebugElement = fixture.debugElement.query(del => del.classes['datagrid-cell']);
      expect(de).toBeTruthy();
      let el: HTMLElement = de.nativeElement;
      expect(el).toBeTruthy();
      expect(el.textContent.trim()).toEqual('user91');
    });
  }));

});
