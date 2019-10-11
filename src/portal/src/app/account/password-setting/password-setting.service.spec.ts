import { TestBed, inject } from '@angular/core/testing';

import { PasswordSettingService } from './password-setting.service';

xdescribe('PasswordSettingService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [PasswordSettingService]
    });
  });

  it('should be created', inject([PasswordSettingService], (service: PasswordSettingService) => {
    expect(service).toBeTruthy();
  }));
});
