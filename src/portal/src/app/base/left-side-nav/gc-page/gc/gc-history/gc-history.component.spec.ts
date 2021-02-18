import {
    ComponentFixture,
    ComponentFixtureAutoDetect,
    TestBed,
    waitForAsync
} from '@angular/core/testing';
import { GcRepoService } from "../gc.service";
import { of } from 'rxjs';
import { GcViewModelFactory } from "../gc.viewmodel.factory";
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { GcHistoryComponent } from './gc-history.component';
import { GcJobData } from "../gcLog";
import { SharedTestingModule } from "../../../../../shared/shared.module";

describe('GcHistoryComponent', () => {
    let component: GcHistoryComponent;
    let fixture: ComponentFixture<GcHistoryComponent>;
    const mockJobs: GcJobData[] = [
        {
            id: 1,
            job_name: 'test',
            job_kind: 'manual',
            schedule: null,
            job_status: 'pending',
            job_parameters: '{"dry_run":true}',
            job_uuid: 'abc',
            creation_time: null,
            update_time: null,
            delete: false
        },
        {
            id: 2,
            job_name: 'test',
            job_kind: 'manual',
            schedule: null,
            job_status: 'finished',
            job_parameters: '{"dry_run":true}',
            job_uuid: 'bcd',
            creation_time: null,
            update_time: null,
            delete: false
        }
    ];
    let fakeGcRepoService = {
        count: 0,
        getJobs() {
            if (this.count === 0) {
                this.count += 1;
                return of([mockJobs[0]]);
            } else {
                this.count += 1;
                return of([mockJobs[1]]);
            }
        },
        getLogLink() {
            return null;
        }
    };
    const fakeGcViewModelFactory = new GcViewModelFactory();
    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [GcHistoryComponent],
            imports: [
                SharedTestingModule,
                TranslateModule.forRoot()
            ],
            providers: [
                ErrorHandler,
                TranslateService,
                GcViewModelFactory,
                { provide: GcRepoService, useValue: fakeGcRepoService },
                { provide: GcViewModelFactory, useValue: fakeGcViewModelFactory },
                 // open auto detect
                { provide: ComponentFixtureAutoDetect, useValue: true }
            ]
        });
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(GcHistoryComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });
    afterEach(() => {
        if (component && component.timerDelay) {
            component.timerDelay.unsubscribe();
            component.timerDelay = null;
        }
    });
    it('should create', () => {
        expect(component).toBeTruthy();
    });
     it('should retry getting jobs', waitForAsync(() => {
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(component.jobs[1].status).toEqual('finished');
        });
    }));
});
