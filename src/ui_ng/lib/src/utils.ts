import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/toPromise';
import { RequestOptions, Headers } from '@angular/http';
import { RequestQueryParams } from './service/RequestQueryParams';
import { DebugElement } from '@angular/core';
import { Comparator } from 'clarity-angular';

/**
 * Convert the different async channels to the Promise<T> type.
 * 
 * @export
 * @template T
 * @param {(Observable<T> | Promise<T> | T)} async
 * @returns {Promise<T>}
 */
export function toPromise<T>(async: Observable<T> | Promise<T> | T): Promise<T> {
    if (!async) {
        return Promise.reject("Bad argument");
    }

    if (async instanceof Observable) {
        let obs: Observable<T> = async;
        return obs.toPromise();
    } else {
        return Promise.resolve(async);
    }
}

/**
 * The default cookie key used to store current used language preference.
 */
export const DEFAULT_LANG_COOKIE_KEY = 'harbor-lang';

/**
 * Declare what languages are supported now.
 */
export const DEFAULT_SUPPORTING_LANGS = ['en-us', 'zh-cn', 'es-es'];

/**
 * The default language.
 */
export const DEFAULT_LANG = 'en-us';

export const HTTP_JSON_OPTIONS: RequestOptions = new RequestOptions({
    headers: new Headers({
        "Content-Type": 'application/json',
        "Accept": 'application/json'
    })
});

/**
 * Build http request options
 * 
 * @export
 * @param {RequestQueryParams} params
 * @returns {RequestOptions}
 */
export function buildHttpRequestOptions(params: RequestQueryParams): RequestOptions {
    let reqOptions: RequestOptions = new RequestOptions({
        headers: new Headers({
            "Content-Type": 'application/json',
            "Accept": 'application/json'
        })
    });

    if (params) {
        reqOptions.search = params;
    }

    return reqOptions;
}



/** Button events to pass to `DebugElement.triggerEventHandler` for RouterLink event handler */
export const ButtonClickEvents = {
   left:  { button: 0 },
   right: { button: 2 }
};


/** Simulate element click. Defaults to mouse left-button click event. */
export function click(el: DebugElement | HTMLElement, eventObj: any = ButtonClickEvents.left): void {
  if (el instanceof HTMLElement) {
    el.click();
  } else {
    el.triggerEventHandler('click', eventObj);
  }
}

/**
 * Comparator for fields with specific type.
 *  
 */
export class CustomComparator<T> implements Comparator<T> {

  fieldName: string;
  type: string;

  constructor(fieldName: string, type: string) {
    this.fieldName = fieldName;
    this.type = type;
  }
  
  compare(a: {[key: string]: any| any[]}, b: {[key: string]: any| any[]}) {
    let comp = 0;
    if(a && b) {
      let fieldA = a[this.fieldName];
      let fieldB = b[this.fieldName];
      switch(this.type) {
      case "number": 
        comp = fieldB - fieldA;
        break;
      case "date":
        comp = new Date(fieldB).getTime() - new Date(fieldA).getTime();
        break;
      }
    }
    return comp;
  }
}