import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactVulnerabilitiesComponent } from './artifact-vulnerabilities.component';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { AdditionsService } from '../additions.service';
import { of } from 'rxjs';
import {
    TranslateFakeLoader,
    TranslateLoader,
    TranslateModule,
} from '@ngx-translate/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import {
    ProjectService,
    ScanningResultService,
    SystemInfoService,
    UserPermissionService,
    VulnerabilityItem,
} from '../../../../../../shared/services';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { SessionService } from '../../../../../../shared/services/session.service';
import { SessionUser } from '../../../../../../shared/entities/session-user';
import { delay } from 'rxjs/operators';

describe('ArtifactVulnerabilitiesComponent', () => {
    let component: ArtifactVulnerabilitiesComponent;
    let fixture: ComponentFixture<ArtifactVulnerabilitiesComponent>;
    const mockedVulnerabilities: VulnerabilityItem[] = [
        {
            id: '123',
            severity: 'low',
            package: 'test',
            version: '1.0',
            links: ['testLink'],
            fix_version: '1.1.1',
            description: 'just a test',
        },
        {
            id: '456',
            severity: 'high',
            package: 'test',
            version: '1.0',
            links: ['testLink'],
            fix_version: '1.1.1',
            description: 'just a test',
        },
    ];
    let scanOverview = {};
    scanOverview['keyForTest'] = {};
    scanOverview['keyForTest'].vulnerabilities = mockedVulnerabilities;
    const mockedLink: AdditionLink = {
        absolute: false,
        href: '/test',
    };
    const fakedAdditionsService = {
        getDetailByLink() {
            return of(scanOverview);
        },
    };
    const fakedUserPermissionService = {
        hasProjectPermissions() {
            return of(true);
        },
    };
    const fakedScanningResultService = {
        getProjectScanner() {
            return of(true);
        },
    };
    const mockedUser: SessionUser = {
        user_id: 1,
        username: 'admin',
        email: 'harbor@vmware.com',
        realname: 'admin',
        has_admin_role: true,
        comment: 'no comment',
    };
    const fakedSessionService = {
        getCurrentUser() {
            return mockedUser;
        },
    };
    const fakedProjectService = {
        getProject() {
            return of({
                name: 'test',
                metadata: {
                    reuse_sys_cve_allowlist: 'false',
                },
                cve_allowlist: {
                    items: [{ cve_id: '123' }],
                },
            }).pipe(delay(0));
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot({
                    loader: {
                        provide: TranslateLoader,
                        useClass: TranslateFakeLoader,
                    },
                }),
            ],
            declarations: [ArtifactVulnerabilitiesComponent],
            providers: [
                ErrorHandler,
                SystemInfoService,
                { provide: AdditionsService, useValue: fakedAdditionsService },
                {
                    provide: UserPermissionService,
                    useValue: fakedUserPermissionService,
                },
                {
                    provide: ScanningResultService,
                    useValue: fakedScanningResultService,
                },
                { provide: SessionService, useValue: fakedSessionService },
                { provide: ProjectService, useValue: fakedProjectService },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactVulnerabilitiesComponent);
        component = fixture.componentInstance;
        component.hasScanningPermission = true;
        component.hasEnabledScanner = true;
        component.vulnerabilitiesLink = mockedLink;
        component.ngOnInit();
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get vulnerability list and render', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
        expect(rows.length).toEqual(2);
    });
    it("should show column 'Listed In CVE Allowlist'", async () => {
        fixture.autoDetectChanges(true);
        await fixture.whenStable();
        const cols = fixture.nativeElement.querySelectorAll('clr-dg-column');
        expect(cols).toBeTruthy();
        expect(cols.length).toEqual(7);
        const firstRow = fixture.nativeElement.querySelector('clr-dg-row');
        const cells = firstRow.querySelectorAll('clr-dg-cell');
        expect(cells[cells.length - 1].innerText).toEqual('TAG_RETENTION.YES');
    });

    it('scan button should show the right text', async () => {
        fixture.autoDetectChanges(true);
        const scanBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#scan-btn');
        expect(scanBtn.innerText).toContain('VULNERABILITY.SCAN_NOW');
    });
});
