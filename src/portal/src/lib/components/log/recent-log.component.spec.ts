import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { DebugElement } from '@angular/core';
import { RecentLogComponent } from './recent-log.component';
import { SERVICE_CONFIG, IServiceConfig } from '../../entities/service.config';
import { ErrorHandler } from '../../utils/error-handler';
import { SharedModule } from '../../utils/shared/shared.module';
import { FilterComponent } from '../filter/filter.component';
import { click, CURRENT_BASE_HREF } from '../../utils/utils';
import { of } from 'rxjs';
import { AuditLog } from "../../../../ng-swagger-gen/models/audit-log";
import { AuditlogService } from "../../../../ng-swagger-gen/services/auditlog.service";
import { HttpHeaders, HttpResponse } from "@angular/common/http";
import ListAuditLogsParams = AuditlogService.ListAuditLogsParams;
import { delay } from "rxjs/operators";

describe('RecentLogComponent (inline template)', () => {
  let component: RecentLogComponent;
  let fixture: ComponentFixture<RecentLogComponent>;
  let serviceConfig: IServiceConfig;
  let auditlogService: AuditlogService;
  let testConfig: IServiceConfig = {
    logBaseEndpoint: CURRENT_BASE_HREF + "/logs/testing"
  };
  const fakedErrorHandler = {
    error() {
      return undefined;
    }
  };
  const mockedAuditLogs: AuditLog [] = [];
  for (let i = 0; i < 18; i++) {
    let item: AuditLog = {
      id: 23 + i,
      resource: "myproject/demo" + i,
      resource_type: "N/A",
      operation: "create",
      op_time: "2017-04-11T10:26:22Z",
      username: "user91" + i
    };
    mockedAuditLogs.push(item);
  }
  const fakedAuditlogService = {
    listAuditLogsResponse(params: ListAuditLogsParams) {
      if (params && params.username) {
        if (params.username === 'demo0') {
          return of(new HttpResponse({
            body: mockedAuditLogs.slice(0, 1),
            headers:  new HttpHeaders({
              "x-total-count": "18"
            })
          })).pipe(delay(0));
        }
        return of(new HttpResponse({
          body: mockedAuditLogs,
          headers:  new HttpHeaders({
            "x-total-count": "18"
          })
        })).pipe(delay(0));
      } else {
        if (params.page === 1) {
          return of(new HttpResponse({
            body: mockedAuditLogs.slice(0, 15),
            headers:  new HttpHeaders({
              "x-total-count": "18"
            })
          })).pipe(delay(0));
        } else {
          return of(new HttpResponse({
            body: mockedAuditLogs.slice(15),
            headers:  new HttpHeaders({
              "x-total-count": "18"
            })
          })).pipe(delay(0));
        }
      }
    }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [FilterComponent, RecentLogComponent],
      providers: [
        { provide: ErrorHandler, useValue: fakedErrorHandler },
        { provide: AuditlogService, useValue: fakedAuditlogService },
        { provide: SERVICE_CONFIG, useValue: testConfig },
      ]
    });

  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RecentLogComponent);
    component = fixture.componentInstance;
    serviceConfig = TestBed.get(SERVICE_CONFIG);
    auditlogService = fixture.debugElement.injector.get(AuditlogService);
    fixture.detectChanges();
  });

  it('should be created', () => {
    expect(component).toBeTruthy();
  });

  it('should inject the SERVICE_CONFIG', () => {
    expect(serviceConfig).toBeTruthy();
    expect(serviceConfig.logBaseEndpoint).toEqual(CURRENT_BASE_HREF + "/logs/testing");
  });

  it('should get data from AccessLogService', async(() => {
    expect(auditlogService).toBeTruthy();
    fixture.detectChanges();
    fixture.whenStable().then(() => { // wait for async getRecentLogs
      fixture.detectChanges();
      expect(component.recentLogs).toBeTruthy();
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
  it('should support pagination', async () => {
    fixture.autoDetectChanges(true);
    await fixture.whenStable();
    let el: HTMLButtonElement = fixture.nativeElement.querySelector('.pagination-next');
    expect(el).toBeTruthy();
    el.click();
    fixture.detectChanges();
    await fixture.whenStable();
    expect(component.currentPage).toEqual(2);
    expect(component.recentLogs.length).toEqual(3);
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
        expect(component.recentLogs.length).toEqual(3);
        let refreshEl: HTMLElement = fixture.nativeElement.querySelector(".refresh-btn");
        expect(refreshEl).toBeTruthy("Not found refresh button");
        refreshEl.click();
        fixture.detectChanges();
        fixture.whenStable().then(() => {
          fixture.detectChanges();
          expect(component.recentLogs).toBeTruthy();
          expect(component.recentLogs.length).toEqual(15);
        });

      });
    });

  }));

});
