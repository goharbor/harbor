import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { CronScheduleComponent } from '../../../../../shared/components/cron-schedule';
import { CronTooltipComponent } from '../../../../../shared/components/cron-schedule';
import { delay, of } from 'rxjs';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { SetJobComponent } from './set-job.component';
import { PurgeService } from 'ng-swagger-gen/services/purge.service';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { AuditlogService } from 'ng-swagger-gen/services';
import { HttpHeaders, HttpResponse } from '@angular/common/http';

describe('GcComponent', () => {
    let component: SetJobComponent;
    let fixture: ComponentFixture<SetJobComponent>;
    let purgeService: PurgeService;
    let auditlogService: AuditlogService;
    let mockSchedule = [];
    const fakedErrorHandler = {
        error(error) {
            return error;
        },
        info(info) {
            return info;
        },
    };
    let spySchedule: jasmine.Spy;
    let spyGcNow: jasmine.Spy;
    const mockedAuditLogs = [
        {
            event_type: 'create_artifact',
        },
        {
            event_type: 'delete_artifact',
        },
        {
            event_type: 'pull_artifact',
        },
    ];
    const fakedAuditlogService = {
        listAuditLogEventTypesResponse() {
            return of(
                new HttpResponse({
                    body: mockedAuditLogs,
                    headers: new HttpHeaders({
                        'x-total-count': '18',
                    }),
                })
            ).pipe(delay(0));
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [
                SetJobComponent,
                CronScheduleComponent,
                CronTooltipComponent,
            ],
            providers: [
                { provide: ErrorHandler, useValue: fakedErrorHandler },
                { provide: AuditlogService, useValue: fakedAuditlogService },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(SetJobComponent);
        component = fixture.componentInstance;
        auditlogService = fixture.debugElement.injector.get(AuditlogService);
        purgeService = fixture.debugElement.injector.get(PurgeService);
        spySchedule = spyOn(purgeService, 'getPurgeSchedule').and.returnValues(
            of(mockSchedule as any)
        );
        spyGcNow = spyOn(purgeService, 'createPurgeSchedule').and.returnValues(
            of(null)
        );
        component.selectedEventTypes = ['create_artifact'];
        fixture.detectChanges();
    });
    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get schedule and job', () => {
        expect(spySchedule.calls.count()).toEqual(1);
    });
    it('should trigger gcNow', () => {
        const ele: HTMLButtonElement =
            fixture.nativeElement.querySelector('#gc-now');
        ele.click();
        fixture.detectChanges();
        expect(spyGcNow.calls.count()).toEqual(1);
    });
    it('should trigger dry run', () => {
        const ele: HTMLButtonElement =
            fixture.nativeElement.querySelector('#gc-dry-run');
        ele.click();
        fixture.detectChanges();
        expect(spyGcNow.calls.count()).toEqual(1);
    });
});
