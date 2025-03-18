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
import { TagRetentionTasksComponent } from './tag-retention-tasks.component';
import { SharedTestingModule } from '../../../../../../shared/shared.module';
import { TagRetentionService } from '../../tag-retention.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../../../../ng-swagger-gen/models/registry';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { TIMEOUT } from '../../retention';
import { RetentionService } from '../../../../../../../../ng-swagger-gen/services/retention.service';

describe('TagRetentionTasksComponent', () => {
    let component: TagRetentionTasksComponent;
    let fixture: ComponentFixture<TagRetentionTasksComponent>;
    const mockedRunningTasks = [
        {
            end_time: '2021-04-26T04:32:21Z',
            execution_id: 57,
            id: 55,
            job_id: '85f5d7edab421456aae0159f',
            repository: 'hello-world',
            retained: 1,
            start_time: '2021-04-26T04:32:18Z',
            status: 'Running',
            status_code: 3,
            total: 1,
        },
    ];
    const mockedSuccessTasks = [
        {
            end_time: '2021-04-26T04:32:21Z',
            execution_id: 57,
            id: 55,
            job_id: '85f5d7edab421456aae0159f',
            repository: 'hello-world',
            retained: 1,
            start_time: '2021-04-26T04:32:18Z',
            status: 'Success',
            status_code: 3,
            total: 1,
        },
    ];
    const mockedTagRetentionService = {
        seeLog() {},
    };

    const mockRetentionService = {
        count: 0,
        listRetentionTasksResponse() {
            if (this.count === 0) {
                this.count += 1;
                const response: HttpResponse<Array<Registry>> =
                    new HttpResponse<Array<Registry>>({
                        headers: new HttpHeaders({
                            'x-total-count':
                                mockedRunningTasks.length.toString(),
                        }),
                        body: mockedRunningTasks,
                    });
                return of(response).pipe(delay(0));
            } else {
                this.count += 1;
                const response: HttpResponse<Array<Registry>> =
                    new HttpResponse<Array<Registry>>({
                        headers: new HttpHeaders({
                            'x-total-count':
                                mockedSuccessTasks.length.toString(),
                        }),
                        body: mockedSuccessTasks,
                    });
                return of(response).pipe(delay(0));
            }
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [TagRetentionTasksComponent],
            providers: [
                {
                    provide: TagRetentionService,
                    useValue: mockedTagRetentionService,
                },
                {
                    provide: RetentionService,
                    useValue: mockRetentionService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TagRetentionTasksComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should retry getting tasks', fakeAsync(() => {
        tick(TIMEOUT);
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(component.tasks[0].status).toEqual('Success');
        });
    }));
});
