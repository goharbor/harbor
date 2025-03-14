// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
