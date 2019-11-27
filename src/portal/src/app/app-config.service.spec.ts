import { TestBed, inject } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { CookieService } from 'ngx-cookie';
import { AppConfigService } from './app-config.service';

describe('AppConfigService', () => {
  let fakeCookieService = {
    get: function (key) {
      return key;
    }
  };

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [HttpClientTestingModule],
      providers: [AppConfigService,
        { provide: CookieService, useValue: fakeCookieService }]
    });
  });

  it('should be created', inject([AppConfigService], (service: AppConfigService) => {
    expect(service).toBeTruthy();
  }));
});
