import { Pipe, PipeTransform } from '@angular/core';
import { DatePipe } from "@angular/common";
import { DEFAULT_LANG_LOCALSTORAGE_KEY, DeFaultLang } from "../entities/shared.const";

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
    // default format medium
    return new DatePipe(lang).transform(value, format ? format : 'medium');
  }
}
