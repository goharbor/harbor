import { ComponentFixture, TestBed } from '@angular/core/testing';
import { ArtifactSbomComponent } from './artifact-sbom.component';
import { NO_ERRORS_SCHEMA } from '@angular/core';
import { ClarityModule } from '@clr/angular';
import { of } from 'rxjs';
import {
    TranslateFakeLoader,
    TranslateLoader,
    TranslateModule,
} from '@ngx-translate/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { UserPermissionService } from '../../../../../../shared/services';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { SessionService } from '../../../../../../shared/services/session.service';
import { SessionUser } from '../../../../../../shared/entities/session-user';
import { AppConfigService } from 'src/app/services/app-config.service';
import { ArtifactSbomPackageItem } from '../../artifact';
import { ArtifactService } from 'ng-swagger-gen/services';
import { ArtifactListPageService } from '../../artifact-list-page/artifact-list-page.service';

describe('ArtifactSbomComponent', () => {
    let component: ArtifactSbomComponent;
    let fixture: ComponentFixture<ArtifactSbomComponent>;
    const artifactSbomPackages: ArtifactSbomPackageItem[] = [
        {
            name: 'alpine-baselayout',
            SPDXID: 'SPDXRef-Package-5b53573c19a59415',
            versionInfo: '3.2.0-r18',
            supplier: 'NOASSERTION',
            downloadLocation: 'NONE',
            checksums: [
                {
                    algorithm: 'SHA1',
                    checksumValue: '132992eab020986b3b5d886a77212889680467a0',
                },
            ],
            sourceInfo: 'built package from: alpine-baselayout 3.2.0-r18',
            licenseConcluded: 'GPL-2.0-only',
            licenseDeclared: 'GPL-2.0-only',
            copyrightText: '',
            externalRefs: [
                {
                    referenceCategory: 'PACKAGE-MANAGER',
                    referenceType: 'purl',
                    referenceLocator:
                        'pkg:apk/alpine/alpine-baselayout@3.2.0-r18?arch=x86_64\u0026distro=3.15.5',
                },
            ],
            attributionTexts: [
                'PkgID: alpine-baselayout@3.2.0-r18',
                'LayerDiffID: sha256:ad543cd673bd9de2bac48599da992506dcc37a183179302ea934853aaa92cb84',
            ],
            primaryPackagePurpose: 'LIBRARY',
        },
        {
            name: 'alpine-keys',
            SPDXID: 'SPDXRef-Package-7e5952f7a76e9643',
            versionInfo: '2.4-r1',
            supplier: 'NOASSERTION',
            downloadLocation: 'NONE',
            checksums: [
                {
                    algorithm: 'SHA1',
                    checksumValue: '903176b2d2a8ddefd1ba6940f19ad17c2c1d4aff',
                },
            ],
            sourceInfo: 'built package from: alpine-keys 2.4-r1',
            licenseConcluded: 'MIT',
            licenseDeclared: 'MIT',
            copyrightText: '',
            externalRefs: [
                {
                    referenceCategory: 'PACKAGE-MANAGER',
                    referenceType: 'purl',
                    referenceLocator:
                        'pkg:apk/alpine/alpine-keys@2.4-r1?arch=x86_64\u0026distro=3.15.5',
                },
            ],
            attributionTexts: [
                'PkgID: alpine-keys@2.4-r1',
                'LayerDiffID: sha256:ad543cd673bd9de2bac48599da992506dcc37a183179302ea934853aaa92cb84',
            ],
            primaryPackagePurpose: 'LIBRARY',
        },
    ];
    const artifactSbomJson = {
        spdxVersion: 'SPDX-2.3',
        dataLicense: 'CC0-1.0',
        SPDXID: 'SPDXRef-DOCUMENT',
        name: 'alpine:3.15.5',
        documentNamespace:
            'http://aquasecurity.github.io/trivy/container_image/alpine:3.15.5-7ead854c-7340-44c9-bbbf-5403c21cc9b6',
        creationInfo: {
            licenseListVersion: '',
            creators: ['Organization: aquasecurity', 'Tool: trivy-0.47.0'],
            created: '2023-11-29T07:06:22Z',
        },
        packages: artifactSbomPackages,
    };
    const fakedArtifactService = {
        getAddition() {
            return of(JSON.stringify(artifactSbomJson));
        },
    };
    const fakedUserPermissionService = {
        hasProjectPermissions() {
            return of(true);
        },
    };
    const fakedAppConfigService = {
        getConfig() {
            return of({ sbom_enabled: true });
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
    const mockedSbomDigest =
        'sha256:51a41cec9de9d62ee60e206f5a8a615a028a65653e45539990867417cb486285';
    const mockedArtifactListPageService = {
        hasScannerSupportSBOM(): boolean {
            return true;
        },
        init() {},
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
            declarations: [ArtifactSbomComponent],
            providers: [
                ErrorHandler,
                { provide: AppConfigService, useValue: fakedAppConfigService },
                { provide: ArtifactService, useValue: fakedArtifactService },
                {
                    provide: UserPermissionService,
                    useValue: fakedUserPermissionService,
                },
                { provide: SessionService, useValue: fakedSessionService },
                {
                    provide: ArtifactListPageService,
                    useValue: mockedArtifactListPageService,
                },
            ],
            schemas: [NO_ERRORS_SCHEMA],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(ArtifactSbomComponent);
        component = fixture.componentInstance;
        component.hasSbomPermission = true;
        component.hasScannerSupportSBOM = true;
        component.sbomDigest = mockedSbomDigest;
        component.ngOnInit();
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get sbom list and render', async () => {
        fixture.detectChanges();
        await fixture.whenStable();
        const rows = fixture.nativeElement.getElementsByTagName('clr-dg-row');
        expect(rows.length).toEqual(2);
    });

    it('download button should show the right text', async () => {
        fixture.autoDetectChanges(true);
        const scanBtn: HTMLButtonElement =
            fixture.nativeElement.querySelector('#sbom-btn');
        expect(scanBtn.innerText).toContain('SBOM.DOWNLOAD');
    });
});
