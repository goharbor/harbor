import { AfterViewInit, Component, Directive, OnInit } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { TranslateService } from '@ngx-translate/core';
import { CookieService } from 'ngx-cookie';

@Directive()
export abstract class DevCenterBaseDirective implements OnInit, AfterViewInit {
    protected constructor(
        public translate: TranslateService,
        public cookieService: CookieService,
        public titleService: Title
    ) {}

    ngOnInit() {
        this.setTitle('APP_TITLE.HARBOR_SWAGGER');
    }

    private setTitle(key: string) {
        this.translate.get(key).subscribe((res: string) => {
            this.titleService.setTitle(res);
        });
    }
    abstract getSwaggerUI();
    abstract ngAfterViewInit();
}
