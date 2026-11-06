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
import {
    ComponentFixture,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { ActivatedRoute, Router } from '@angular/router';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { ModelSyncService } from '../../../../../../ng-swagger-gen/services/model-sync.service';
import { ModelSyncExecutionTasksComponent } from './execution-tasks.component';
import { ErrorHandler } from '../../../../shared/units/error-handler';

describe('ModelSyncExecutionTasksComponent', () => {
    let component: ModelSyncExecutionTasksComponent;
    let fixture: ComponentFixture<ModelSyncExecutionTasksComponent>;

    const execution = {
        id: 10,
        policy_id: 1,
        status: 'InProgress',
        trigger: 'manual',
    };
    const task = {
        id: 11,
        execution_id: 10,
        status: 'Succeed',
        src_repository: 'Qwen/Qwen3-8B',
        dest_repository: 'library/qwen/qwen3-8b',
        revision: '71034c5d8bde858ff824298bdedc65515b97d2b9',
        digest: 'sha256:abcdef1234567890abcdef',
        tags: ['sha-71034c5d8bde', 'main'],
        files: 3,
        size: 2048,
    };
    const mockService = {
        getModelSyncExecution: () => of(execution).pipe(delay(0)),
        listModelSyncTasksResponse: () =>
            of(
                new HttpResponse({
                    body: [task],
                    headers: new HttpHeaders({ 'X-Total-Count': '1' }),
                })
            ).pipe(delay(0)),
        stopModelSync: jasmine
            .createSpy('stopModelSync')
            .and.returnValue(of(null)),
    };
    const mockRouter = { navigate: jasmine.createSpy('navigate') };
    const mockRoute = { snapshot: { params: { id: '10' } } };
    const mockErrorHandler = { error: () => {}, info: () => {} };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [ModelSyncExecutionTasksComponent],
            providers: [
                { provide: ModelSyncService, useValue: mockService },
                { provide: Router, useValue: mockRouter },
                { provide: ActivatedRoute, useValue: mockRoute },
                { provide: ErrorHandler, useValue: mockErrorHandler },
            ],
        }).compileComponents();
        fixture = TestBed.createComponent(ModelSyncExecutionTasksComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    afterEach(() => {
        component.ngOnDestroy();
    });

    it('should create and load the execution', fakeAsync(() => {
        expect(component).toBeTruthy();
        expect(component.executionId).toEqual(10);
        component.loadExecution();
        tick();
        expect(component.execution.id).toEqual(10);
    }));

    it('should load tasks', fakeAsync(() => {
        component.loadExecution();
        tick();
        component.clrLoadTasks(true, { page: { size: 15 } });
        tick();
        expect(component.tasks.length).toEqual(1);
        expect(component.totalCount).toEqual(1);
        // in progress execution schedules a refresh
        expect(component.refreshTimeout).toBeTruthy();
        component.ngOnDestroy();
    }));

    it('should format helpers', () => {
        expect(component.size(2048)).toEqual('2.00KiB');
        expect(component.viewLog(11)).toContain(
            '/model-sync/executions/10/tasks/11/log'
        );
        expect(component.getStatusStr('Succeed')).toEqual('Succeeded');
        expect(component.getStatusStr('Failed')).toEqual('Failed');
    });

    it('should stop the execution', fakeAsync(() => {
        component.loadExecution();
        tick();
        expect(component.canStop()).toBeTrue();
        component.stop();
        tick();
        expect(mockService.stopModelSync).toHaveBeenCalledWith({
            executionId: 10,
        });
        component.ngOnDestroy();
    }));

    it('should navigate back', () => {
        component.onBack();
        expect(mockRouter.navigate).toHaveBeenCalledWith([
            'harbor',
            'model-sync',
        ]);
    });
});
