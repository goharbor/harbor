// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { WorkerListComponent } from './worker-list.component';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { JobserviceService } from '../../../../../../ng-swagger-gen/services/jobservice.service';
import { Worker, WorkerPool } from 'ng-swagger-gen/models';
import { ScheduleListResponse } from '../job-service-dashboard.interface';
import { JobServiceDashboardSharedDataService } from '../job-service-dashboard-shared-data.service';
import {
    ConfirmationState,
    ConfirmationTargets,
} from '../../../../shared/entities/shared.const';
import { ConfirmationDialogService } from '../../../global-confirmation-dialog/confirmation-dialog.service';
import { ConfirmationAcknowledgement } from '../../../global-confirmation-dialog/confirmation-state-message';
import { MessageHandlerService } from 'src/app/shared/services/message-handler.service';
import { OperationService } from '../../../../shared/components/operation/operation.service';

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
        stopRunningJob() {
            return of(null);
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

    it('canFree should return correct value based on selected workers', () => {
        component.selected = [];
        expect(component.canFree()).toBeFalse();

        // One selected worker with job_id
        component.selected = [mockedWorkers[0]];
        expect(component.canFree()).toBeTrue();

        // One selected worker without job_id
        component.selected = [{ id: '3', pool_id: '1' }];
        expect(component.canFree()).toBeFalse();

        // Multiple selected, one without job_id
        component.selected = [mockedWorkers[0], { id: '3', pool_id: '1' }];
        expect(component.canFree()).toBeFalse();
    });

    it('freeWorker should open confirm dialog', () => {
        const dialogService = TestBed.inject(ConfirmationDialogService);
        const spy = spyOn(dialogService, 'openComfirmDialog').and.callThrough();

        component.selected = [mockedWorkers[0]];
        component.freeWorker();

        expect(spy).toHaveBeenCalled();
        const arg = spy.calls.mostRecent().args[0];
        expect(arg.title).toEqual('JOB_SERVICE_DASHBOARD.CONFIRM_FREE_WORKERS');
        expect(arg.message).toEqual(
            'JOB_SERVICE_DASHBOARD.CONFIRM_FREE_WORKERS_CONTENT'
        );
    });

    it('should execute free workers when confirmed', () => {
        const dialogService = TestBed.inject(ConfirmationDialogService);
        const jobService = TestBed.inject(JobserviceService);
        const messageHandler = TestBed.inject(MessageHandlerService);
        const operationService = TestBed.inject(OperationService);

        const stopJobSpy = spyOn(jobService, 'stopRunningJob').and.returnValue(
            of(null)
        );
        const messageInfoSpy = spyOn(messageHandler, 'info');
        const operationSpy = spyOn(operationService, 'publishInfo');

        component.selected = [mockedWorkers[0]];
        component.executeFreeWorkers();

        expect(stopJobSpy).toHaveBeenCalledWith({ jobId: '1' });
        expect(messageInfoSpy).toHaveBeenCalledWith(
            'JOB_SERVICE_DASHBOARD.FREE_WORKER_SUCCESS'
        );
        expect(operationSpy).toHaveBeenCalled();
    });

    it('should trigger executeFreeWorkers when confirm message is received', () => {
        const dialogService = TestBed.inject(ConfirmationDialogService);
        const executeSpy = spyOn(component, 'executeFreeWorkers');

        const ack = new ConfirmationAcknowledgement(
            ConfirmationState.CONFIRMED,
            mockedWorkers,
            ConfirmationTargets.FREE_SPECIFIED_WORKERS
        );
        dialogService.confirm(ack);

        expect(executeSpy).toHaveBeenCalled();
    });
});
