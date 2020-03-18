import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { AuditLogComponent } from './audit-log.component';
import { ClarityModule } from '@clr/angular';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { ActivatedRoute, Router } from '@angular/router';
import { of } from 'rxjs';
import { CUSTOM_ELEMENTS_SCHEMA, DebugElement } from '@angular/core';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
import { RouterTestingModule } from '@angular/router/testing';
import { delay } from 'rxjs/operators';
import { HarborLibraryModule } from "../../lib/harbor-library.module";
import { AuditLog } from "../../../ng-swagger-gen/models/audit-log";
import { HttpHeaders, HttpResponse } from "@angular/common/http";
import { ProjectService } from "../../../ng-swagger-gen/services/project.service";
import { click } from "../../lib/utils/utils";
import { FormsModule } from "@angular/forms";

describe('AuditLogComponent', () => {
    let component: AuditLogComponent;
    let fixture: ComponentFixture<AuditLogComponent>;
    const mockMessageHandlerService = {
        handleError: () => {}
    };
    const mockActivatedRoute = {
        parent: {
            snapshot: {
                data: null
            }
        },
        snapshot: {
            data: null
        },
        data: of({
            auditLogResolver: ""
        }).pipe(delay(0))
    };
    const mockRouter = null;
    const mockedAuditLogs: AuditLog [] = [];
    for (let i = 0; i < 18; i++) {
        let item: AuditLog = {
            id: 234 + i,
            resource: "myProject/Demo" + i,
            resource_type: "N/A",
            operation: "create",
            op_time: "2017-04-11T10:26:22Z",
            username: "user91" + i
        };
        mockedAuditLogs.push(item);
    }
    const fakedAuditlogService = {
        getLogsResponse(params: ProjectService.GetLogsParams) {
                if (params.q && params.q.indexOf('Demo0') !== -1) {
                    return of(new HttpResponse({
                        body: mockedAuditLogs.slice(0, 1),
                        headers:  new HttpHeaders({
                            "x-total-count": "18"
                        })
                    })).pipe(delay(0));
                }
            if (params.page <= 1) {
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
    };

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            schemas: [
                CUSTOM_ELEMENTS_SCHEMA
            ],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
                HarborLibraryModule
            ],
            declarations: [AuditLogComponent],
            providers: [
                TranslateService,
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: Router, useValue: mockRouter },
                { provide: ProjectService, useValue: fakedAuditlogService },
                { provide: MessageHandlerService, useValue: mockMessageHandlerService },

            ]
        }).compileComponents();
    }));

    beforeEach(() => {
        fixture = TestBed.createComponent(AuditLogComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get data from AccessLogService', async(() => {
        fixture.detectChanges();
        fixture.whenStable().then(() => { // wait for async getRecentLogs
            fixture.detectChanges();
            expect(component.auditLogs).toBeTruthy();
            expect(component.auditLogs.length).toEqual(15);
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
        expect(component.auditLogs.length).toEqual(3);
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
            component.doSearchAuditLogs("Demo0");
            fixture.detectChanges();
            fixture.whenStable().then(() => {
                fixture.detectChanges();
                expect(component.auditLogs).toBeTruthy();
                expect(component.auditLogs.length).toEqual(1);
            });
        });
    }));
});
