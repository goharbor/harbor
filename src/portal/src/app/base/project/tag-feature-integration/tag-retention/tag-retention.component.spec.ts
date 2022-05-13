import {
    ComponentFixture,
    TestBed,
    fakeAsync,
    tick,
} from '@angular/core/testing';
import { TagRetentionComponent } from './tag-retention.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { of } from 'rxjs';
import { ActivatedRoute } from '@angular/router';
import { AddRuleComponent } from './add-rule/add-rule.component';
import { TagRetentionService } from './tag-retention.service';
import { RuleMetadate, Retention, TIMEOUT } from './retention';
import { delay } from 'rxjs/operators';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../../ng-swagger-gen/models/registry';

describe('TagRetentionComponent', () => {
    const mockedRunningExecutions = [
        {
            dry_run: true,
            end_time: '2021-04-26T04:32:21Z',
            id: 57,
            policy_id: 1,
            start_time: '2021-04-26T04:32:18.032419Z',
            status: 'Running',
            trigger: 'MANUAL',
        },
    ];
    const mockedSuccessExecutions = [
        {
            dry_run: true,
            end_time: '2021-04-26T04:32:21Z',
            id: 57,
            policy_id: 1,
            start_time: '2021-04-26T04:32:18.032419Z',
            status: 'Success',
            trigger: 'MANUAL',
        },
    ];
    let component: TagRetentionComponent;
    let fixture: ComponentFixture<TagRetentionComponent>;
    const mockTagRetentionService = {
        createRetention: () => of(null).pipe(delay(0)),
        updateRetention: () => of(null).pipe(delay(0)),
        runNowTrigger: () => of(null).pipe(delay(0)),
        whatIfRunTrigger: () => of(null).pipe(delay(0)),
        AbortRun: () => of(null).pipe(delay(0)),
        seeLog: () => of(null).pipe(delay(0)),
        getExecutionHistory: () =>
            of({
                body: [],
            }).pipe(delay(0)),
        count: 0,
        getRunNowList() {
            if (this.count === 0) {
                this.count += 1;
                const response: HttpResponse<Array<Registry>> =
                    new HttpResponse<Array<Registry>>({
                        headers: new HttpHeaders({
                            'x-total-count':
                                mockedRunningExecutions.length.toString(),
                        }),
                        body: mockedRunningExecutions,
                    });
                return of(response).pipe(delay(0));
            } else {
                this.count += 1;
                const response: HttpResponse<Array<Registry>> =
                    new HttpResponse<Array<Registry>>({
                        headers: new HttpHeaders({
                            'x-total-count':
                                mockedSuccessExecutions.length.toString(),
                        }),
                        body: mockedSuccessExecutions,
                    });
                return of(response).pipe(delay(0));
            }
        },
        getProjectInfo: () =>
            of({
                metadata: {
                    retention_id: 1,
                },
            }).pipe(delay(0)),
        getRetentionMetadata: () => of(new RuleMetadate()).pipe(delay(0)),
        getRetention: () => of(new Retention()).pipe(delay(0)),
    };
    const mockActivatedRoute = {
        snapshot: {
            parent: {
                parent: {
                    parent: {
                        params: { id: 1 },
                        data: {
                            projectResolver: {
                                metadata: {
                                    retention_id: 1,
                                },
                            },
                        },
                    },
                },
            },
        },
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            imports: [SharedTestingModule],
            declarations: [TagRetentionComponent, AddRuleComponent],
            providers: [
                {
                    provide: TagRetentionService,
                    useValue: mockTagRetentionService,
                },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(TagRetentionComponent);
        component = fixture.componentInstance;
        component.loadingRule = false;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should retry getting executions', fakeAsync(() => {
        tick(TIMEOUT);
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(component.executionList[0].status).toEqual('Success');
        });
    }));
});
