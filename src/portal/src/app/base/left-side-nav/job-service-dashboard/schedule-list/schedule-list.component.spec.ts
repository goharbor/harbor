import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ScheduleListComponent } from './schedule-list.component';
import { ScheduleTask } from '../../../../../../ng-swagger-gen/models/schedule-task';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';
import { ScheduleListResponse } from '../job-service-dashboard.interface';

describe('ScheduleListComponent', () => {
    let component: ScheduleListComponent;
    let fixture: ComponentFixture<ScheduleListComponent>;

    const mockedSchedules: ScheduleTask[] = [
        { id: 1, vendor_type: 'test1' },
        { id: 2, vendor_type: 'test2' },
    ];

    const fakedJobServiceDashboardSharedDataService = {
        _scheduleListResponse: {
            scheduleList: mockedSchedules,
        },
        getScheduleListResponse(): ScheduleListResponse {
            return this._scheduleListResponse;
        },
        retrieveScheduleListResponse() {
            return of({});
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [ScheduleListComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: JobServiceDashboardSharedDataService,
                    useValue: fakedJobServiceDashboardSharedDataService,
                },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(ScheduleListComponent);
        component = fixture.componentInstance;
        component.loadingSchedules = false;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render list', async () => {
        await fixture.whenStable();
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(2);
    });
});
