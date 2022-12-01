import { ComponentFixture, TestBed } from '@angular/core/testing';
import { PendingListComponent } from './pending-job-list.component';
import { JobQueue } from '../../../../../../ng-swagger-gen/models/job-queue';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';

describe('PendingListComponent', () => {
    let component: PendingListComponent;
    let fixture: ComponentFixture<PendingListComponent>;

    const mockedJobs: JobQueue[] = [
        {
            job_type: 'test1',
            count: 1,
        },
        {
            job_type: 'test2',
            count: 2,
        },
        {
            job_type: 'test3',
            count: 3,
        },
    ];

    const fakedJobServiceDashboardSharedDataService = {
        _jobQueues: mockedJobs,
        getJobQueues(): JobQueue[] {
            return this._jobQueues;
        },
        retrieveJobQueues() {
            return of(mockedJobs).pipe(delay(0));
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [PendingListComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: JobServiceDashboardSharedDataService,
                    useValue: fakedJobServiceDashboardSharedDataService,
                },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(PendingListComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render list', async () => {
        await fixture.whenStable();
        component.loading = false;
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(3);
    });
});
