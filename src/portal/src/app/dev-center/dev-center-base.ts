import { AfterViewInit, OnInit } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { TranslateService } from '@ngx-translate/core';
import { CookieService } from "ngx-cookie";

export  abstract class DevCenterBase implements OnInit, AfterViewInit {
  constructor(
    public translate: TranslateService,
    public cookieService: CookieService,
    public titleService: Title) {
  }

  ngOnInit() {
    this.setTitle("APP_TITLE.HARBOR_SWAGGER");
  }

  private setTitle(key: string) {
    this.translate.get(key).subscribe((res: string) => {
      this.titleService.setTitle(res);
    });
  }
  public getCsrfInterceptor() {
    return {
        requestInterceptor: {
          apply: (requestObj) => {
            const csrfCookie = this.cookieService.get('__csrf');
            const headers = requestObj.headers || {};
            if (csrfCookie) {
              headers["X-Harbor-CSRF-Token"] = csrfCookie;
            }
            return requestObj;
          }
        }
      };
  }
  abstract getSwaggerUI();
  abstract ngAfterViewInit();
}
