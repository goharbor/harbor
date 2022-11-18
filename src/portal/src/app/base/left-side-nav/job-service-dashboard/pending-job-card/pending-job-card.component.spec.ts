import { ComponentFixture, TestBed } from '@angular/core/testing';
import { PendingCardComponent } from './pending-job-card.component';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { JobQueue } from '../../../../../../ng-swagger-gen/models/job-queue';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';

describe('PendingCardComponent', () => {
    let component: PendingCardComponent;
    let fixture: ComponentFixture<PendingCardComponent>;

    const mockedJobs: JobQueue[] = [
        {
            job_type: 'test1',
            count: 1,
        },
        {
            job_type: 'test2',
            count: 2,
        },
        {
            job_type: 'test3',
            count: 3,
        },
    ];

    const fakedJobserviceService = {
        listJobQueues() {
            return of(mockedJobs).pipe(delay(0));
        },
    };

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            declarations: [PendingCardComponent],
            imports: [SharedTestingModule],
        }).compileComponents();

        fixture = TestBed.createComponent(PendingCardComponent);
        component = fixture.componentInstance;
        spyOn(component, 'loopGetPendingJobs').and.callFake(() => {
            fakedJobserviceService.listJobQueues().subscribe(res => {
                component.loading = false;
                component.jobQueue = res.sort((a, b) => {
                    const ACount: number = a?.count | 0;
                    const BCount: number = b?.count | 0;
                    return BCount - ACount;
                });
            });
        });
        fixture.autoDetectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });

    it('should render data', async () => {
        await fixture.whenStable();
        expect(component.total()).toEqual(6);
    });
});
