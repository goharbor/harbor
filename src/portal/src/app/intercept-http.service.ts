import { Injectable } from '@angular/core';
import { HttpInterceptor, HttpRequest, HttpHandler, HttpEvent, HttpResponse } from '@angular/common/http';
import { Observable, throwError } from 'rxjs';
import { tap, catchError } from 'rxjs/operators';
import { CookieService } from 'ngx-cookie';

@Injectable({
  providedIn: 'root'
})
export class InterceptHttpService implements HttpInterceptor {

  constructor(private cookie: CookieService) { }

  intercept(request: HttpRequest<any>, next: HttpHandler): Observable<any> {

    return next.handle(request).pipe(catchError(error => {
      if (error.status === 403) {
        let Xsrftoken = this.cookie.get("__csrf");
        if (Xsrftoken && !request.headers.has('X-Harbor-CSRF-Token')) {
          request = request.clone({ headers: request.headers.set('X-Harbor-CSRF-Token', Xsrftoken) });
          return next.handle(request);
        }
      }
      return throwError(error);
    }));
  }
}

