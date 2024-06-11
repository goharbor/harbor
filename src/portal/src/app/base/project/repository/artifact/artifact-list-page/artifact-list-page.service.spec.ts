import { inject, TestBed } from '@angular/core/testing';
import { ArtifactListPageService } from './artifact-list-page.service';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import {
    ScanningResultService,
    UserPermissionService,
} from 'src/app/shared/services';
import { of } from 'rxjs';
import { ClrLoadingState } from '@clr/angular';

describe('ArtifactListPageService', () => {
    const FakedScanningResultService = {
        getProjectScanner: () =>
            of({
                access_credential: '',
                adapter: 'Trivy',
                auth: '',
                capabilities: {
                    support_sbom: true,
                    support_vulnerability: true,
                },
                create_time: '2024-03-06T09:29:43.789Z',
                description: 'The Trivy scanner adapter',
                disabled: false,
                health: 'healthy',
                is_default: true,
                name: 'Trivy',
                skip_certVerify: false,
                update_time: '2024-03-06T09:29:43.789Z',
                url: 'http://trivy-adapter:8080',
                use_internal_addr: true,
                uuid: '10c68b62-db9c-11ee-9c72-0242ac130009',
                vendor: 'Aqua Security',
                version: 'v0.47.0',
            }),
    };
    const FakedUserPermissionService = {
        hasProjectPermissions: () => of([true, true, true, true, true]),
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            providers: [
                ArtifactListPageService,
                {
                    provide: ScanningResultService,
                    useValue: FakedScanningResultService,
                },
                {
                    provide: UserPermissionService,
                    useValue: FakedUserPermissionService,
                },
            ],
        });
    });

    it('should be initialized', inject(
        [ArtifactListPageService],
        (service: ArtifactListPageService) => {
            expect(service).toBeTruthy();
        }
    ));
    it('Test ArtifactListPageService Permissions validation ', inject(
        [ArtifactListPageService],
        (service: ArtifactListPageService) => {
            service.init(3);
            expect(service.hasSbomPermission()).toBeTruthy();
            expect(service.hasAddLabelImagePermission()).toBeTruthy();
            expect(service.hasRetagImagePermission()).toBeTruthy();
            expect(service.hasDeleteImagePermission()).toBeTruthy();
            expect(service.hasScanImagePermission()).toBeTruthy();
            expect(service.hasScannerSupportVulnerability()).toBeTruthy();
            expect(service.hasScannerSupportSBOM()).toBeTruthy();
        }
    ));
    it('Test ArtifactListPageService updateStates', inject(
        [ArtifactListPageService],
        (service: ArtifactListPageService) => {
            service.init(3);
            expect(service.hasEnabledScanner()).toBeTruthy();
            expect(service.getScanBtnState()).toBe(ClrLoadingState.SUCCESS);
            expect(service.getSbomBtnState()).toBe(ClrLoadingState.SUCCESS);
            service.updateStates(
                false,
                ClrLoadingState.ERROR,
                ClrLoadingState.ERROR
            );
            expect(service.hasEnabledScanner()).toBeFalsy();
            expect(service.getScanBtnState()).toBe(ClrLoadingState.ERROR);
            expect(service.getSbomBtnState()).toBe(ClrLoadingState.ERROR);
        }
    ));
    it('Test ArtifactListPageService updateCapabilities ', inject(
        [ArtifactListPageService],
        (service: ArtifactListPageService) => {
            service.updateCapabilities({
                support_vulnerability: true,
                support_sbom: true,
            });
            expect(service.hasScannerSupportVulnerability()).toBeTruthy();
            expect(service.hasScannerSupportSBOM()).toBeTruthy();
            service.updateCapabilities({
                support_vulnerability: false,
                support_sbom: false,
            });
            expect(service.hasScannerSupportVulnerability()).toBeFalsy();
            expect(service.hasScannerSupportSBOM()).toBeFalsy();
        }
    ));
});
