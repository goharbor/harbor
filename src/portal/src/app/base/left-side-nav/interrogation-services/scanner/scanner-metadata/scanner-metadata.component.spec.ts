import { ComponentFixture, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { SharedTestingModule } from '../../../../../shared/shared.module';
import { of } from 'rxjs';
import { ScannerMetadataComponent } from './scanner-metadata.component';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { ScannerService } from '../../../../../../../ng-swagger-gen/services/scanner.service';

describe('ScannerMetadataComponent', () => {
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
    let component: ScannerMetadataComponent;
    let fixture: ComponentFixture<ScannerMetadataComponent>;
    let fakedConfigScannerService = {
        getScannerMetadata() {
            return of(mockScannerMetadata);
        },
    };
    beforeEach(() => {
        TestBed.configureTestingModule({
            imports: [
                SharedTestingModule,
                BrowserAnimationsModule,
                ClarityModule,
            ],
            declarations: [ScannerMetadataComponent],
            providers: [
                ErrorHandler,
                {
                    provide: ScannerService,
                    useValue: fakedConfigScannerService,
                },
            ],
        });
    });
    beforeEach(() => {
        fixture = TestBed.createComponent(ScannerMetadataComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });
    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('should get metadata', () => {
        fixture.whenStable().then(() => {
            let el: HTMLElement = fixture.nativeElement.querySelector(
                '#scannerMetadata-name'
            );
            expect(el.textContent).toEqual('test1');
        });
    });
});
