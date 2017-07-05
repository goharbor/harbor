import { ComponentFixture, TestBed, async } from '@angular/core/testing';

import { SharedModule } from '../shared/shared.module';
import { ErrorHandler } from '../error-handler/error-handler';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';

import { ReplicationConfigComponent } from './replication/replication-config.component';
import { SystemSettingsComponent } from './system/system-settings.component';
import { VulnerabilityConfigComponent } from './vulnerability/vulnerability-config.component';
import { RegistryConfigComponent } from './registry-config.component';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';

import { 
  ConfigurationService, 
  ConfigurationDefaultService,
  ScanningResultService,
  ScanningResultDefaultService
 } from '../service/index';
import { Configuration } from './config';

describe('RegistryConfigComponent (inline template)', () => {

  let comp: RegistryConfigComponent;
  let fixture: ComponentFixture<RegistryConfigComponent>;
  let cfgService: ConfigurationService;
  let spy: jasmine.Spy;
  let saveSpy: jasmine.Spy;
  let mockConfig: Configuration = new Configuration();
  mockConfig.token_expiration.value = 90;
  mockConfig.verify_remote_cert.value = true;
  mockConfig.scan_all_policy.value = {
    type: "daily",
    parameter: {
      daily_time: 0
    }
  };
  let config: IServiceConfig = {
    configurationEndpoint: '/api/configurations/testing'
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [
        ReplicationConfigComponent,
        SystemSettingsComponent,
        VulnerabilityConfigComponent,
        RegistryConfigComponent,
        ConfirmationDialogComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ConfigurationService, useClass: ConfigurationDefaultService },
        { provide: ScanningResultService, useClass: ScanningResultDefaultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RegistryConfigComponent);
    comp = fixture.componentInstance;

    cfgService = fixture.debugElement.injector.get(ConfigurationService);
    spy = spyOn(cfgService, 'getConfigurations').and.returnValue(Promise.resolve(mockConfig));
    saveSpy = spyOn(cfgService, 'saveConfigurations').and.returnValue(Promise.resolve(true));

    fixture.detectChanges();
  });

  it('should render configurations to the view', async(() => {
    expect(spy.calls.count()).toEqual(1);
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let el: HTMLInputElement = fixture.nativeElement.querySelector('input[type="text"]');
      expect(el).toBeTruthy();
      expect(el.value).toEqual('30');

      let el2: HTMLInputElement = fixture.nativeElement.querySelector('input[type="checkbox"]');
      expect(el2).toBeTruthy();
      expect(el2.value).toEqual('on');

      let el3: HTMLInputElement = fixture.nativeElement.querySelector('input[type="time"]');
      expect(el3).toBeTruthy();
      expect(el3.value).toBeTruthy();
    });
  }));

  it('should save the configuration changes', async(() => {
    comp.save();
    fixture.detectChanges();

    expect(saveSpy.calls.any).toBeTruthy();
  }));
});