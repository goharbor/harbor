import { async, ComponentFixture, ComponentFixtureAutoDetect, TestBed } from '@angular/core/testing';
import { ClrLoadingState } from "@clr/angular";
import { ConfigScannerService } from "../config-scanner.service";
import { NewScannerModalComponent } from "./new-scanner-modal.component";
import { MessageHandlerService } from "../../../shared/message-handler/message-handler.service";
import { NewScannerFormComponent } from "../new-scanner-form/new-scanner-form.component";
import { FormBuilder } from "@angular/forms";
import { of, Subscription } from "rxjs";
import { delay } from "rxjs/operators";
import { SharedModule } from "@harbor/ui";
import { SharedModule as AppSharedModule } from "../../../shared/shared.module";

describe('NewScannerModalComponent', () => {
  let component: NewScannerModalComponent;
  let fixture: ComponentFixture<NewScannerModalComponent>;

  let mockScanner1 = {
    name: 'test1',
    description: 'just a sample',
    url: 'http://168.0.0.1',
    auth: "",
  };
  let fakedConfigScannerService = {
    getScannersByName() {
      return of([mockScanner1]);
    },
    testEndpointUrl() {
      return of(true).pipe(delay(200));
    },
    addScanner() {
      return of(true).pipe(delay(200));
    },
    updateScanner() {
      return of(true).pipe(delay(200));
    }
  };
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule,
        AppSharedModule
      ],
      declarations: [
        NewScannerFormComponent,
        NewScannerModalComponent,
      ],
      providers: [
        MessageHandlerService,
        { provide: ConfigScannerService, useValue: fakedConfigScannerService },
        FormBuilder,
        // open auto detect
        { provide: ComponentFixtureAutoDetect, useValue: true }
      ]
    })
    .compileComponents();
  }));
  beforeEach(() => {
    fixture = TestBed.createComponent(NewScannerModalComponent);
    component = fixture.componentInstance;
    component.opened = true;
    component.newScannerFormComponent.checkNameSubscribe = new Subscription();
    component.newScannerFormComponent.checkEndpointUrlSubscribe = new Subscription();
    fixture.detectChanges();
  });
  it('should creat', () => {
    expect(component).toBeTruthy();
  });
  it('should be add mode', () => {
    component.isEdit = false;
    fixture.detectChanges();
    let el = fixture.nativeElement.querySelector('#button-add');
    expect(el).toBeTruthy();
  });
  it('should be edit mode', () => {
    component.isEdit = true;
    fixture.detectChanges();
    let el = fixture.nativeElement.querySelector('#button-save');
    expect(el).toBeTruthy();
    // set origin value
    component.originValue = mockScanner1;
    component.editScanner = {};
    // input same value to origin
    fixture.nativeElement.querySelector('#scanner-name').value = "test2";
    fixture.nativeElement.querySelector('#description').value = "just a sample";
    fixture.nativeElement.querySelector('#scanner-endpoint').value = "http://168.0.0.1";
    fixture.nativeElement.querySelector('#scanner-authorization').value = "";
    fixture.nativeElement.querySelector('#scanner-name').dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('#description').dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('#scanner-endpoint').dispatchEvent(new Event('input'));
    fixture.nativeElement.querySelector('#scanner-authorization').dispatchEvent(new Event('input'));
    // save button should not be disabled
    expect(component.validForSaving).toBeTruthy();
    fixture.nativeElement.querySelector('#scanner-name').value = "test3";
    fixture.nativeElement.querySelector('#scanner-name').dispatchEvent(new Event('input'));
    fixture.detectChanges();
    expect(component.validForSaving).toBeTruthy();
    el.click();
    el.dispatchEvent(new Event('click'));
    setTimeout(() => {
        expect(component.opened).toBeFalsy();
    }, 300);
  });
  it('test connection button should not be disabled', () => {
    let nameInput = fixture.nativeElement.querySelector('#scanner-name');
    nameInput.value = "test2";
    nameInput.dispatchEvent(new Event('input'));
    let urlInput = fixture.nativeElement.querySelector('#scanner-endpoint');
    urlInput.value = "http://168.0.0.1";
    urlInput.dispatchEvent(new Event('input'));
    expect(component.canTestEndpoint).toBeTruthy();
    let el = fixture.nativeElement.querySelector('#button-test');
    el.click();
    el.dispatchEvent(new Event('click'));
    expect(component.checkBtnState).toBe(ClrLoadingState.LOADING);
    setTimeout(() => {
      expect(component.checkBtnState).toBe(ClrLoadingState.SUCCESS);
    }, 300);
  });
  it('add button should not be disabled', () => {
    fixture.nativeElement.querySelector('#scanner-name').value = "test2";
    fixture.nativeElement.querySelector('#scanner-endpoint').value = "http://168.0.0.1";
    let authInput = fixture.nativeElement.querySelector('#scanner-authorization');
    authInput.value = "Basic";
    authInput.dispatchEvent(new Event('change'));
    let usernameInput = fixture.nativeElement.querySelector('#scanner-username');
    let passwordInput = fixture.nativeElement.querySelector('#scanner-password');
    expect(usernameInput).toBeTruthy();
    expect(passwordInput).toBeTruthy();
    usernameInput.value = "user";
    passwordInput.value = "12345";
    usernameInput.dispatchEvent(new Event('input'));
    passwordInput.dispatchEvent(new Event('input'));
    let el = fixture.nativeElement.querySelector('#button-add');
    expect(component.valid).toBeFalsy();
    el.click();
    el.dispatchEvent(new Event('click'));
    setTimeout(() => {
      expect(component.opened).toBeFalsy();
    }, 300);
  });
});



