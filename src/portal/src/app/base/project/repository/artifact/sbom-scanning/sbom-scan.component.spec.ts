import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ResultSbomComponent } from './sbom-scan.component';
import {
    ScanningResultDefaultService,
    ScanningResultService,
} from '../../../../../shared/services';
import { SBOM_SCAN_STATUS } from '../../../../../shared/units/utils';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { SbomTipHistogramComponent } from './sbom-tip-histogram/sbom-tip-histogram.component';
import { SBOMOverview } from './sbom-overview';
import { Subscription } from 'rxjs';

describe('ResultSbomComponent (inline template)', () => {
    let component: ResultSbomComponent;
    let fixture: ComponentFixture<ResultSbomComponent>;
    let mockData: SBOMOverview = {
        scan_status: SBOM_SCAN_STATUS.SUCCESS,
        end_time: new Date().toUTCString(),
    };
    const mockedSbomDigest =
        'sha256:51a41cec9de9d62ee60e206f5a8a615a028a65653e45539990867417cb486285';
    const mockedSbomOverview = {
        report_id: '12345',
        scan_status: 'Error',
        scanner: {
            name: 'Trivy',
            vendor: 'vm',
            version: 'v1.2',
        },
    };
    const mockedCloneSbomOverview = {
        report_id: '12346',
        scan_status: 'Pending',
        scanner: {
            name: 'Trivy',
            vendor: 'vm',
            version: 'v1.2',
        },
    };

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
        component.sbomOverview = mockData;
        fixture.detectChanges();
    });

    it('should be created', () => {
        expect(component).toBeTruthy();
    });
    it('should show "scan stopped" if status is STOPPED', () => {
        component.sbomOverview.scan_status = SBOM_SCAN_STATUS.STOPPED;
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            let el: HTMLElement = fixture.nativeElement.querySelector('span');
            expect(el).toBeTruthy();
            expect(el.textContent).toEqual('SBOM.STATE.STOPPED');
        });
    });

    it('should show progress if status is SCANNING', () => {
        component.sbomOverview.scan_status = SBOM_SCAN_STATUS.RUNNING;
        fixture.detectChanges();

        fixture.whenStable().then(() => {
            fixture.detectChanges();

            let el: HTMLElement =
                fixture.nativeElement.querySelector('.progress');
            expect(el).toBeTruthy();
        });
    });

    it('should show QUEUED if status is QUEUED', () => {
        component.sbomOverview.scan_status = SBOM_SCAN_STATUS.PENDING;
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
        component.sbomOverview.scan_status = SBOM_SCAN_STATUS.SUCCESS;
        fixture.detectChanges();

        fixture.whenStable().then(() => {
            fixture.detectChanges();
            let el: HTMLElement = fixture.nativeElement.querySelector('a');
            expect(el).not.toBeNull();
        });
    });
    it('Test ResultSbomComponent getScanner', () => {
        fixture.detectChanges();
        expect(component.getScanner()).toBeUndefined();
        component.sbomOverview = mockedSbomOverview;
        expect(component.getScanner()).toBe(mockedSbomOverview.scanner);
        component.projectName = 'test';
        component.repoName = 'ui';
        component.artifactDigest = 'dg';
        expect(component.viewLog()).toBe(
            '/api/v2.0/projects/test/repositories/ui/artifacts/dg/scan/12345/log'
        );
        component.copyValue(mockedCloneSbomOverview);
        expect(component.sbomOverview.report_id).toBe(
            mockedCloneSbomOverview.report_id
        );
    });
    it('Test ResultSbomComponent status', () => {
        component.sbomOverview = mockedSbomOverview;
        fixture.detectChanges();
        expect(component.status).toBe(SBOM_SCAN_STATUS.ERROR);
        expect(component.completed).toBeFalsy();
        expect(component.queued).toBeFalsy();
        expect(component.generating).toBeFalsy();
        expect(component.stopped).toBeFalsy();
        expect(component.otherStatus).toBeFalsy();
        expect(component.error).toBeTruthy();
    });
    it('Test ResultSbomComponent ngOnDestroy', () => {
        component.ngOnDestroy();
        expect(component.stateCheckTimer).toBeUndefined();
    });
});
