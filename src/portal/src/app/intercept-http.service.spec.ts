import { TestBed, inject } from '@angular/core/testing';

import { InterceptHttpService } from './intercept-http.service';
import { CookieService } from 'ngx-cookie';
import { HttpRequest, HttpResponse } from '@angular/common/http';
import { of, throwError } from 'rxjs';

describe('InterceptHttpService', () => {
  let cookie = "fdsa|ds";
  const mockCookieService = {
    get: function () {
      return cookie;
    },
    set: function (cookieStr: string) {
      cookie = cookieStr;
    }
  };
  const mockRequest = new HttpRequest('PUT', "", {
    headers: new Map()
  });
  const mockHandle = {
    handle: (request) => {
      if (request.headers.has('X-Xsrftoken')) {
        return of(new HttpResponse({status: 200}));
      } else {
        return throwError(new HttpResponse( {
        status: 422
        }));
      }
    }
  };
  beforeEach(() => TestBed.configureTestingModule({}));
  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [],
      providers: [
        InterceptHttpService,
        { provide: CookieService, useValue: mockCookieService }
      ]
    });

  });
  it('should be initialized', inject([InterceptHttpService], (service: InterceptHttpService) => {
    expect(service).toBeTruthy();
  }));

  it('should be get right token and send right request when the cookie not exists', inject([InterceptHttpService],
    (service: InterceptHttpService) => {
      mockCookieService.set("fdsa|ds");
      service.intercept(mockRequest, mockHandle).subscribe(res => {
        if (res.status === 422) {
          expect(btoa(mockRequest.headers.get("X-Xsrftoken"))).toEqual(cookie.split("|")[0]);
        } else {
          expect(res.status).toEqual(200);
        }
      });
    }));
});
