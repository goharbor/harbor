import { ComponentFixture, TestBed } from '@angular/core/testing';
import { CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA } from '@angular/core';
import { ReplicationExecution } from '../../../../../../../ng-swagger-gen/models/replication-execution';
import { ReplicationTasksComponent } from './replication-tasks.component';
import { ActivatedRoute } from '@angular/router';
import { ReplicationService } from '../../../../../../../ng-swagger-gen/services';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { of, Subscription } from 'rxjs';
import { delay } from 'rxjs/operators';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { ReplicationTask } from '../../../../../../../ng-swagger-gen/models/replication-task';
import { SharedTestingModule } from '../../../../../shared/shared.module';

describe('ReplicationTasksComponent', () => {
    const mockJob: ReplicationExecution = {
        id: 1,
        status: 'Failed',
        policy_id: 1,
        trigger: 'Manual',
        total: 0,
        failed: 1,
        succeed: 0,
        in_progress: 0,
        stopped: 0,
    };
    const mockTask: ReplicationTask = {
        dst_resource: 'library/lightstreamer [1 item(s) in total]',
        end_time: '2020-12-21T05:56:04.000Z',
        execution_id: 15,
        id: 30,
        job_id: '8f45cd0c512ba3d8f23ee3fa',
        operation: 'copy',
        resource_type: 'image',
        src_resource: 'library/lightstreamer [1 item(s) in total]',
        start_time: '2020-12-21T05:56:03.000Z',
        status: 'Failed',
    };
    let fixture: ComponentFixture<ReplicationTasksComponent>;
    let comp: ReplicationTasksComponent;
    const fakedErrorHandler = {
        error() {},
    };
    const fakedActivatedRoute = {
        snapshot: {
            params: {
                id: 1,
            },
            data: {
                replicationTasksRoutingResolver: mockJob,
            },
        },
    };
    const fakedReplicationService = {
        listReplicationTasksResponse() {
            return of(
                new HttpResponse({
                    body: [mockTask],
                    headers: new HttpHeaders({
                        'x-total-count': '1',
                    }),
                })
            ).pipe(delay(0));
        },
        getReplicationExecution() {
            return of(mockJob).pipe(delay(0));
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [ReplicationTasksComponent],
            providers: [
                { provide: ErrorHandler, useValue: fakedErrorHandler },
                {
                    provide: ReplicationService,
                    useValue: fakedReplicationService,
                },
                { provide: ActivatedRoute, useValue: fakedActivatedRoute },
            ],
        }).compileComponents();
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(ReplicationTasksComponent);
        comp = fixture.componentInstance;
        comp.timerDelay = new Subscription();
        comp.getExecutionDetail();
        fixture.detectChanges();
    });
    afterEach(() => {
        if (comp.timerDelay) {
            comp.timerDelay.unsubscribe();
            comp.timerDelay = null;
        }
    });
    it('should be created', () => {
        expect(comp).toBeTruthy();
    });
    it('job status should be failed', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        const span: HTMLSpanElement = fixture.nativeElement.querySelector(
            '.status-failed>span'
        );
        expect(span.innerText).toEqual('REPLICATION.FAILURE');
    });
    it('should render task list', async () => {
        fixture.autoDetectChanges();
        await fixture.whenStable();
        const row = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(row.length).toEqual(1);
    });
});
