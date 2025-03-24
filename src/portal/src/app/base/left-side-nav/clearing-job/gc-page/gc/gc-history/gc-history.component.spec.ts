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
    ComponentFixtureAutoDetect,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { of } from 'rxjs';
import { GcHistoryComponent } from './gc-history.component';
import { SharedTestingModule } from '../../../../../../shared/shared.module';
import { GCHistory } from '../../../../../../../../ng-swagger-gen/models/gchistory';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../../../../ng-swagger-gen/models/registry';
import { GcService } from '../../../../../../../../ng-swagger-gen/services/gc.service';
import { CURRENT_BASE_HREF } from '../../../../../../shared/units/utils';
import { delay } from 'rxjs/operators';
import { ConfirmationDialogService } from '../../../../../global-confirmation-dialog/confirmation-dialog.service';

describe('GcHistoryComponent', () => {
    let component: GcHistoryComponent;
    let fixture: ComponentFixture<GcHistoryComponent>;
    const mockJobs: GCHistory[] = [
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
    const fakedGcService = {
        count: 0,
        getGCHistoryResponse() {
            if (this.count === 0) {
                this.count += 1;
                const response: HttpResponse<Array<Registry>> =
                    new HttpResponse<Array<Registry>>({
                        headers: new HttpHeaders({
                            'x-total-count': [mockJobs[0]].length.toString(),
                        }),
                        body: [mockJobs[0]],
                    });
                return of(response).pipe(delay(0));
            } else {
                this.count += 1;
                const response: HttpResponse<Array<Registry>> =
                    new HttpResponse<Array<Registry>>({
                        headers: new HttpHeaders({
                            'x-total-count': [mockJobs[1]].length.toString(),
                        }),
                        body: [mockJobs[1]],
                    });
                return of(response).pipe(delay(0));
            }
        },
        stopGC() {
            return of(null);
        },
    };
    const fakedConfirmationDialogService = {
        openComfirmDialog() {},
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [GcHistoryComponent],
            imports: [SharedTestingModule],
            providers: [
                {
                    provide: ConfirmationDialogService,
                    useValue: fakedConfirmationDialogService,
                },
                { provide: GcService, useValue: fakedGcService },
                // open auto detect
                { provide: ComponentFixtureAutoDetect, useValue: true },
            ],
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
    it('should retry getting jobs', fakeAsync(() => {
        tick(10000);
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(component.jobs[0].job_status).toEqual('finished');
        });
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
