import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ProjectAuditLogComponent } from './audit-log.component';
import { MessageHandlerService } from '../../../shared/services/message-handler.service';
import { ActivatedRoute, Router } from '@angular/router';
import { of } from 'rxjs';
import { CUSTOM_ELEMENTS_SCHEMA, DebugElement, LOCALE_ID } from '@angular/core';
import { delay } from 'rxjs/operators';
import { AuditLogExt } from '../../../../../ng-swagger-gen/models/audit-log-ext';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { ProjectService } from '../../../../../ng-swagger-gen/services/project.service';
import { click } from '../../../shared/units/utils';
import { SharedTestingModule } from '../../../shared/shared.module';
import { registerLocaleData } from '@angular/common';
import locale_en from '@angular/common/locales/en';
import { DatePickerComponent } from '../../../shared/components/datetime-picker/datetime-picker.component';

describe('ProjectAuditLogComponent', () => {
    let component: ProjectAuditLogComponent;
    let fixture: ComponentFixture<ProjectAuditLogComponent>;
    const mockMessageHandlerService = {
        handleError: () => {},
    };
    const mockActivatedRoute = {
        parent: {
            parent: {
                parent: {
                    snapshot: {
                        data: null,
                    },
                },
            },
        },
        snapshot: {
            data: null,
        },
        data: of({
            auditLogResolver: '',
        }).pipe(delay(0)),
    };
    const mockRouter = null;
    const mockedAuditLogExts: AuditLogExt[] = [];
    for (let i = 0; i < 18; i++) {
        let item: AuditLogExt = {
            id: 234 + i,
            resource: 'myProject/Demo' + i,
            resource_type: 'N/A',
            operation: 'create',
            op_time: '2017-04-11T10:26:22Z',
            username: 'user91' + i,
        };
        mockedAuditLogExts.push(item);
    }
    const fakedAuditlogExtService = {
        getLogExtsResponse(params: ProjectService.GetLogsParams) {
            if (params.q && params.q.indexOf('Demo0') !== -1) {
                return of(
                    new HttpResponse({
                        body: mockedAuditLogExts.slice(0, 1),
                        headers: new HttpHeaders({
                            'x-total-count': '18',
                        }),
                    })
                ).pipe(delay(0));
            }
            if (params.page <= 1) {
                return of(
                    new HttpResponse({
                        body: mockedAuditLogExts.slice(0, 15),
                        headers: new HttpHeaders({
                            'x-total-count': '18',
                        }),
                    })
                ).pipe(delay(0));
            } else {
                return of(
                    new HttpResponse({
                        body: mockedAuditLogExts.slice(15),
                        headers: new HttpHeaders({
                            'x-total-count': '18',
                        }),
                    })
                ).pipe(delay(0));
            }
        },
    };
    registerLocaleData(locale_en, 'en-us');
    beforeEach(async () => {
        TestBed.overrideComponent(DatePickerComponent, {
            set: {
                providers: [
                    {
                        provide: LOCALE_ID,
                        useValue: 'en-us',
                    },
                ],
            },
        });
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [ProjectAuditLogComponent],
            providers: [
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: Router, useValue: mockRouter },
                { provide: ProjectService, useValue: fakedAuditlogExtService },
                {
                    provide: MessageHandlerService,
                    useValue: mockMessageHandlerService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ProjectAuditLogComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get data from AccessLogService', () => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            // wait for async getRecentLogs
            fixture.detectChanges();
            expect(component.auditLogs).toBeTruthy();
            expect(component.auditLogs.length).toEqual(15);
        });
    });

    it('should render data to view', () => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();

            let de: DebugElement = fixture.debugElement.query(
                del => del.classes['datagrid-cell']
            );
            expect(de).toBeTruthy();
            let el: HTMLElement = de.nativeElement;
            expect(el).toBeTruthy();
            expect(el.textContent.trim()).toEqual('user910');
        });
    });
    it('should support pagination', async () => {
        fixture.autoDetectChanges(true);
        await fixture.whenStable();
        let el: HTMLButtonElement =
            fixture.nativeElement.querySelector('.pagination-next');
        expect(el).toBeTruthy();
        el.click();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(component.currentPage).toEqual(2);
        expect(component.auditLogs.length).toEqual(3);
    });

    it('should support filtering list by keywords', () => {
        fixture.detectChanges();
        let el: HTMLElement =
            fixture.nativeElement.querySelector('.search-btn');
        expect(el).toBeTruthy('Not found search icon');
        click(el);
        fixture.detectChanges();
        let el2: HTMLInputElement =
            fixture.nativeElement.querySelector('input');
        expect(el2).toBeTruthy('Not found input');
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            component.doSearchAuditLogs('Demo0');
            fixture.detectChanges();
            fixture.whenStable().then(() => {
                fixture.detectChanges();
                expect(component.auditLogs).toBeTruthy();
                expect(component.auditLogs.length).toEqual(1);
            });
        });
    });
});
