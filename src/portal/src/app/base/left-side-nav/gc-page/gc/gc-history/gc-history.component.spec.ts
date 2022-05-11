import {
    ComponentFixture,
    ComponentFixtureAutoDetect,
    fakeAsync,
    TestBed,
    tick,
} from '@angular/core/testing';
import { of } from 'rxjs';
import { GcHistoryComponent } from './gc-history.component';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { GCHistory } from '../../../../../../../ng-swagger-gen/models/gchistory';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../../../ng-swagger-gen/models/registry';
import { GcService } from '../../../../../../../ng-swagger-gen/services/gc.service';
import { CURRENT_BASE_HREF } from '../../../../../shared/units/utils';
import { delay } from 'rxjs/operators';

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
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            declarations: [GcHistoryComponent],
            imports: [SharedTestingModule],
            providers: [
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
});
