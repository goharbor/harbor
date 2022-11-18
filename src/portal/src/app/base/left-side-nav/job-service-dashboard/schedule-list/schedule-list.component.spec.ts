import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ScheduleListComponent } from './schedule-list.component';
import { ScheduleTask } from '../../../../../../ng-swagger-gen/models/schedule-task';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { ScheduleService } from '../../../../../../ng-swagger-gen/services/schedule.service';

describe('ScheduleListComponent', () => {
    let component: ScheduleListComponent;
    let fixture: ComponentFixture<ScheduleListComponent>;

    const mockedSchedules: ScheduleTask[] = [
        { id: 1, vendor_type: 'test1' },
        { id: 2, vendor_type: 'test2' },
    ];

    const fakedScheduleService = {
        listSchedulesResponse() {
            const res: HttpResponse<Array<ScheduleTask>> = new HttpResponse<
                Array<ScheduleTask>
            >({
                headers: new HttpHeaders({ 'x-total-count': '2' }),
                body: mockedSchedules,
            });
            return of(res).pipe(delay(0));
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ScheduleListComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: ScheduleService,
                    useValue: fakedScheduleService,
                },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(ScheduleListComponent);
        component = fixture.componentInstance;
        component.loadingSchedules = true;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render list', async () => {
        await fixture.whenStable();
        component.loadingSchedules = false;
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(2);
    });
});
