import { AfterViewInit, Component, ElementRef, OnInit } from '@angular/core';
import { Http } from '@angular/http';
import { throwError as observableThrowError, Observable } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { Title } from '@angular/platform-browser';
import { TranslateService } from '@ngx-translate/core';
import { CookieService } from "ngx-cookie";

const SwaggerUI = require('swagger-ui');
@Component({
  selector: 'dev-center',
  templateUrl: 'dev-center.component.html',
  viewProviders: [Title],
  styleUrls: ['dev-center.component.scss']
})
export class DevCenterComponent implements AfterViewInit, OnInit {
  private ui: any;
  private host: any;
  private json: any;
  constructor(
    private el: ElementRef,
    private http: Http,
    private translate: TranslateService,
    private cookieService: CookieService,
    private titleService: Title) {
  }

  ngOnInit() {
    this.setTitle("APP_TITLE.HARBOR_SWAGGER");
  }

  public setTitle(key: string) {
    this.translate.get(key).subscribe((res: string) => {
      this.titleService.setTitle(res);
    });
  }

  ngAfterViewInit() {
    const csrfCookie = this.cookieService.get('_xsrf');
    const interceptor = {
      requestInterceptor: {
        apply: function (requestObj) {
          const headers = requestObj.headers || {};
          if (csrfCookie) {
            headers["X-Xsrftoken"] = atob(csrfCookie.split("|")[0]);
          }
          return requestObj;
        }
      }
    };
    this.http.get("/swagger.json")
    .pipe(catchError(error => observableThrowError(error)))
    .pipe(map(response => response.json())).subscribe(json => {
      json.host = window.location.host;
      const protocal = window.location.protocol;
      json.schemes = [protocal.replace(":", "")];
      let ui = SwaggerUI({
        spec: json,
        domNode: this.el.nativeElement.querySelector('.swagger-container'),
        deepLinking: true,
        presets: [
          SwaggerUI.presets.apis
        ],
        requestInterceptor: interceptor.requestInterceptor,
          authorizations: {
            csrf: function () {
              this.headers['X-Xsrftoken'] = csrfCookie;
              return true;
            }
          }
        });
      });
  }

}
