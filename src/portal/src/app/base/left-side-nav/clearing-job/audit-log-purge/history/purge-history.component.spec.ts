import {
    ComponentFixture,
    ComponentFixtureAutoDetect,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { of } from 'rxjs';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { CURRENT_BASE_HREF } from '../../../../../shared/units/utils';
import { delay } from 'rxjs/operators';
import { PurgeHistoryComponent } from './purge-history.component';
import { ExecHistory } from '../../../../../../../ng-swagger-gen/models/exec-history';
import { PurgeService } from 'ng-swagger-gen/services/purge.service';
import { ConfirmationDialogService } from '../../../../global-confirmation-dialog/confirmation-dialog.service';

describe('GcHistoryComponent', () => {
    let component: PurgeHistoryComponent;
    let fixture: ComponentFixture<PurgeHistoryComponent>;
    const mockJobs: ExecHistory[] = [
        {
            id: 1,
            job_name: 'test',
            job_kind: 'manual',
            schedule: null,
            job_status: 'pending',
            job_parameters: '{"dry_run":true}',
            creation_time: null,
            update_time: null,
        },
        {
            id: 2,
            job_name: 'test',
            job_kind: 'manual',
            schedule: null,
            job_status: 'finished',
            job_parameters: '{"dry_run":true}',
            creation_time: null,
            update_time: null,
        },
    ];
    const fakedPurgeService = {
        count: 0,
        getPurgeHistoryResponse() {
            if (this.count === 0) {
                this.count += 1;
                const response: HttpResponse<Array<ExecHistory>> =
                    new HttpResponse<Array<ExecHistory>>({
                        headers: new HttpHeaders({
                            'x-total-count': [mockJobs[0]].length.toString(),
                        }),
                        body: [mockJobs[0]],
                    });
                return of(response).pipe(delay(0));
            } else {
                this.count += 1;
                const response: HttpResponse<Array<ExecHistory>> =
                    new HttpResponse<Array<ExecHistory>>({
                        headers: new HttpHeaders({
                            'x-total-count': [mockJobs[1]].length.toString(),
                        }),
                        body: [mockJobs[1]],
                    });
                return of(response).pipe(delay(0));
            }
        },
    };
    const fakedConfirmationDialogService = {
        openComfirmDialog() {},
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [PurgeHistoryComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: ConfirmationDialogService,
                    useValue: fakedConfirmationDialogService,
                },
                { provide: PurgeService, useValue: fakedPurgeService },
                // open auto detect
                { provide: ComponentFixtureAutoDetect, useValue: true },
            ],
        });
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(PurgeHistoryComponent);
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
    it('should retry getting jobs', fakeAsync(() => {
        tick(10000);
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(component.jobs[0].job_status).toEqual('finished');
        });
    }));
    it('should return right log link', () => {
        expect(component.getLogLink('1')).toEqual(
            `${CURRENT_BASE_HREF}/system/purgeaudit/1/log`
        );
    });
    it('stopping purging should work', () => {
        const sy: jasmine.Spy = spyOn(
            fakedConfirmationDialogService,
            'openComfirmDialog'
        ).and.returnValue(undefined);
        const stopBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#stop-purge');
        stopBtn.dispatchEvent(new Event('click'));
        expect(sy.calls.count()).toEqual(1);
    });
});
