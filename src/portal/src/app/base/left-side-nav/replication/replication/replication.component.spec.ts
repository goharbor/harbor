import { ComponentFixture, TestBed } from '@angular/core/testing';
import {
    DebugElement,
    CUSTOM_ELEMENTS_SCHEMA,
    NO_ERRORS_SCHEMA,
} from '@angular/core';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { ConfirmationDialogComponent } from '../../../../shared/components/confirmation-dialog';
import { ReplicationComponent } from './replication.component';
import { CronScheduleComponent } from '../../../../shared/components/cron-schedule';
import { ReplicationJob, Endpoint } from '../../../../shared/services';
import { CronTooltipComponent } from '../../../../shared/components/cron-schedule';
import { ErrorHandler } from '../../../../shared/units/error-handler';
import { ReplicationJobItem } from '../../../../shared/services';
import { OperationService } from '../../../../shared/components/operation/operation.service';
import { RouterTestingModule } from '@angular/router/testing';
import { of, Subscription } from 'rxjs';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { ReplicationPolicy } from '../../../../../../ng-swagger-gen/models/replication-policy';
import { ReplicationService } from 'ng-swagger-gen/services/replication.service';

describe('Replication Component (inline template)', () => {
    let mockRules: ReplicationPolicy[] = [
        {
            id: 1,
            name: 'sync_01',
            description: '',
            filters: null,
            trigger: { type: 'Manual', trigger_settings: null },
            deletion: false,
            src_registry: { id: 3 },
            dest_namespace: '',
            enabled: true,
            override: true,
            speed: -1,
        },
        {
            id: 2,
            name: 'sync_02',
            description: '',
            filters: null,
            trigger: { type: 'Manual', trigger_settings: null },
            deletion: false,
            dest_registry: { id: 5 },
            dest_namespace: '',
            enabled: true,
            override: true,
            speed: -1,
        },
    ];

    let mockJobs: ReplicationJobItem[] = [
        {
            id: 1,
            status: 'stopped',
            policy_id: 1,
            trigger: 'Manual',
            total: 0,
            failed: 0,
            succeed: 0,
            in_progress: 0,
            stopped: 0,
        },
        {
            id: 2,
            status: 'stopped',
            policy_id: 1,
            trigger: 'Manual',
            total: 1,
            failed: 0,
            succeed: 1,
            in_progress: 0,
            stopped: 0,
        },
        {
            id: 3,
            status: 'stopped',
            policy_id: 2,
            trigger: 'Manual',
            total: 1,
            failed: 1,
            succeed: 0,
            in_progress: 0,
            stopped: 0,
        },
    ];

    let mockEndpoints: Endpoint[] = [
        {
            id: 1,
            credential: {
                access_key: 'admin',
                access_secret: '',
                type: 'basic',
            },
            description: 'test',
            insecure: false,
            name: 'target_01',
            type: 'Harbor',
            url: 'https://10.117.4.151',
        },
        {
            id: 2,
            credential: {
                access_key: 'admin',
                access_secret: '',
                type: 'basic',
            },
            description: 'test',
            insecure: false,
            name: 'target_02',
            type: 'Harbor',
            url: 'https://10.117.5.142',
        },
    ];

    let mockJob: ReplicationJob = {
        metadata: { xTotalCount: 3 },
        data: mockJobs,
    };
    let fixture: ComponentFixture<ReplicationComponent>;
    let comp: ReplicationComponent;
    let deJobs: DebugElement;
    const fakedErrorHandler = {
        error() {},
    };
    const fakedReplicationService = {
        listReplicationPoliciesResponse() {
            return of(
                new HttpResponse({
                    body: mockRules,
                    headers: new HttpHeaders({
                        'x-total-count': '2',
                    }),
                })
            ).pipe(delay(0));
        },
        getExecutions() {
            return of(mockJob).pipe(delay(0));
        },
        getEndpoints() {
            return of(mockEndpoints).pipe(delay(0));
        },
    };

    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA, NO_ERRORS_SCHEMA],
            imports: [
                SharedTestingModule,
                NoopAnimationsModule,
                RouterTestingModule,
            ],
            declarations: [
                ReplicationComponent,
                CronTooltipComponent,
                CronScheduleComponent,
                ConfirmationDialogComponent,
            ],
            providers: [
                { provide: ErrorHandler, useValue: fakedErrorHandler },
                {
                    provide: ReplicationService,
                    useValue: fakedReplicationService,
                },
                { provide: OperationService },
            ],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(ReplicationComponent);
        comp = fixture.componentInstance;
        comp.projectId = 1;
        comp.search.ruleId = 1;
        comp.withReplicationJob = true;
        comp.hiddenJobList = false;
        comp.searchSub = new Subscription();
        spyOn(comp, 'clrLoadJobs').and.returnValue(undefined);
        comp.jobs = mockJobs;
        fixture.detectChanges();
    });
    it('Should load replication jobs', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.querySelectorAll('clr-dg-row');
        expect(rows).toBeTruthy();
        expect(rows.length).toEqual(3);
    });
    it('function "getDuration" should work', () => {
        // ms level
        const item: ReplicationJobItem = {
            start_time: 1589340503637,
            end_time: 1589340503638,
            id: 3,
            status: 'stopped',
            policy_id: 2,
            trigger: 'Manual',
            total: 1,
            failed: 1,
            succeed: 0,
            in_progress: 0,
            stopped: 0,
        };
        expect(comp.getDuration(item)).toEqual('1ms');
        // sec level
        item.start_time = 1589340503637;
        item.end_time = 1589340504638;
        expect(comp.getDuration(item)).toEqual('1s');
        // min level
        item.start_time = 1589340503637;
        item.end_time = 1589340564638;
        expect(comp.getDuration(item)).toEqual('1m1s');
        // hour level
        item.start_time = 1589340503637;
        item.end_time = 1589344164638;
        expect(comp.getDuration(item)).toEqual('61m1s');
        // day level
        item.start_time = '5/8/20,11:20 AM';
        item.end_time = '5/9/20,11:24 AM';
        expect(comp.getDuration(item)).toEqual('1444m');
    });
});
