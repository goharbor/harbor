import { ComponentFixture, TestBed } from '@angular/core/testing';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ArtifactListTabComponent } from './artifact-list-tab.component';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { ActivatedRoute, Router } from '@angular/router';
import {
    ArtifactDefaultService,
    ArtifactService,
} from '../../../artifact.service';
import {
    ScanningResultDefaultService,
    ScanningResultService,
} from '../../../../../../../shared/services';
import { ArtifactFront as Artifact, ArtifactFront } from '../../../artifact';
import { ErrorHandler } from '../../../../../../../shared/units/error-handler';
import { OperationService } from '../../../../../../../shared/components/operation/operation.service';
import { ArtifactService as NewArtifactService } from '../../../../../../../../../ng-swagger-gen/services/artifact.service';
import { Tag } from '../../../../../../../../../ng-swagger-gen/models/tag';
import { SharedTestingModule } from '../../../../../../../shared/shared.module';
import { AppConfigService } from '../../../../../../../services/app-config.service';
import { ArtifactListPageService } from '../../artifact-list-page.service';
import { ClrLoadingState } from '@clr/angular';
import { Accessory } from 'ng-swagger-gen/models/accessory';
import { ArtifactModule } from '../../../artifact.module';
import {
    SBOM_SCAN_STATUS,
    VULNERABILITY_SCAN_STATUS,
} from '../../../../../../../shared/units/utils';
import { Scanner } from '../../../../../../left-side-nav/interrogation-services/scanner/scanner';

describe('ArtifactListTabComponent', () => {
    let comp: ArtifactListTabComponent;
    let fixture: ComponentFixture<ArtifactListTabComponent>;
    const mockScanner = {
        name: 'Trivy',
        vendor: 'vm',
        version: 'v1.2',
    };
    const mockActivatedRoute = {
        snapshot: {
            params: {
                parent: {
                    parent: {
                        id: 1,
                        repo: 'test',
                        digest: 'ABC',
                    },
                },
            },
            data: {
                projectResolver: {
                    has_project_admin_role: true,
                    current_user_role_id: 3,
                    name: 'demo',
                },
            },
        },
        data: of({
            projectResolver: {
                name: 'library',
            },
        }),
    };
    const mockArtifacts: Artifact[] = [
        {
            id: 1,
            type: 'image',
            tags: [
                {
                    id: 1,
                    name: 'tag1',
                    artifact_id: 1,
                },
                {
                    id: 2,
                    name: 'tag2',
                    artifact_id: 2,
                    pull_time: '2020-01-06T09:40:08.036866579Z',
                    push_time: '2020-01-06T09:40:08.036866579Z',
                },
            ],
            references: [],
            media_type: 'string',
            digest: 'sha256:4875cda368906fd670c9629b5e416ab3d6c0292015f3c3f12ef37dc9a32fc8d4',
            size: 20372934,
            scan_overview: {
                'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0':
                    {
                        report_id: '5e64bc05-3102-11ea-93ae-0242ac140004',
                        scan_status: 'Error',
                        severity: '',
                        duration: 118,
                        summary: null,
                        start_time: '2020-01-07T04:01:23.157711Z',
                        end_time: '2020-01-07T04:03:21.662766Z',
                    },
            },
            labels: [
                {
                    id: 3,
                    name: 'aaa',
                    description: '',
                    color: '#0095D3',
                    scope: 'g',
                    project_id: 0,
                    creation_time: '2020-01-13T05:44:00.580198Z',
                    update_time: '2020-01-13T05:44:00.580198Z',
                },
                {
                    id: 6,
                    name: 'dbc',
                    description: '',
                    color: '',
                    scope: 'g',
                    project_id: 0,
                    creation_time: '2020-01-13T08:27:19.279123Z',
                    update_time: '2020-01-13T08:27:19.279123Z',
                },
            ],
            push_time: '2020-01-07T03:33:41.162319Z',
            pull_time: '0001-01-01T00:00:00Z',
        },
        {
            id: 1,
            type: 'image',
            tags: [
                {
                    id: 1,
                    name: 'tag1',
                    artifact_id: 1,
                },
                {
                    id: 2,
                    name: 'tag2',
                    artifact_id: 2,
                    pull_time: '2020-01-06T09:40:08.036866579Z',
                    push_time: '2020-01-06T09:40:08.036866579Z',
                },
            ],
            references: [],
            media_type: 'string',
            digest: 'sha256:3e33e3e3',
            size: 20372934,
            scan_overview: {
                'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0':
                    {
                        report_id: '5e64bc05-3102-11ea-93ae-0242ac140004',
                        scan_status: 'Error',
                        severity: '',
                        duration: 118,
                        summary: null,
                        start_time: '2020-01-07T04:01:23.157711Z',
                        end_time: '2020-01-07T04:03:21.662766Z',
                    },
            },
            labels: [
                {
                    id: 3,
                    name: 'aaa',
                    description: '',
                    color: '#0095D3',
                    scope: 'g',
                    project_id: 0,
                    creation_time: '2020-01-13T05:44:00.580198Z',
                    update_time: '2020-01-13T05:44:00.580198Z',
                },
                {
                    id: 6,
                    name: 'dbc',
                    description: '',
                    color: '',
                    scope: 'g',
                    project_id: 0,
                    creation_time: '2020-01-13T08:27:19.279123Z',
                    update_time: '2020-01-13T08:27:19.279123Z',
                },
            ],
            push_time: '2020-01-07T03:33:41.162319Z',
            pull_time: '0001-01-01T00:00:00Z',
        },
    ];
    const mockAccessory = <Accessory>{
        id: 1,
        artifact_id: 2,
        subject_artifact_id: 3,
        subject_artifact_digest: 'fakeDigest',
        subject_artifact_repo: 'test',
        size: 120,
        digest: 'fakeDigest',
        type: 'test',
    };
    const mockErrorHandler = {
        error: () => {},
    };
    const mockRouter = {
        events: {
            subscribe: () => {
                return of(null);
            },
        },
        navigate: () => {},
    };
    const mockOperationService = {
        publishInfo: () => {},
    };
    const mockTag: Tag = {
        id: 1,
        name: 'latest',
    };
    const mockNewArtifactService = {
        TriggerArtifactChan$: {
            subscribe: fn => {},
        },
        listAccessoriesResponse() {
            const res: HttpResponse<Array<Accessory>> = new HttpResponse<
                Array<Accessory>
            >({
                headers: new HttpHeaders({ 'x-total-count': '0' }),
                body: [],
            });
            return of(res).pipe(delay(0));
        },
        listAccessories() {
            return of(null).pipe(delay(0));
        },
        listArtifactsResponse: () => {
            return of({
                headers: new HttpHeaders({ 'x-total-count': '2' }),
                body: mockArtifacts,
            }).pipe(delay(0));
        },
        deleteArtifact: () => of(null),
        getIconsFromBackEnd() {
            return undefined;
        },
        getIcon() {
            return undefined;
        },
        listTagsResponse: () => {
            const res: HttpResponse<Array<Tag>> = new HttpResponse<Array<Tag>>({
                headers: new HttpHeaders({ 'x-total-count': '1' }),
                body: [mockTag],
            });
            return of(res).pipe(delay(0));
        },
    };
    const mockedAppConfigService = {
        getConfig() {
            return {};
        },
    };

    const mockedArtifactListPageService = {
        getScanBtnState(): ClrLoadingState {
            return ClrLoadingState.DEFAULT;
        },
        getSbomBtnState(): ClrLoadingState {
            return ClrLoadingState.DEFAULT;
        },
        hasEnabledScanner(): boolean {
            return true;
        },
        hasSbomPermission(): boolean {
            return true;
        },
        hasScannerSupportSBOM(): boolean {
            return true;
        },
        hasAddLabelImagePermission(): boolean {
            return true;
        },
        hasRetagImagePermission(): boolean {
            return true;
        },
        hasDeleteImagePermission(): boolean {
            return true;
        },
        hasScanImagePermission(): boolean {
            return true;
        },
        getProjectScanner(): Scanner {
            return mockScanner;
        },
        init() {},
    };
    beforeEach(async () => {
        await TestBed.configureTestingModule({
            imports: [SharedTestingModule, ArtifactModule],
            schemas: [NO_ERRORS_SCHEMA],
            declarations: [ArtifactListTabComponent],
            providers: [
                {
                    provide: ArtifactListPageService,
                    useValue: mockedArtifactListPageService,
                },
                { provide: ArtifactService, useClass: ArtifactDefaultService },
                { provide: AppConfigService, useValue: mockedAppConfigService },
                { provide: Router, useValue: mockRouter },
                { provide: ArtifactService, useValue: mockNewArtifactService },
                {
                    provide: ScanningResultService,
                    useClass: ScanningResultDefaultService,
                },
                { provide: ErrorHandler, useValue: mockErrorHandler },
                { provide: ActivatedRoute, useValue: mockActivatedRoute },
                { provide: OperationService, useValue: mockOperationService },
                {
                    provide: NewArtifactService,
                    useValue: mockNewArtifactService,
                },
            ],
        }).compileComponents();
    });

    beforeEach(async () => {
        fixture = TestBed.createComponent(ArtifactListTabComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
        await fixture.whenStable();
        comp.loading = false;
        fixture.detectChanges();
    });

    it('should load and render data', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        comp.artifactList = mockArtifacts;
        fixture.detectChanges();
        await fixture.whenStable();
        const el: HTMLAnchorElement =
            fixture.nativeElement.querySelector('.digest');
        expect(el).toBeTruthy();
        expect(el.textContent).toBeTruthy();
        expect(el.textContent.trim()).toEqual('sha256:4875cda3');
    });

    it('should open copy digest modal', async () => {
        await fixture.whenStable();
        comp.selectedRow = [mockArtifacts[0]];
        await stepOpenAction(fixture, comp);
        fixture.nativeElement
            .querySelector('#artifact-list-copy-digest')
            .click();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(fixture.nativeElement.querySelector('textarea')).toBeTruthy();
    });

    it('should open add labels modal', async () => {
        await fixture.whenStable();
        comp.selectedRow = [mockArtifacts[1]];
        await stepOpenAction(fixture, comp);
        fixture.nativeElement
            .querySelector('#artifact-list-add-labels')
            .click();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(
            fixture.nativeElement.querySelector('app-label-selector')
        ).toBeTruthy();
    });

    it('should open copy artifact modal', async () => {
        await fixture.whenStable();
        comp.selectedRow = [mockArtifacts[1]];
        await stepOpenAction(fixture, comp);
        fixture.nativeElement.querySelector('#artifact-list-copy').click();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(
            fixture.nativeElement.querySelector('hbr-image-name-input')
        ).toBeTruthy();
    });

    it('should open delete modal', async () => {
        await fixture.whenStable();
        comp.selectedRow = [mockArtifacts[1]];
        await stepOpenAction(fixture, comp);
        fixture.nativeElement.querySelector('#artifact-list-delete').click();
        fixture.detectChanges();
        await fixture.whenStable();
        expect(
            fixture.nativeElement.querySelector('.confirmation-title')
        ).toBeTruthy();
    });
    it('Generate SBOM button should be disabled', async () => {
        await fixture.whenStable();
        comp.selectedRow = [mockArtifacts[1]];
        await stepOpenAction(fixture, comp);
        const generatedButton =
            fixture.nativeElement.querySelector('#generate-sbom-btn');
        fixture.detectChanges();
        await fixture.whenStable();
        expect(generatedButton.disabled).toBeTruthy();
    });
    it('Stop SBOM button should be disabled', async () => {
        await fixture.whenStable();
        comp.selectedRow = [mockArtifacts[1]];
        await stepOpenAction(fixture, comp);
        const stopButton =
            fixture.nativeElement.querySelector('#stop-sbom-btn');
        fixture.detectChanges();
        await fixture.whenStable().then(() => {
            expect(stopButton.disabled).toBeTruthy();
        });
    });
    it('the length of hide array should equal to the number of column', async () => {
        comp.loading = false;
        fixture.detectChanges();
        await fixture.whenStable();
        const cols = fixture.nativeElement.querySelectorAll('.datagrid-column');
        expect(cols.length).toEqual(comp.hiddenArray.length);
    });

    it('Test isEllipsisActive', async () => {
        fixture = TestBed.createComponent(ArtifactListTabComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
        await fixture.whenStable().then(() => {
            expect(
                comp.isEllipsisActive(document.createElement('span'))
            ).toBeFalsy();
        });
    });
    it('Test deleteAccessory', async () => {
        fixture = TestBed.createComponent(ArtifactListTabComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
        comp.deleteAccessory(mockAccessory);
        fixture.detectChanges();
        await fixture.whenStable().then(() => {
            expect(
                fixture.nativeElement.querySelector('.confirmation-content')
            ).toBeTruthy();
        });
    });
    it('Test scanNow', async () => {
        fixture = TestBed.createComponent(ArtifactListTabComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
        comp.selectedRow = mockArtifacts.slice(0, 1);
        comp.scanNow();
        expect(comp.onScanArtifactsLength).toBe(1);
    });
    it('Test stopNow', async () => {
        fixture = TestBed.createComponent(ArtifactListTabComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
        comp.selectedRow = mockArtifacts.slice(0, 1);
        comp.stopNow();
        expect(comp.onStopScanArtifactsLength).toBe(1);
    });
    it('Test stopSbom', async () => {
        fixture = TestBed.createComponent(ArtifactListTabComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
        comp.selectedRow = mockArtifacts.slice(0, 1);
        comp.stopSbom();
        expect(comp.onStopSbomArtifactsLength).toBe(1);
    });
    it('Test tagsString and isRunningState and canStopSbom and canStopScan', async () => {
        fixture = TestBed.createComponent(ArtifactListTabComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
        await fixture.whenStable();
        expect(comp.tagsString([])).toBeNull();
        expect(
            comp.isRunningState(VULNERABILITY_SCAN_STATUS.RUNNING)
        ).toBeTruthy();
        expect(
            comp.isRunningState(VULNERABILITY_SCAN_STATUS.ERROR)
        ).toBeFalsy();
        expect(comp.canStopSbom()).toBeFalsy();
        expect(comp.canStopScan()).toBeFalsy();
    });
    it('Test status and handleScanOverview', async () => {
        fixture = TestBed.createComponent(ArtifactListTabComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
        await fixture.whenStable();
        expect(comp.scanStatus(mockArtifacts[0])).toBe(
            VULNERABILITY_SCAN_STATUS.ERROR
        );
        expect(comp.sbomStatus(null)).toBe(SBOM_SCAN_STATUS.NOT_GENERATED_SBOM);
        expect(comp.sbomStatus(mockArtifacts[0])).toBe(
            SBOM_SCAN_STATUS.NOT_GENERATED_SBOM
        );
        expect(comp.handleScanOverview(mockArtifacts[0])).not.toBeNull();
    });
    it('Test utils', async () => {
        fixture = TestBed.createComponent(ArtifactListTabComponent);
        comp = fixture.componentInstance;
        fixture.detectChanges();
        await fixture.whenStable();
        expect(comp.selectedRowHasSbom()).toBeFalsy();
        expect(comp.selectedRowHasVul()).toBeFalsy();
        expect(comp.canScanNow()).toBeFalsy();
        expect(comp.hasEnabledSbom()).toBeTruthy();
        expect(comp.canAddLabel()).toBeFalsy();
    });
});

async function stepOpenAction(fixture, comp) {
    comp.projectId = 1;
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.nativeElement.querySelector('#artifact-list-action').click();
    fixture.detectChanges();
    await fixture.whenStable();
}
