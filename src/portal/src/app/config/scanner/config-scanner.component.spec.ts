import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { ClarityModule } from "@clr/angular";
import { of } from "rxjs";
import { ErrorHandler } from "@harbor/ui";
import { ConfigurationScannerComponent } from "./config-scanner.component";
import { ConfigScannerService } from "./config-scanner.service";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { ConfirmationDialogService } from "../../shared/confirmation-dialog/confirmation-dialog.service";
import { SharedModule } from "../../shared/shared.module";
import { ScannerMetadataComponent } from "./scanner-metadata/scanner-metadata.component";
import { NewScannerModalComponent } from "./new-scanner-modal/new-scanner-modal.component";
import { NewScannerFormComponent } from "./new-scanner-form/new-scanner-form.component";
import { TranslateService } from "@ngx-translate/core";

describe('ConfigurationScannerComponent', () => {
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
  let mockScanner1 = {
    name: 'test1',
    description: 'just a sample',
    version: '1.0.0',
    url: 'http://168.0.0.1'
  };
  let component: ConfigurationScannerComponent;
  let fixture: ComponentFixture<ConfigurationScannerComponent>;
  let fakedConfigScannerService = {
    getScannerMetadata() {
      return of(mockScannerMetadata);
    },
    getScanners() {
      return of([mockScanner1]);
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
        ConfigurationScannerComponent,
        ScannerMetadataComponent,
        NewScannerModalComponent,
        NewScannerFormComponent
      ],
      providers: [
        ErrorHandler,
        MessageHandlerService,
        ConfirmationDialogService,
        TranslateService,
        { provide: ConfigScannerService, useValue: fakedConfigScannerService },
      ]
    })
    .compileComponents();
  }));
  beforeEach(() => {
    fixture = TestBed.createComponent(ConfigurationScannerComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });
  it('should create', () => {
    expect(component).toBeTruthy();
    expect(component.scanners.length).toBe(1);
  });
  it('should be clickable', () => {
    component.selectedRow = mockScanner1;
    fixture.detectChanges();
    fixture.whenStable().then(() => {
      let el: HTMLElement = fixture.nativeElement.querySelector('#set-default');
      expect(el.getAttribute('disable')).toBeFalsy();
    });
  });
});
