import { ComponentFixture, TestBed } from '@angular/core/testing';
import { GcComponent } from './gc.component';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { CronScheduleComponent } from '../../../../shared/components/cron-schedule';
import { CronTooltipComponent } from '../../../../shared/components/cron-schedule';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { GcService } from '../../../../../../ng-swagger-gen/services/gc.service';
import { ScheduleType } from '../../../../shared/entities/shared.const';

describe('GcComponent', () => {
    let component: GcComponent;
    let fixture: ComponentFixture<GcComponent>;
    let gcRepoService: GcService;
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
                GcComponent,
                CronScheduleComponent,
                CronTooltipComponent,
            ],
            providers: [{ provide: ErrorHandler, useValue: fakedErrorHandler }],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(GcComponent);
        component = fixture.componentInstance;

        gcRepoService = fixture.debugElement.injector.get(GcService);
        spySchedule = spyOn(gcRepoService, 'getGCSchedule').and.returnValues(
            of(mockSchedule as any)
        );
        spyGcNow = spyOn(gcRepoService, 'createGCSchedule').and.returnValues(
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
    it('getScheduleType function should work', () => {
        expect(GcComponent.getScheduleType).toBeTruthy();
        expect(GcComponent.getScheduleType(null)).toEqual(ScheduleType.NONE);
        expect(GcComponent.getScheduleType('0 0 0 0 0 0')).toEqual(
            ScheduleType.CUSTOM
        );
        expect(GcComponent.getScheduleType('0 0 * * * *')).toEqual(
            ScheduleType.HOURLY
        );
        expect(GcComponent.getScheduleType('0 0 0 * * *')).toEqual(
            ScheduleType.DAILY
        );
        expect(GcComponent.getScheduleType('0 0 0 * * 0')).toEqual(
            ScheduleType.WEEKLY
        );
    });
});
