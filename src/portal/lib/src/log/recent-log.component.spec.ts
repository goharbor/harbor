import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { DebugElement } from '@angular/core';
import { AccessLog, AccessLogItem, RequestQueryParams } from '../service/index';

import { RecentLogComponent } from './recent-log.component';
import { AccessLogService, AccessLogDefaultService } from '../service/access-log.service';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { ErrorHandler } from '../error-handler/index';
import { SharedModule } from '../shared/shared.module';
import { FilterComponent } from '../filter/filter.component';

import { click } from '../utils';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';

describe('RecentLogComponent (inline template)', () => {
  let component: RecentLogComponent;
  let fixture: ComponentFixture<RecentLogComponent>;
  let serviceConfig: IServiceConfig;
  let logService: AccessLogService;
  let spy: jasmine.Spy;
  let mockItems: AccessLogItem[] = [];
  let mockData: AccessLog = {
    metadata: {
      xTotalCount: 18
    },
    data: []
  };
  let mockData2: AccessLog = {
    metadata: {
      xTotalCount: 1
    },
    data: []
  };
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

  beforeEach(() => {
    fixture = TestBed.createComponent(RecentLogComponent);
    component = fixture.componentInstance;
    serviceConfig = TestBed.get(SERVICE_CONFIG);
    logService = fixture.debugElement.injector.get(AccessLogService);
    // Mock data
    for (let i = 0; i < 18; i++) {
      let item: AccessLogItem = {
        log_id: 23 + i,
        user_id: 45 + i,
        project_id: 11 + i,
        repo_name: "myproject/demo" + i,
        repo_tag: "N/A",
        operation: "create",
        op_time: "2017-04-11T10:26:22Z",
        username: "user91" + i
      };
      mockItems.push(item);
    }
    mockData2.data = mockItems.slice(0, 1);
    mockData.data = mockItems;

    spy = spyOn(logService, 'getRecentLogs')
      .and.callFake(function (params: RequestQueryParams) {
        if (params && params.get('username')) {
          return of(mockData2).pipe(delay(0));
        } else {
          if (params.get('page') === '1') {
            mockData.data = mockItems.slice(0, 15);
          } else {
            mockData.data = mockItems.slice(15, 18);
          }
          return of(mockData).pipe(delay(0));
        }
      });

    fixture.detectChanges();
  });

  it('should be created', () => {
    expect(component).toBeTruthy();
  });

  it('should inject the SERVICE_CONFIG', () => {
    expect(serviceConfig).toBeTruthy();
    expect(serviceConfig.logBaseEndpoint).toEqual("/api/logs/testing");
  });

  it('should get data from AccessLogService', async(() => {
    expect(logService).toBeTruthy();
    expect(spy.calls.any()).toBe(true, 'getRecentLogs called');

    fixture.detectChanges();

    fixture.whenStable().then(() => { // wait for async getRecentLogs
      fixture.detectChanges();
      expect(component.recentLogs).toBeTruthy();
      expect(component.logsCache).toBeTruthy();
      expect(component.recentLogs.length).toEqual(15);
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
      expect(el.textContent.trim()).toEqual('user910');
    });
  }));

  // Will fail after upgrade to angular 6. todo: need to fix it.
  it('should support pagination', () => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let el: HTMLButtonElement = fixture.nativeElement.querySelector('.pagination-next');
      expect(el).toBeTruthy();
      el.click();
      jasmine.clock().tick(100);
      fixture.detectChanges();

      fixture.whenStable().then(() => {
        fixture.detectChanges();

        let els: HTMLElement[] = fixture.nativeElement.querySelectorAll('.datagrid-row');
        expect(els).toBeTruthy();
        expect(els.length).toEqual(4);
      });
    });
  });

  it('should support filtering list by keywords', async(() => {
    fixture.detectChanges();
    let el: HTMLElement = fixture.nativeElement.querySelector('.search-btn');
    expect(el).toBeTruthy("Not found search icon");
    click(el);

    fixture.detectChanges();
    let el2: HTMLInputElement = fixture.nativeElement.querySelector('input');
    expect(el2).toBeTruthy("Not found input");

    fixture.detectChanges();
    fixture.whenStable().then(() => {
      fixture.detectChanges();

      component.doFilter("demo0");

      fixture.detectChanges();
      fixture.whenStable().then(() => {
        fixture.detectChanges();
        expect(component.recentLogs).toBeTruthy();
        expect(component.logsCache).toBeTruthy();
        expect(component.recentLogs.length).toEqual(1);
      });
    });
  }));

  it('should support refreshing', async(() => {
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let el: HTMLButtonElement = fixture.nativeElement.querySelector('.pagination-next');
      expect(el).toBeTruthy();
      el.click();

      fixture.detectChanges();

      fixture.whenStable().then(() => {
        fixture.detectChanges();
        expect(component.recentLogs).toBeTruthy();
        expect(component.logsCache).toBeTruthy();
        expect(component.recentLogs.length).toEqual(3);

        let refreshEl: HTMLElement = fixture.nativeElement.querySelector(".refresh-btn");
        expect(refreshEl).toBeTruthy("Not found refresh button");
        refreshEl.click();

        fixture.detectChanges();

        fixture.whenStable().then(() => {
          fixture.detectChanges();
          expect(component.recentLogs).toBeTruthy();
          expect(component.logsCache).toBeTruthy();
          expect(component.recentLogs.length).toEqual(3);
        });

      });
    });

  }));

});
