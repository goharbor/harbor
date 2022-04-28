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
        if (value && value <= baseTimeLine) {// invalid date
            return '-';
        }
        const langFromLocalStorage = localStorage && localStorage.getItem(DEFAULT_LANG_LOCALSTORAGE_KEY);
        const lang = langFromLocalStorage || DeFaultLang;
        // default format medium
        return new DatePipe(lang).transform(value, format ? format : 'medium');
    }
}
