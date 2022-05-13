import { Inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { map, catchError } from 'rxjs/operators';
import { Observable, throwError as observableThrowError } from 'rxjs';
import { CustomStyle } from './theme';
import { DOCUMENT } from '@angular/common';
import { environment } from 'src/environments/environment';
@Injectable()
export class SkinableConfig {
    private customSkinData: CustomStyle;
    constructor(
        private http: HttpClient,
        @Inject(DOCUMENT) private document: Document
    ) {}

    public getCustomFile(): Observable<any> {
        return this.http
            .get(`setting.json?buildTimeStamp=${environment?.buildTimestamp}`)
            .pipe(
                map(
                    response => (this.customSkinData = response as CustomStyle)
                ),
                catchError((error: any) => {
                    console.error('custom skin json file load failed');
                    return observableThrowError(error);
                })
            );
    }

    public getSkinConfig() {
        return this.customSkinData;
    }

    public setTitleIcon() {
        if (
            this.customSkinData &&
            this.customSkinData.product &&
            this.customSkinData.product.logo
        ) {
            const titleIcon: HTMLLinkElement =
                this.document.querySelector('link');
            titleIcon.href = `images/${this.customSkinData.product.logo}`;
        }
    }
}
