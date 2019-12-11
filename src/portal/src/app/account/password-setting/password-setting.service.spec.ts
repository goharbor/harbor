import { TestBed, inject, getTestBed } from '@angular/core/testing';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { PasswordSettingService } from './password-setting.service';

describe('PasswordSettingService', () => {
  let injector: TestBed;
  let service: PasswordSettingService;
  let httpMock: HttpTestingController;
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        HttpClientTestingModule
      ],
      providers: [PasswordSettingService]
    });
    injector = getTestBed();
    service = injector.get(PasswordSettingService);
    httpMock = injector.get(HttpTestingController);
  });

  it('should be created', inject([PasswordSettingService], (service1: PasswordSettingService) => {
    expect(service1).toBeTruthy();
  }));

  const mockPasswordSetting = {
    old_password: 'string',
    new_password: 'string1'
  };

  it('changePassword() should success', () => {
    service.changePassword(1, mockPasswordSetting).subscribe((res) => {
      expect(res).toEqual(null);
    });

    const req = httpMock.expectOne('/api/users/1/password');
    expect(req.request.method).toBe('PUT');
    req.flush(null);
  });
  it('sendResetPasswordMail() should return data', () => {
    service.sendResetPasswordMail("123").subscribe((res) => {
      expect(res).toEqual(null);
    });

    const req = httpMock.expectOne('/c/sendEmail?email=123');
    expect(req.request.method).toBe('GET');
    req.flush(null);
  });
  it('resetPassword() should return data', () => {
    service.resetPassword('1234', 'Harbor12345').subscribe((res) => {
      expect(res).toEqual(null);
    });

    const req = httpMock.expectOne('/c/reset');
    expect(req.request.method).toBe('POST');
    req.flush(null);
  });
});
