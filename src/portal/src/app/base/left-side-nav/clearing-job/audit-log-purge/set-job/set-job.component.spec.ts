import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { CronScheduleComponent } from '../../../../../shared/components/cron-schedule';
import { CronTooltipComponent } from '../../../../../shared/components/cron-schedule';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { SetJobComponent } from './set-job.component';
import { PurgeService } from 'ng-swagger-gen/services/purge.service';
import { NO_ERRORS_SCHEMA } from '@angular/core';

describe('GcComponent', () => {
    let component: SetJobComponent;
    let fixture: ComponentFixture<SetJobComponent>;
    let purgeService: PurgeService;
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
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [
                SetJobComponent,
                CronScheduleComponent,
                CronTooltipComponent,
            ],
            providers: [{ provide: ErrorHandler, useValue: fakedErrorHandler }],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(SetJobComponent);
        component = fixture.componentInstance;

        purgeService = fixture.debugElement.injector.get(PurgeService);
        spySchedule = spyOn(purgeService, 'getPurgeSchedule').and.returnValues(
            of(mockSchedule as any)
        );
        spyGcNow = spyOn(purgeService, 'createPurgeSchedule').and.returnValues(
            of(null)
        );
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
