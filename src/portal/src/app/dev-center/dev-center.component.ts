import { AfterViewInit, Component, ElementRef, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { throwError as observableThrowError, forkJoin } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { Title } from '@angular/platform-browser';
import { TranslateService } from '@ngx-translate/core';
import { CookieService } from "ngx-cookie";
import * as SwaggerUI from 'swagger-ui';
import { mergeDeep } from "../../lib/utils/utils";

enum SwaggerJsonUrls {
  SWAGGER1 = '/swagger.json',
  SWAGGER2 = '/swagger2.json'
}

@Component({
  selector: 'dev-center',
  templateUrl: 'dev-center.component.html',
  viewProviders: [Title],
  styleUrls: ['dev-center.component.scss']
})
export class DevCenterComponent implements AfterViewInit, OnInit {
  private ui: any;
  constructor(
    private el: ElementRef,
    private http: HttpClient,
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
    forkJoin([this.http.get(SwaggerJsonUrls.SWAGGER1), this.http.get(SwaggerJsonUrls.SWAGGER2)])
      .pipe(catchError(error => observableThrowError(error)))
      .subscribe(jsonArr => {
        const json: object = {};
        mergeDeep(json, jsonArr[0], jsonArr[1]);
        json['host'] = window.location.host;
        const protocal = window.location.protocol;
        json['schemes'] = [protocal.replace(":", "")];
        this.ui = SwaggerUI({
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
