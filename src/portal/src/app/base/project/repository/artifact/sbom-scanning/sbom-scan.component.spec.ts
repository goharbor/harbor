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
import { of, timer } from 'rxjs';
import { ArtifactService, ScanService } from 'ng-swagger-gen/services';
import { AccessoryType } from '../artifact';

describe('ResultSbomComponent (inline template)', () => {
    let component: ResultSbomComponent;
    let fixture: ComponentFixture<ResultSbomComponent>;
    const mockedSbomDigest =
        'sha256:052240e8190b7057439d2bee1dffb9b37c8800e5c1af349f667635ae1debf8f3';
    const mockScanner = {
        name: 'Trivy',
        vendor: 'vm',
        version: 'v1.2',
    };
    let mockData: SBOMOverview = {
        scan_status: SBOM_SCAN_STATUS.SUCCESS,
        end_time: new Date().toUTCString(),
    };
    const mockedSbomOverview = {
        report_id: '12345',
        scan_status: 'Error',
    };
    const mockedCloneSbomOverview = {
        report_id: '12346',
        scan_status: 'Pending',
    };
    const mockedAccessories = [
        {
            type: AccessoryType.SBOM,
            digest: mockedSbomDigest,
        },
    ];
    const FakedScanService = {
        scanArtifact: () => of({}),
        stopScanArtifact: () => of({}),
    };
    const FakedArtifactService = {
        getArtifact: () =>
            of({
                accessories: mockedAccessories,
                addition_links: {
                    build_history: {
                        absolute: false,
                        href: '/api/v2.0/projects/xuel/repositories/ui%252Fserver%252Fconfig-dev/artifacts/sha256:052240e8190b7057439d2bee1dffb9b37c8800e5c1af349f667635ae1debf8f3/additions/build_history',
                    },
                    vulnerabilities: {
                        absolute: false,
                        href: '/api/v2.0/projects/xuel/repositories/ui%252Fserver%252Fconfig-dev/artifacts/sha256:052240e8190b7057439d2bee1dffb9b37c8800e5c1af349f667635ae1debf8f3/additions/vulnerabilities',
                    },
                },
                digest: 'sha256:052240e8190b7057439d2bee1dffb9b37c8800e5c1af349f667635ae1debf8f3',
                extra_attrs: {
                    architecture: 'amd64',
                    author: '',
                    config: {
                        Env: [
                            'PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin',
                        ],
                        WorkingDir: '/',
                    },
                    created: '2024-01-10T10:05:33.2702206Z',
                    os: 'linux',
                },
                icon: 'sha256:0048162a053eef4d4ce3fe7518615bef084403614f8bca43b40ae2e762e11e06',
                id: 3,
                labels: null,
                manifest_media_type:
                    'application/vnd.docker.distribution.manifest.v2+json',
                media_type: 'application/vnd.docker.container.image.v1+json',
                project_id: 3,
                pull_time: '2024-04-02T01:50:58.332Z',
                push_time: '2024-03-06T09:47:08.163Z',
                references: null,
                repository_id: 2,
                sbom_overview: {
                    duration: 2,
                    end_time: '2024-04-02T01:50:59.406Z',
                    sbom_digest:
                        'sha256:8cca43ea666e0e7990c2433e3b185313e6ba303cc7a3124bb767823c79fb74a6',
                    scan_status: 'Success',
                    start_time: '2024-04-02T01:50:57.176Z',
                    scanner: mockScanner,
                },
                size: 3957,
                tags: null,
                type: 'IMAGE',
            }),
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
                {
                    provide: ScanService,
                    useValue: FakedScanService,
                },
                {
                    provide: ArtifactService,
                    useValue: FakedArtifactService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ResultSbomComponent);
        component = fixture.componentInstance;
        component.repoName = 'mockRepo';
        component.inputScanner = mockScanner;
        component.artifactDigest = mockedSbomDigest;
        component.sbomDigest = mockedSbomDigest;
        component.sbomOverview = mockData;
        component.accessories = mockedAccessories;
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
        component.sbomOverview = { ...mockedSbomOverview };
        component.sbomOverview.scan_status = SBOM_SCAN_STATUS.SUCCESS;
        component.sbomOverview.sbom_digest = mockedSbomDigest;
        component.artifactDigest = mockedSbomDigest;
        component.sbomDigest = mockedSbomDigest;
        component.accessories = mockedAccessories;
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            const el: HTMLElement =
                fixture.nativeElement.querySelector('.tip-block');
            expect(el).not.toBeNull();
            const textContent = el?.textContent;
            expect(component.sbomOverview.scan_status).toBe(
                SBOM_SCAN_STATUS.SUCCESS
            );
            expect(textContent?.trim()).toBe('SBOM.Details');
        });
    });
    it('Test ResultSbomComponent getScanner', () => {
        fixture.detectChanges();
        component.inputScanner = undefined;
        expect(component.getScanner()).toBeUndefined();
        component.inputScanner = mockScanner;
        component.sbomOverview = mockedSbomOverview;
        expect(component.getScanner()).toBe(mockScanner);
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
        component.stateCheckTimer = timer(0, 10000).subscribe(() => {});
        component.ngOnDestroy();
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            expect(component.stateCheckTimer).toBeNull();
            expect(component.generateSbomSubscription).toBeNull();
            expect(component.stopSubscription).toBeNull();
        });
    });
    it('Test ResultSbomComponent generateSbom', () => {
        fixture.detectChanges();
        component.generateSbom();
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            expect(component.onSubmitting).toBeFalse();
        });
    });
    it('Test ResultSbomComponent stopSbom', () => {
        fixture.detectChanges();
        component.stopSbom();
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            expect(component.onStopping).toBeFalse();
        });
    });
    it('Test ResultSbomComponent getSbomOverview', () => {
        fixture.detectChanges();
        component.getSbomOverview();
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            fixture.detectChanges();
            expect(component.sbomOverview.scan_status).toBe(
                SBOM_SCAN_STATUS.SUCCESS
            );
        });
    });
});
