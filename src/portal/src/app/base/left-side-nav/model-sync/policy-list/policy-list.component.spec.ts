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
import { Router } from '@angular/router';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { ModelSyncService } from '../../../../../../ng-swagger-gen/services/model-sync.service';
import { ModelSyncPolicy } from '../../../../../../ng-swagger-gen/models/model-sync-policy';
import { ModelSyncPolicyListComponent } from './policy-list.component';
import { ErrorHandler } from '../../../../shared/units/error-handler';

describe('ModelSyncPolicyListComponent', () => {
    let component: ModelSyncPolicyListComponent;
    let fixture: ComponentFixture<ModelSyncPolicyListComponent>;

    const policy: ModelSyncPolicy = {
        id: 1,
        name: 'qwen',
        enabled: true,
        registry_id: 2,
        registry: {
            id: 2,
            name: 'hf',
            type: 'huggingface',
            url: 'https://huggingface.co',
        },
        src_repository: 'Qwen/Qwen3-8B',
        dest_project_id: 1,
        dest_project_name: 'library',
        dest_repository: 'qwen/qwen3-8b',
        trigger: {
            type: 'scheduled',
            trigger_settings: { cron: '0 0 * * * *' },
        },
        last_synced_revision: 'abcdef1234567890',
    };
    const execution = {
        id: 10,
        policy_id: 1,
        status: 'Succeed',
        trigger: 'manual',
        operator: 'admin',
    };
    const mockService = {
        listModelSyncPoliciesResponse: () =>
            of(
                new HttpResponse({
                    body: [policy],
                    headers: new HttpHeaders({ 'X-Total-Count': '1' }),
                })
            ).pipe(delay(0)),
        listModelSyncExecutionsResponse: () =>
            of(
                new HttpResponse({
                    body: [execution],
                    headers: new HttpHeaders({ 'X-Total-Count': '1' }),
                })
            ).pipe(delay(0)),
        startModelSync: () => of(null),
        deleteModelSyncPolicy: () => of(null),
        updateModelSyncPolicy: () => of(null),
        stopModelSync: () => of(null),
    };
    const mockRouter = { navigate: jasmine.createSpy('navigate') };
    const mockErrorHandler = { error: () => {}, info: () => {} };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [ModelSyncPolicyListComponent],
            providers: [
                { provide: ModelSyncService, useValue: mockService },
                { provide: Router, useValue: mockRouter },
                { provide: ErrorHandler, useValue: mockErrorHandler },
            ],
        }).compileComponents();
        fixture = TestBed.createComponent(ModelSyncPolicyListComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should load policies', fakeAsync(() => {
        component.clrLoadPolicies({ page: { size: 5 } });
        tick();
        expect(component.policies.length).toEqual(1);
        expect(component.policiesTotal).toEqual(1);
    }));

    it('should render policy columns', () => {
        expect(component.destination(policy)).toEqual('library/qwen/qwen3-8b');
        expect(component.revision(policy)).toEqual('main');
        expect(component.isScheduled(policy)).toBeTrue();
        expect(component.triggerI18n(policy)).toEqual('REPLICATION.SCHEDULED');
        expect(component.truncatedDescription('a'.repeat(40)).length).toEqual(
            35
        );
        expect(component.getStatusStr('Succeed', '')).toEqual('Succeeded');
        expect(
            component.getStatusStr('Failed', 'Execution skipped: busy')
        ).toEqual('Skipped');
        expect(
            component.getDuration({
                start_time: '2026-01-01T00:00:00Z',
                end_time: '2026-01-01T00:01:05Z',
            })
        ).toEqual('1m5s');
    });

    it('should load executions when a policy is selected', fakeAsync(() => {
        component.selectPolicy(policy);
        tick();
        expect(component.executions.length).toEqual(1);
        component.selectPolicy(null);
        expect(component.executions.length).toEqual(0);
    }));

    it('should navigate to the tasks page', () => {
        component.goToTasks(execution);
        expect(mockRouter.navigate).toHaveBeenCalledWith([
            'harbor',
            'model-sync',
            'executions',
            10,
            'tasks',
        ]);
    });

    it('should only stop in progress executions', () => {
        component.selectedExecutions = [execution];
        expect(component.canStop()).toBeFalse();
        component.selectedExecutions = [{ ...execution, status: 'InProgress' }];
        expect(component.canStop()).toBeTrue();
    });
});
