import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { ClarityModule } from "@clr/angular";
import { SharedModule } from "../../../shared/shared.module";
import { ConfigScannerService } from "../config-scanner.service";
import { of } from "rxjs";
import { ScannerMetadataComponent } from "./scanner-metadata.component";
import { ErrorHandler } from "../../../../lib/utils/error-handler";

describe('ScannerMetadataComponent', () => {
  let mockScannerMetadata = {
    scanner: {
      name: 'test1',
      vendor: 'clair',
      version: '1.0.1',
     },
    capabilities: [{
      consumes_mime_types: ['consumes_mime_types'],
      produces_mime_types: ['consumes_mime_types']
    }]
  };
  let component: ScannerMetadataComponent;
  let fixture: ComponentFixture<ScannerMetadataComponent>;
  let fakedConfigScannerService = {
    getScannerMetadata() {
      return of(mockScannerMetadata);
    }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        BrowserAnimationsModule,
        ClarityModule,
      ],
      declarations: [
        ScannerMetadataComponent
      ],
      providers: [
        ErrorHandler,
        { provide: ConfigScannerService, useValue: fakedConfigScannerService },
      ]
    })
    .compileComponents();
  }));
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
      let el: HTMLElement = fixture.nativeElement.querySelector('#scannerMetadata-name');
      expect(el.textContent).toEqual('test1');
    });
  });
});
