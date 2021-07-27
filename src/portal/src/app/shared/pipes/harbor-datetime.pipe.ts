import { Pipe, PipeTransform } from '@angular/core';
import { DatePipe } from "@angular/common";
import { DEFAULT_LANG_LOCALSTORAGE_KEY, DeFaultLang } from "../entities/shared.const";

const baseTimeLine: Date = new Date('1970-1-1');

@Pipe({
    name: 'harborDatetime',
    pure: false
})
export class HarborDatetimePipe implements PipeTransform {

    transform(value: any, format?: string): string {
        let lang: string = DeFaultLang;
        if (localStorage && localStorage.getItem(DEFAULT_LANG_LOCALSTORAGE_KEY)) {
            lang = localStorage.getItem(DEFAULT_LANG_LOCALSTORAGE_KEY);
        }
        if (value && value <= baseTimeLine) {// invalid date
            return '-';
        }
        // default format medium
        return new DatePipe(lang).transform(value, format ? format : 'medium');
    }
}
