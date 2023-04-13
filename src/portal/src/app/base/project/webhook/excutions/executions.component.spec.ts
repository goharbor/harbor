import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ExecutionsComponent } from './executions.component';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { WebhookService } from '../../../../../../ng-swagger-gen/services/webhook.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Execution } from '../../../../../../ng-swagger-gen/models/execution';
import { of } from 'rxjs';
import { WebhookPolicy } from '../../../../../../ng-swagger-gen/models/webhook-policy';
import { delay } from 'rxjs/operators';
import { ProjectWebhookService } from '../webhook.service';

describe('ExecutionsComponent', () => {
    let component: ExecutionsComponent;
    let fixture: ComponentFixture<ExecutionsComponent>;

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
        {
            end_time: '2023-02-28T03:53:40Z',
            extra_attrs: {
                event_data: {
                    repository: {
                        date_created: 1677556403,
                        name: 'redis',
                        namespace: 'library',
                        repo_full_name: 'library/redis',
                        repo_type: 'public',
                    },
                    resources: [
                        {
                            digest: 'sha256:6a59f1cbb8d28ac484176d52c473494859a512ddba3ea62a547258cf16c9b3ae',
                            resource_url:
                                'nightly-oidc.harbor.io/library/redis:latest',
                            tag: 'latest',
                        },
                    ],
                },
                occur_at: 1677556415,
                operator: 'harbor-jobservice',
                type: 'PUSH_ARTIFACT',
            },
            id: 28,
            metrics: { success_task_count: 1, task_count: 1 },
            start_time: '2023-02-28T03:53:35Z',
            status: 'Success',
            trigger: 'EVENT',
            vendor_id: 2,
            vendor_type: 'WEBHOOK',
        },
    ];

    const mockedWebhookService = {
        ListExecutionsOfWebhookPolicyResponse() {
            return of(
                new HttpResponse<Array<Execution>>({
                    headers: new HttpHeaders({
                        'x-total-count': '2',
                    }),
                    body: mockedExecutions,
                })
            ).pipe(delay(0));
        },
    };

    const mockedWebhookPolicy: WebhookPolicy = {
        id: 1,
        project_id: 1,
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ExecutionsComponent],
            providers: [
                {
                    provide: WebhookService,
                    useValue: mockedWebhookService,
                },
                ProjectWebhookService,
            ],
        }).compileComponents();

        fixture = TestBed.createComponent(ExecutionsComponent);
        component = fixture.componentInstance;
        component.selectedWebhook = mockedWebhookPolicy;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render execution list and no timeout', async () => {
        await fixture.whenStable();
        fixture.detectChanges();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows.length).toEqual(2);
        expect(component.timeout).toBeFalsy();
    });
});
