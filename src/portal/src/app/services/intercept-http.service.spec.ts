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
      if (request.headers.has('X-Harbor-CSRF-Token')) {
        return of(new HttpResponse({status: 200}));
      } else {
        return throwError(new HttpResponse( {
        status: 403
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
        if (res.status === 403) {
          expect(mockRequest.headers.get("X-Harbor-CSRF-Token")).toEqual(cookie);
        } else {
          expect(res.status).toEqual(200);
        }
      });
    }));
});
