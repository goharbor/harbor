import { async, ComponentFixture, TestBed } from '@angular/core/testing';
import { BrowserAnimationsModule } from "@angular/platform-browser/animations";
import { ClarityModule } from "@clr/angular";
import { of } from "rxjs";
import { TranslateService } from "@ngx-translate/core";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { ErrorHandler } from "@harbor/ui";
import { ScannerComponent } from "./scanner.component";
import { ConfigScannerService } from "../../config/scanner/config-scanner.service";
import { SharedModule } from "../../shared/shared.module";
import { ActivatedRoute } from "@angular/router";

xdescribe('ScannerComponent', () => {
  let mockScanner1 = {
    name: 'test1',
    description: 'just a sample',
    version: '1.0.0',
    url: 'http://168.0.0.1'
  };
  let component: ScannerComponent;
  let fixture: ComponentFixture<ScannerComponent>;
  let fakedConfigScannerService = {
    getProjectScanner() {
      return of(mockScanner1);
    },
    getScanners() {
      return of([mockScanner1]);
    }
  };
  let fakedRoute = {
    snapshot: {
      parent: {
        params: {
          id: 1
        }
      }
    }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        BrowserAnimationsModule,
        ClarityModule,
      ],
      declarations: [ ScannerComponent ],
      providers: [
        TranslateService,
        MessageHandlerService,
        ErrorHandler,
        {provide: ActivatedRoute, useValue: fakedRoute},
        { provide: ConfigScannerService, useValue: fakedConfigScannerService },
      ]
    })
    .compileComponents();
  }));
  beforeEach(() => {
    fixture = TestBed.createComponent(ScannerComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });
  it('should creat', () => {
    expect(component).toBeTruthy();
  });
  it('should get scanner and render', () => {
    fixture.whenStable().then(() => {
      let el: HTMLElement = fixture.nativeElement.querySelector('#scanner-name');
      expect(el.textContent.trim).toEqual('test1');
    });
  });
  it('should get scanners and edit button is available', () => {
    fixture.whenStable().then(() => {
      let el: HTMLElement = fixture.nativeElement.querySelector('#edit-scanner');
      expect(el).toBeTruthy();
    });
  });
});
