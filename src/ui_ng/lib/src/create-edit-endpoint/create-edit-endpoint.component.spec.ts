import { ComponentFixture, TestBed, async, fakeAsync, tick } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';

import { TranslateService } from '@ngx-translate/core';

import { SharedModule } from '../shared/shared.module';

import { FilterComponent } from '../filter/filter.component';

import { CreateEditEndpointComponent } from '../create-edit-endpoint/create-edit-endpoint.component';
import { InlineAlertComponent } from '../inline-alert/inline-alert.component';
import { ErrorHandler } from '../error-handler/error-handler';
import { Endpoint } from '../service/interface';
import { EndpointService, EndpointDefaultService } from '../service/endpoint.service';
import { IServiceConfig, SERVICE_CONFIG } from '../service.config';
describe('CreateEditEndpointComponent (inline template)', () => {

  let mockData: Endpoint = {
    "id": 1,
    "endpoint": "https://10.117.4.151",
    "name": "target_01",
    "username": "admin",
    "password": "",
    "type": 0
  };

  let comp: CreateEditEndpointComponent;
  let fixture: ComponentFixture<CreateEditEndpointComponent>;
  let de: DebugElement;
  let el: HTMLElement;

  let config: IServiceConfig = {
    systemInfoEndpoint: '/api/endpoints/testing'
  };

  let endpointService: EndpointService;
  
  let spy: jasmine.Spy;
  
  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [ SharedModule ],
      declarations: [ 
          FilterComponent, 
          CreateEditEndpointComponent, 
          InlineAlertComponent ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: EndpointService, useClass: EndpointDefaultService },
        { provide: TranslateService, useClass: TranslateService}
      ]
    });
  }));

  beforeEach(()=>{
    fixture = TestBed.createComponent(CreateEditEndpointComponent);
    comp = fixture.componentInstance;
    
    endpointService = fixture.debugElement.injector.get(EndpointService);
    spy = spyOn(endpointService, 'getEndpoint').and.returnValue(Promise.resolve(mockData));
    fixture.detectChanges();
  });

  it('should be created', () => {
    fixture.detectChanges();
    expect(comp).toBeTruthy();
  });

  it('should get endpoint be called', async(()=>{
    fixture.detectChanges();
    comp.openCreateEditTarget(true, 1);
    comp.createEditDestinationOpened = false;
    fixture.whenStable().then(()=>{
      fixture.detectChanges(); 
      expect(spy.calls.any()).toBeTruthy();
    });
  }));

  it('should get endpoint to open modal', async(()=>{
    fixture.detectChanges();
    comp.openCreateEditTarget(true, 1);
    comp.createEditDestinationOpened = false;
    fixture.whenStable().then(()=>{
      fixture.detectChanges();  
      expect(comp.target.name).toEqual('target_01');
    });
  }));

  it('should endpoint be initialized', () => {
    fixture.detectChanges();
    expect(config.systemInfoEndpoint).toEqual('/api/endpoints/testing');
  });

});