import { async, ComponentFixture, ComponentFixtureAutoDetect, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { ClarityModule } from "@clr/angular";
import { of } from "rxjs";
import { ConfigurationScannerComponent } from "./config-scanner.component";
import { ConfigScannerService } from "./config-scanner.service";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { ConfirmationDialogService } from "../../shared/confirmation-dialog/confirmation-dialog.service";
import { SharedModule } from "../../shared/shared.module";
import { ScannerMetadataComponent } from "./scanner-metadata/scanner-metadata.component";
import { NewScannerModalComponent } from "./new-scanner-modal/new-scanner-modal.component";
import { NewScannerFormComponent } from "./new-scanner-form/new-scanner-form.component";
import { TranslateService } from "@ngx-translate/core";
import { ErrorHandler } from "../../../lib/utils/error-handler";

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
    },
    updateScanner() {
      return of(true);
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
          // open auto detect
        { provide: ComponentFixtureAutoDetect, useValue: true }
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
