import { ComponentFixture, TestBed, async } from '@angular/core/testing';

import { SharedModule } from '../shared/shared.module';
import { ErrorHandler } from '../error-handler/error-handler';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';

import { SystemSettingsComponent } from './system/system-settings.component';
import { VulnerabilityConfigComponent } from './vulnerability/vulnerability-config.component';
import { RegistryConfigComponent } from './registry-config.component';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';

import {
  ConfigurationService,
  ConfigurationDefaultService,
  ScanningResultService,
  ScanningResultDefaultService,
  SystemInfoService,
  SystemInfoDefaultService,
  SystemInfo
} from '../service/index';
import { Configuration } from './config';

describe('RegistryConfigComponent (inline template)', () => {

  let comp: RegistryConfigComponent;
  let fixture: ComponentFixture<RegistryConfigComponent>;
  let cfgService: ConfigurationService;
  let systemInfoService: SystemInfoService;
  let spy: jasmine.Spy;
  let saveSpy: jasmine.Spy;
  let spySystemInfo: jasmine.Spy;
  let mockConfig: Configuration = new Configuration();
  mockConfig.token_expiration.value = 90;
  mockConfig.scan_all_policy.value = {
    type: "daily",
    parameter: {
      daily_time: 0
    }
  };
  let config: IServiceConfig = {
    configurationEndpoint: '/api/configurations/testing'
  };
  let mockSystemInfo: SystemInfo = {
    "with_notary": true,
    "with_admiral": false,
    "with_clair": true,
    "admiral_endpoint": "NA",
    "auth_mode": "db_auth",
    "registry_url": "10.112.122.56",
    "project_creation_restriction": "everyone",
    "self_registration": true,
    "has_ca_root": true,
    "harbor_version": "v1.1.1-rc1-160-g565110d",
    "next_scan_all": 0
  };

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      declarations: [
        SystemSettingsComponent,
        VulnerabilityConfigComponent,
        RegistryConfigComponent,
        ConfirmationDialogComponent
      ],
      providers: [
        ErrorHandler,
        { provide: SERVICE_CONFIG, useValue: config },
        { provide: ConfigurationService, useClass: ConfigurationDefaultService },
        { provide: ScanningResultService, useClass: ScanningResultDefaultService },
        { provide: SystemInfoService, useClass: SystemInfoDefaultService }
      ]
    });
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(RegistryConfigComponent);
    comp = fixture.componentInstance;

    cfgService = fixture.debugElement.injector.get(ConfigurationService);
    systemInfoService = fixture.debugElement.injector.get(SystemInfoService);
    spy = spyOn(cfgService, 'getConfigurations').and.returnValue(Promise.resolve(mockConfig));
    saveSpy = spyOn(cfgService, 'saveConfigurations').and.returnValue(Promise.resolve(true));
    spySystemInfo = spyOn(systemInfoService, 'getSystemInfo').and.returnValue(Promise.resolve(mockSystemInfo));

    fixture.detectChanges();
  });

  it('should render configurations to the view', async(() => {
    expect(spy.calls.count()).toEqual(1);
    expect(spySystemInfo.calls.count()).toEqual(1);
    fixture.detectChanges();

    fixture.whenStable().then(() => {
      fixture.detectChanges();

      let el: HTMLInputElement = fixture.nativeElement.querySelector('input[type="text"]');
      expect(el).not.toBeFalsy();
      expect(el.value).toEqual('30');


      fixture.detectChanges();
      let el3: HTMLInputElement = fixture.nativeElement.querySelector('.btn-scan');
      expect(el3).toBeTruthy();
      expect(el3).not.toBeFalsy();
    });
  }));

  it('should save the configuration changes', async(() => {
    comp.save();
    fixture.detectChanges();

    expect(saveSpy.calls.any).toBeTruthy();
  }));
});
