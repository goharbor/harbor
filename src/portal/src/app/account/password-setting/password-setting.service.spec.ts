import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { PasswordSettingService } from './password-setting.service';

describe('PasswordSettingService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule
      ],
      providers: [PasswordSettingService]
    });
  });

  it('should be created', inject([PasswordSettingService], (service: PasswordSettingService) => {
    expect(service).toBeTruthy();
  }));
});
