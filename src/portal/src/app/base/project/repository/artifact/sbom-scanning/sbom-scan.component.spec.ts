import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ResultSbomComponent } from './sbom-scan.component';
import {
    ScanningResultDefaultService,
    ScanningResultService,
} from '../../../../../shared/services';
import { SBOM_SCAN_STATUS } from '../../../../../shared/units/utils';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { NativeSbomReportSummary } from 'ng-swagger-gen/models';
import { SbomTipHistogramComponent } from './sbom-tip-histogram/sbom-tip-histogram.component';

describe('ResultSbomComponent (inline template)', () => {
    let component: ResultSbomComponent;
    let fixture: ComponentFixture<ResultSbomComponent>;
    let mockData: NativeSbomReportSummary = {
        scan_status: SBOM_SCAN_STATUS.SUCCESS,
        severity: 'High',
        end_time: new Date().toUTCString(),
        summary: {
            total: 124,
            fixable: 50,
            summary: {
                High: 5,
                Low: 5,
            },
        },
    };
    const mockedSbomDigest =
        'sha256:51a41cec9de9d62ee60e206f5a8a615a028a65653e45539990867417cb486285';

    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [ResultSbomComponent, SbomTipHistogramComponent],
            providers: [
                {
                    provide: ScanningResultService,
                    useValue: ScanningResultDefaultService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ResultSbomComponent);
        component = fixture.componentInstance;
        component.artifactDigest = 'mockTag';
        component.sbomDigest = mockedSbomDigest;
        component.summary = mockData;
        fixture.detectChanges();
    });

    it('should be created', () => {
        expect(component).toBeTruthy();
    });
    it('should show "scan stopped" if status is STOPPED', () => {
        component.summary.scan_status = SBOM_SCAN_STATUS.STOPPED;
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            let el: HTMLElement = fixture.nativeElement.querySelector('span');
            expect(el).toBeTruthy();
            expect(el.textContent).toEqual('SBOM.STATE.STOPPED');
        });
    });

    it('should show progress if status is SCANNING', () => {
        component.summary.scan_status = SBOM_SCAN_STATUS.RUNNING;
        fixture.detectChanges();

        fixture.whenStable().then(() => {
            fixture.detectChanges();

            let el: HTMLElement =
                fixture.nativeElement.querySelector('.progress');
            expect(el).toBeTruthy();
        });
    });

    it('should show QUEUED if status is QUEUED', () => {
        component.summary.scan_status = SBOM_SCAN_STATUS.PENDING;
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();

            let el: HTMLElement =
                fixture.nativeElement.querySelector('.bar-state');
            expect(el).toBeTruthy();
            let el2: HTMLElement = el.querySelector('span');
            expect(el2).toBeTruthy();
            expect(el2.textContent).toEqual('SBOM.STATE.QUEUED');
        });
    });

    it('should show summary bar chart if status is COMPLETED', () => {
        component.summary.scan_status = SBOM_SCAN_STATUS.SUCCESS;
        fixture.detectChanges();

        fixture.whenStable().then(() => {
            fixture.detectChanges();
            let el: HTMLElement = fixture.nativeElement.querySelector('a');
            expect(el).not.toBeNull();
        });
    });
});
