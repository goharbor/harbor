import { ComponentFixture, TestBed } from '@angular/core/testing';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { of } from 'rxjs';
import { WorkerCardComponent } from './worker-card.component';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { Worker } from '../../../../../../ng-swagger-gen/models/worker';
import { ScheduleListResponse } from '../job-service-dashboard.interface';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';

describe('WorkerCardComponent', () => {
    let component: WorkerCardComponent;
    let fixture: ComponentFixture<WorkerCardComponent>;

    const mockedWorkers: Worker[] = [
        { id: '1', job_id: '1', job_name: 'test1' },
        { id: '2', job_id: '2', job_name: 'test2' },
    ];

    const fakedJobServiceDashboardSharedDataService = {
        _allWorkers: mockedWorkers,
        getAllWorkers(): ScheduleListResponse {
            return this._allWorkers;
        },
        retrieveAllWorkers() {
            return of(mockedWorkers);
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [NO_ERRORS_SCHEMA],
            declarations: [WorkerCardComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: JobServiceDashboardSharedDataService,
                    useValue: fakedJobServiceDashboardSharedDataService,
                },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(WorkerCardComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should init timeout', () => {
        expect(component.statusTimeout).toBeTruthy();
    });

    it('should get workers', () => {
        expect(component.busyWorkers.length).toEqual(2);
    });
});
