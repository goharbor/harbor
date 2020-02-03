import { TestBed, inject } from '@angular/core/testing';

import { HttpXsrfTokenExtractorToBeUsed } from './http-xsrf-token-extractor.service';
import { SharedModule } from '../utils/shared/shared.module';
import { CookieService } from "ngx-cookie";

describe('HttpXsrfTokenExtractorToBeUsed', () => {
  let cookie =  "fdsa|ds";
  let mockCookieService =  {
      get: function () {
          return cookie;
      },
      set: function (cookieStr: string) {
          cookie = cookieStr;
      }
  };
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [
        SharedModule
      ],
      providers: [
        HttpXsrfTokenExtractorToBeUsed,
        { provide: CookieService, useValue: mockCookieService}
    ]
    });

  });

  it('should be initialized', inject([HttpXsrfTokenExtractorToBeUsed], (service: HttpXsrfTokenExtractorToBeUsed) => {
    expect(service).toBeTruthy();
  }));

  it('should be get right token when the cookie exists', inject([HttpXsrfTokenExtractorToBeUsed],
    (service: HttpXsrfTokenExtractorToBeUsed) => {
    mockCookieService.set("fdsa|ds");
    let token = service.getToken();
    expect(btoa(token)).toEqual(cookie.split("|")[0]);
  }));

  it('should be get right token when the cookie does not exist', inject([HttpXsrfTokenExtractorToBeUsed],
    (service: HttpXsrfTokenExtractorToBeUsed) => {
    mockCookieService.set(null);
    let token = service.getToken();
    expect(token).toBeNull();
  }));


});
