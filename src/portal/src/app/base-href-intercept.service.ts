import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpHandler, HttpRequest } from '@angular/common/http';
import { Observable } from "rxjs";

const BASE_HREF = '/api';
enum APILevels {
  'V1.0' = '',
  'V2.0' = '/v2.0'
}
@Injectable()
export class BaseHrefInterceptService implements HttpInterceptor {
  intercept(req: HttpRequest<any>, next: HttpHandler): Observable<any> {
    let url: string = req.url;
    // use API level v2.0
    if (url && url.indexOf(BASE_HREF) !== -1 && url.indexOf(BASE_HREF + APILevels["V2.0"]) === -1) {
      url = BASE_HREF + APILevels["V2.0"] + url.split(BASE_HREF)[1];
    }
    const apiReq = req.clone({url});
    return next.handle(apiReq);
  }
}

