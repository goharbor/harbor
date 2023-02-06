import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WorkerListComponent } from './worker-list.component';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { JobserviceService } from '../../../../../../ng-swagger-gen/services/jobservice.service';
import { Worker, WorkerPool } from 'ng-swagger-gen/models';
import { ScheduleListResponse } from '../job-service-dashboard.interface';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';

describe('WorkerListComponent', () => {
    let component: WorkerListComponent;
    let fixture: ComponentFixture<WorkerListComponent>;

    const mockedWorkers: Worker[] = [
        { id: '1', job_id: '1', job_name: 'test1', pool_id: '1' },
        { id: '2', job_id: '2', job_name: 'test2', pool_id: '1' },
    ];

    const mockedPools: WorkerPool[] = [
        { pid: 1, concurrency: 10, worker_pool_id: '1' },
    ];

    const fakedJobserviceService = {
        getWorkerPools() {
            return of(mockedPools).pipe(delay(0));
        },
    };
    const fakedJobServiceDashboardSharedDataService = {
        _allWorkers: mockedWorkers,
        getAllWorkers(): ScheduleListResponse {
            return this._allWorkers;
        },
        retrieveAllWorkers() {
            return of([]);
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [WorkerListComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: JobserviceService,
                    useValue: fakedJobserviceService,
                },
                {
                    provide: JobServiceDashboardSharedDataService,
                    useValue: fakedJobServiceDashboardSharedDataService,
                },
            ],
        }).compileComponents();
        fixture = TestBed.createComponent(WorkerListComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render worker list', async () => {
        await fixture.whenStable();
        component.loadingPools = false;
        component.loadingWorkers = false;
        component.selectedPool = mockedPools[0];
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(3); // 1 + 2
    });
});
