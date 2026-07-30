// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import {
    ComponentFixture,
    TestBed,
    fakeAsync,
    tick,
} from '@angular/core/testing';
import { AuditLogComponent } from './audit-log.component';
import { ErrorHandler } from '../../../shared/units/error-handler';
import { FilterComponent } from '../../../shared/components/filter/filter.component';
import { of } from 'rxjs';
import { AuditLogExt } from '../../../../../ng-swagger-gen/models/audit-log-ext';
import { AuditlogService } from '../../../../../ng-swagger-gen/services/auditlog.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../shared/shared.module';

describe('AuditLogComponent (inline template)', () => {
    let component: AuditLogComponent;
    let fixture: ComponentFixture<AuditLogComponent>;
    let auditlogService: AuditlogService;
    const fakedErrorHandler = {
        error() {
            return undefined;
        },
    };
    const mockedAuditLogs: AuditLogExt[] = [];
    for (let i = 0; i < 18; i++) {
        let item: AuditLogExt = {
            id: 23 + i,
            resource: 'myproject/demo' + i,
            resource_type: 'N/A',
            operation: 'create',
            op_time: '2017-04-11T10:26:22Z',
            username: 'user91' + i,
        };
        mockedAuditLogs.push(item);
    }
    const fakedAuditlogExtsService = {
        listAuditLogExtsResponse(
            params: AuditlogService.ListAuditLogExtsParams
        ) {
            if (params && params.q) {
                if (params.q.indexOf('demo0') !== -1) {
                    return of(
                        new HttpResponse({
                            body: mockedAuditLogs.slice(0, 1),
                            headers: new HttpHeaders({
                                'x-total-count': '18',
                            }),
                        })
                    ).pipe(delay(0));
                }
                return of(
                    new HttpResponse({
                        body: mockedAuditLogs,
                        headers: new HttpHeaders({
                            'x-total-count': '18',
                        }),
                    })
                ).pipe(delay(0));
            } else {
                if (params.page === 1) {
                    return of(
                        new HttpResponse({
                            body: mockedAuditLogs.slice(0, 15),
                            headers: new HttpHeaders({
                                'x-total-count': '18',
                            }),
                        })
                    ).pipe(delay(0));
                } else {
                    return of(
                        new HttpResponse({
                            body: mockedAuditLogs.slice(15),
                            headers: new HttpHeaders({
                                'x-total-count': '18',
                            }),
                        })
                    ).pipe(delay(0));
                }
            }
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [FilterComponent, AuditLogComponent],
            providers: [
                { provide: ErrorHandler, useValue: fakedErrorHandler },
                {
                    provide: AuditlogService,
                    useValue: fakedAuditlogExtsService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AuditLogComponent);
        component = fixture.componentInstance;
        component.pageSize = 15;
        auditlogService = fixture.debugElement.injector.get(AuditlogService);
        fixture.detectChanges();
    });

    it('should be created', () => {
        expect(component).toBeTruthy();
    });
    it('should get data from AccessLogService', async () => {
        expect(auditlogService).toBeTruthy();
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.recentLogs).toBeTruthy();
        expect(component.recentLogs.length).toEqual(15);
    });

    it('should render data to view', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.recentLogs.length).toBeGreaterThan(0);
        expect(component.recentLogs[0].username).toEqual('user910');
    });
    it('should support pagination', async () => {
        fixture.autoDetectChanges(true);
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.totalCount).toBe(18);
        const el: HTMLButtonElement =
            fixture.nativeElement.querySelector('.pagination-next');
        expect(el).toBeTruthy();
        el.click();
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.currentPage).toEqual(2);
        expect(component.recentLogs.length).toEqual(3);
    });

    it('should support filtering list by keywords', fakeAsync(() => {
        fixture.detectChanges();
        tick();
        fixture.detectChanges();
        tick();
        component.doFilter('demo0');
        fixture.detectChanges();
        tick();
        fixture.detectChanges();
        tick();
        expect(component.recentLogs).toBeTruthy();
        expect(component.recentLogs.length).toEqual(1);
    }));

    it('should support refreshing', async () => {
        fixture.autoDetectChanges(true);
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.totalCount).toBe(18);
        const el: HTMLButtonElement =
            fixture.nativeElement.querySelector('.pagination-next');
        expect(el).toBeTruthy();
        el.click();
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.recentLogs).toBeTruthy();
        expect(component.recentLogs.length).toEqual(3);
        const refreshEl: HTMLElement =
            fixture.nativeElement.querySelector('.refresh-btn');
        expect(refreshEl).toBeTruthy('Not found refresh button');
        refreshEl.click();
        fixture.detectChanges();
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.recentLogs).toBeTruthy();
        expect(component.recentLogs.length).toEqual(15);
    });
});
