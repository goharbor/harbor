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
import { Scanner } from "../../config/scanner/scanner";

describe('ScannerComponent', () => {
  const mockScanner1: Scanner = {
    uuid: 'abc',
    name: 'test1',
    description: 'just a sample',
    version: '1.0.0',
    url: 'http://168.0.0.1',
    health: 'healthy'
  };
  const mockScanner2: Scanner = {
    uuid: 'def',
    name: 'test2',
    description: 'just a sample',
    version: '2.0.0',
    url: 'http://168.0.0.2',
    health: 'healthy'
  };
  let component: ScannerComponent;
  let fixture: ComponentFixture<ScannerComponent>;
  let fakedConfigScannerService = {
    getProjectScanner() {
      return of(mockScanner1);
    },
    getScanners() {
      return of([mockScanner1, mockScanner2]);
    },
    getProjectScanners() {
      return of([mockScanner1, mockScanner2]);
    },
    updateProjectScanner() {
      return of(true);
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
    spyOn(component, 'getPermission').and.returnValue(undefined);
    fixture.detectChanges();
  });
  it('should creat', () => {
    expect(component).toBeTruthy();
  });
  it('should get scanner and render', () => {
    component.hasCreatePermission = true;
    let el: HTMLElement = fixture.nativeElement.querySelector('#scanner-name');
    expect(el.textContent.trim()).toEqual('test1');
  });
  it('select another scanner', () => {
    component.hasCreatePermission = true;
    component.getScanners();
    fixture.detectChanges();
    const editButton = fixture.nativeElement.querySelector('#edit-scanner');
    expect(editButton).toBeTruthy();
    editButton.click();
    fixture.detectChanges();
    component.selectedScanner = mockScanner2;
    fixture.detectChanges();
    const saveButton = fixture.nativeElement.querySelector('#save-scanner');
    saveButton.click();
    fixture.detectChanges();
    expect(component.opened).toBeFalsy();
  });
});
