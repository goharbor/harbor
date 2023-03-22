import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TasksComponent } from './tasks.component';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { Execution } from '../../../../../../ng-swagger-gen/models/execution';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Task } from '../../../../../../ng-swagger-gen/models/task';
import { of } from 'rxjs';
import { WebhookService } from '../../../../../../ng-swagger-gen/services/webhook.service';
import { ActivatedRoute } from '@angular/router';
import { delay } from 'rxjs/operators';

describe('TasksComponent', () => {
    let component: TasksComponent;
    let fixture: ComponentFixture<TasksComponent>;

    const mockedExecutions: Execution[] = [
        {
            end_time: '2023-02-28T03:54:01Z',
            extra_attrs: {
                event_data: {
                    replication: {
                        artifact_type: 'image',
                        authentication_type: 'basic',
                        dest_resource: {
                            endpoint: 'https://nightly-oidc.harbor.io',
                            namespace: 'library',
                            registry_type: 'harbor',
                        },
                        execution_timestamp: 1677556395,
                        harbor_hostname: 'nightly-oidc.harbor.io',
                        job_status: 'Success',
                        override_mode: true,
                        policy_creator: 'admin',
                        src_resource: {
                            endpoint: 'https://hub.docker.com',
                            namespace: 'library',
                            registry_name: 'docker hub',
                            registry_type: 'docker-hub',
                        },
                        successful_artifact: [
                            {
                                name_tag: 'redis [1 item(s) in total]',
                                status: 'Success',
                                type: 'image',
                            },
                        ],
                        trigger_type: 'MANUAL',
                    },
                },
                occur_at: 1677556415,
                operator: 'MANUAL',
                type: 'REPLICATION',
            },
            id: 30,
            metrics: { success_task_count: 1, task_count: 1 },
            start_time: '2023-02-28T03:53:35Z',
            status: 'Success',
            trigger: 'EVENT',
            vendor_id: 2,
            vendor_type: 'WEBHOOK',
        },
    ];

    const tasks: Array<Task> = [
        {
            creation_time: '2023-02-28T03:53:35Z',
            end_time: '2023-02-28T03:54:01Z',
            execution_id: 30,
            extra_attrs: {
                event_data: {
                    replication: {
                        artifact_type: 'image',
                        authentication_type: 'basic',
                        dest_resource: {
                            endpoint: 'https://nightly-oidc.harbor.io',
                            namespace: 'library',
                            registry_type: 'harbor',
                        },
                        execution_timestamp: 1677556395,
                        harbor_hostname: 'nightly-oidc.harbor.io',
                        job_status: 'Success',
                        override_mode: true,
                        policy_creator: 'admin',
                        src_resource: {
                            endpoint: 'https://hub.docker.com',
                            namespace: 'library',
                            registry_name: 'docker hub',
                            registry_type: 'docker-hub',
                        },
                        successful_artifact: [
                            {
                                name_tag: 'redis [1 item(s) in total]',
                                status: 'Success',
                                type: 'image',
                            },
                        ],
                        trigger_type: 'MANUAL',
                    },
                },
                occur_at: 1677556415,
                operator: 'MANUAL',
                type: 'REPLICATION',
            },
            id: 30,
            run_count: 1,
            start_time: '2023-02-28T03:53:35Z',
            status: 'Success',
            update_time: '2023-02-28T03:54:01Z',
        },
    ];

    const mockedWebhookService = {
        ListExecutionsOfWebhookPolicy() {
            return of(mockedExecutions).pipe(delay(0));
        },
        ListTasksOfWebhookExecutionResponse() {
            return of(
                new HttpResponse<Array<Task>>({
                    headers: new HttpHeaders({
                        'x-total-count': '1',
                    }),
                    body: tasks,
                })
            ).pipe(delay(0));
        },
    };

    const mockActivatedRoute = {
        snapshot: {
            params: {
                policyId: 1,
                executionId: 1,
            },
            parent: {
                parent: {
                    params: { id: 1 },
                },
            },
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [TasksComponent],
            providers: [
                {
                    provide: WebhookService,
                    useValue: mockedWebhookService,
                },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(TasksComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render task list and no timeout', async () => {
        await fixture.whenStable();
        fixture.detectChanges();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(1);
        expect(component.timeoutForTaskList).toBeFalsy();
        expect(component.executionTimeout).toBeFalsy();
    });

    it('should show success state', async () => {
        await fixture.whenStable();
        fixture.detectChanges();
        const successState =
            fixture.nativeElement.querySelectorAll('.status-success');
        expect(successState).toBeTruthy();
        expect(component.timeoutForTaskList).toBeFalsy();
    });
});
