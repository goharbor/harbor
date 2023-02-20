import {
    ComponentFixture,
    ComponentFixtureAutoDetect,
    TestBed,
} from '@angular/core/testing';
import { of } from 'rxjs';
import { delay } from 'rxjs/operators';
import { ConfigurationScannerComponent } from './config-scanner.component';
import { SharedTestingModule } from '../../../../shared/shared.module';
import { ScannerMetadataComponent } from './scanner-metadata/scanner-metadata.component';
import { NewScannerModalComponent } from './new-scanner-modal/new-scanner-modal.component';
import { NewScannerFormComponent } from './new-scanner-form/new-scanner-form.component';
import { ScannerService } from '../../../../../../ng-swagger-gen/services/scanner.service';
import { HttpHeaders, HttpResponse } from '@angular/common/http';
import { Registry } from '../../../../../../ng-swagger-gen/models/registry';
import { ClrLoadingState } from '@clr/angular';

describe('ConfigurationScannerComponent', () => {
    let mockScannerMetadata = {
        scanner: {
            name: 'test1',
            vendor: 'trivy',
            version: '1.0.1',
        },
        capabilities: [
            {
                consumes_mime_types: ['consumes_mime_types'],
                produces_mime_types: ['consumes_mime_types'],
            },
        ],
    };
    let mockScanner1 = {
        name: 'test1',
        description: 'just a sample',
        version: '1.0.0',
        url: 'http://168.0.0.1',
    };
    let component: ConfigurationScannerComponent;
    let fixture: ComponentFixture<ConfigurationScannerComponent>;
    let fakedConfigScannerService = {
        getScannerMetadata() {
            return of(mockScannerMetadata).pipe(delay(10));
        },
        listScannersResponse() {
            const response: HttpResponse<Array<Registry>> = new HttpResponse<
                Array<Registry>
            >({
                headers: new HttpHeaders({
                    'x-total-count': [mockScanner1].length.toString(),
                }),
                body: [mockScanner1],
            });
            return of(response).pipe(delay(10));
        },
        updateScanner() {
            return of(true);
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [SharedTestingModule],
            declarations: [
                ConfigurationScannerComponent,
                ScannerMetadataComponent,
                NewScannerModalComponent,
                NewScannerFormComponent,
            ],
            providers: [
                {
                    provide: ScannerService,
                    useValue: fakedConfigScannerService,
                },
                // open auto detect
                { provide: ComponentFixtureAutoDetect, useValue: true },
            ],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(ConfigurationScannerComponent);
        component = fixture.componentInstance;
        component.newScannerDialog.saveBtnState = ClrLoadingState.LOADING;
        fixture.detectChanges();
    });
    it('should create', async () => {
        await fixture.whenStable();
        expect(component).toBeTruthy();
        expect(component.scanners.length).toBe(1);
    });
    it('should be clickable', () => {
        component.selectedRow = mockScanner1;
        fixture.detectChanges();
        fixture.whenStable().then(() => {
            let el: HTMLElement =
                fixture.nativeElement.querySelector('#set-default');
            expect(el.getAttribute('disable')).toBeFalsy();
        });
    });
    it('edit a scanner', () => {
        component.selectedRow = mockScanner1;
        component.editScanner();
        expect(component.newScannerDialog.opened).toBeTruthy();
        fixture.detectChanges();
        fixture.nativeElement.querySelector('#scanner-name').value = 'test456';
        fixture.nativeElement.querySelector('#button-save').click();
        fixture.detectChanges();
        expect(component.newScannerDialog.opened).toBeFalsy();
    });
});
