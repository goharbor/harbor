import { AfterViewInit, Component, ElementRef, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { throwError as observableThrowError, forkJoin } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { Title } from '@angular/platform-browser';
import { TranslateService } from '@ngx-translate/core';
import { CookieService } from 'ngx-cookie';
import SwaggerUI from 'swagger-ui';
import { mergeDeep } from '../shared/units/utils';
import { DevCenterBaseDirective } from './dev-center-base';
import { SAFE_METHODS } from '../services/intercept-http.service';

enum SwaggerJsonUrls {
    SWAGGER1 = '/swagger.json',
    SWAGGER2 = '/swagger2.json',
}

const helpInfo: string =
    ' If you want to enable basic authorization,' +
    ' please logout Harbor first or manually delete the cookies under the current domain.';

@Component({
    selector: 'dev-center',
    templateUrl: 'dev-center.component.html',
    viewProviders: [Title],
    styleUrls: ['dev-center.component.scss'],
})
export class DevCenterComponent
    extends DevCenterBaseDirective
    implements AfterViewInit, OnInit
{
    private ui: any;
    constructor(
        private el: ElementRef,
        private http: HttpClient,
        public translate: TranslateService,
        public cookieService: CookieService,
        public titleService: Title
    ) {
        super(translate, cookieService, titleService);
    }

    ngAfterViewInit() {
        this.getSwaggerUI();
    }
    getSwaggerUI() {
        forkJoin([
            this.http.get(SwaggerJsonUrls.SWAGGER1),
            this.http.get(SwaggerJsonUrls.SWAGGER2),
        ])
            .pipe(catchError(error => observableThrowError(error)))
            .subscribe(jsonArr => {
                const json: any = {};
                mergeDeep(json, jsonArr[0], jsonArr[1]);
                json['host'] = window.location.host;
                const protocal = window.location.protocol;
                json['schemes'] = [protocal.replace(':', '')];
                json.info.description = json.info.description + helpInfo;
                this.ui = SwaggerUI({
                    spec: json,
                    domNode:
                        this.el.nativeElement.querySelector(
                            '.swagger-container'
                        ),
                    deepLinking: true,
                    presets: [SwaggerUI.presets.apis],
                    requestInterceptor: request => {
                        // Get the csrf token from localstorage
                        const token = localStorage.getItem('__csrf');
                        const headers = request.headers || {};
                        if (token) {
                            if (
                                request.method &&
                                SAFE_METHODS.indexOf(
                                    request.method.toUpperCase()
                                ) === -1
                            ) {
                                headers['X-Harbor-CSRF-Token'] = token;
                            }
                        }
                        return request;
                    },
                    responseInterceptor: response => {
                        const headers = response.headers || {};
                        const responseToken: string =
                            headers['X-Harbor-CSRF-Token'];
                        if (responseToken) {
                            // Set the csrf token to localstorage
                            localStorage.setItem('__csrf', responseToken);
                        }
                        return response;
                    },
                });
            });
    }
}
