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
    discardPeriodicTasks,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { of, Subject } from 'rxjs';
import { delay } from 'rxjs/operators';
import { GcHistoryComponent } from './gc-history.component';
import { SharedTestingModule } from '../../../../../../shared/shared.module';
import { GCHistory } from '../../../../../../../../ng-swagger-gen/models/gchistory';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { GcService } from '../../../../../../../../ng-swagger-gen/services/gc.service';
import { CURRENT_BASE_HREF } from '../../../../../../shared/units/utils';
import { ConfirmationDialogService } from '../../../../../global-confirmation-dialog/confirmation-dialog.service';
import { ConfirmationAcknowledgement } from '../../../../../global-confirmation-dialog/confirmation-state-message';
import { JOB_STATUS } from '../../../clearing-job-interfact';

describe('GcHistoryComponent', () => {
    let component: GcHistoryComponent;
    let fixture: ComponentFixture<GcHistoryComponent>;
    const mockJobs: GCHistory[] = [
        {
            id: 1,
            job_name: 'test',
            job_kind: 'manual',
            schedule: null,
            job_status: JOB_STATUS.PENDING,
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
    const fakedGcService = {
        callCount: 0,
        getGCHistoryResponse() {
            this.callCount += 1;
            const pendingResponse: HttpResponse<GCHistory[]> = new HttpResponse<
                GCHistory[]
            >({
                headers: new HttpHeaders({
                    'x-total-count': '1',
                }),
                body: [mockJobs[0]],
            });
            // Clarity datagrid can invoke refresh more than once on init; keep both as "first load".
            if (this.callCount <= 2) {
                return of(pendingResponse).pipe(delay(0));
            }
            const response: HttpResponse<GCHistory[]> = new HttpResponse<
                GCHistory[]
            >({
                headers: new HttpHeaders({
                    'x-total-count': '1',
                }),
                body: [{ ...mockJobs[0], job_status: 'finished' }],
            });
            return of(response);
        },
        stopGC() {
            return of(null);
        },
    };
    const confirmationConfirmSource =
        new Subject<ConfirmationAcknowledgement>();
    const fakedConfirmationDialogService = {
        openComfirmDialog() {},
        confirmationConfirm$: confirmationConfirmSource.asObservable(),
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [GcHistoryComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: ConfirmationDialogService,
                    useValue: fakedConfirmationDialogService,
                },
                { provide: GcService, useValue: fakedGcService },
            ],
        }).compileComponents();
    });

    beforeEach(fakeAsync(() => {
        fakedGcService.callCount = 0;
        fixture = TestBed.createComponent(GcHistoryComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
        tick(0);
        discardPeriodicTasks();
    }));
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
        expect(component.jobs.length).toBe(1);
        expect(component.jobs[0].job_status).toEqual(JOB_STATUS.PENDING);
        component.getJobs(false, component.state);
        expect(component.jobs[0].job_status).toEqual('finished');
    }));
    it('should return right log link', () => {
        expect(component.getLogLink('1')).toEqual(
            `${CURRENT_BASE_HREF}/system/gc/1/log`
        );
    });
    it('stopping GC should work', () => {
        const sy: jasmine.Spy = spyOn(
            fakedConfirmationDialogService,
            'openComfirmDialog'
        ).and.returnValue(undefined);
        const stopBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#stop-gc');
        stopBtn.dispatchEvent(new Event('click'));
        expect(sy.calls.count()).toEqual(1);
    });
});
