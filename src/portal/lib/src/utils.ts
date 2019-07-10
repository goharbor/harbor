import { Observable } from "rxjs";

import { HttpHeaders, HttpParams } from '@angular/common/http';
import { RequestQueryParams } from './service/RequestQueryParams';
import { DebugElement } from '@angular/core';
import { Comparator, State, HttpOptionInterface, HttpOptionTextInterface, QuotaUnitInterface } from './service/interface';
import { QuotaHardInterface } from './service/interface';
import { QuotaUnits } from './shared/shared.const';
/**
 * Convert the different async channels to the Promise<T> type.
 *
 **
 * template T
 *  ** deprecated param {(Observable<T> | Promise<T> | T)} async
 * returns {Promise<T>}
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
export const DEFAULT_SUPPORTING_LANGS = ['en-us', 'zh-cn', 'es-es', 'fr-fr', 'pt-br'];

/**
 * The default language.
 */
export const DEFAULT_LANG = 'en-us';


export const HTTP_JSON_OPTIONS: HttpOptionInterface = {
    headers: new HttpHeaders({
        "Content-Type": 'application/json',
        "Accept": 'application/json'
    }),
    responseType: 'json'
};

export const HTTP_GET_OPTIONS: HttpOptionInterface = {
    headers: new HttpHeaders({
        "Content-Type": 'application/json',
        "Accept": 'application/json',
        "Cache-Control": 'no-cache',
        "Pragma": 'no-cache'
    }),
    responseType: 'json'
};
export const HTTP_GET_OPTIONS_OBSERVE_RESPONSE: HttpOptionInterface = {
    headers: new HttpHeaders({
        "Content-Type": 'application/json',
        "Accept": 'application/json',
        "Cache-Control": 'no-cache',
        "Pragma": 'no-cache'
    }),
    observe: 'response' as 'body',
    responseType: 'json'
};
export const HTTP_GET_OPTIONS_TEXT: HttpOptionTextInterface = {
    headers: new HttpHeaders({
        "Content-Type": 'application/json',
        "Accept": 'application/json',
        "Cache-Control": 'no-cache',
        "Pragma": 'no-cache'
    }),
    responseType: 'text'
};

export const HTTP_FORM_OPTIONS: HttpOptionInterface = {
    headers: new HttpHeaders({
        "Content-Type": 'application/x-www-form-urlencoded'
    }),
    responseType: 'json'
};

export const HTTP_GET_HEADER: HttpHeaders = new HttpHeaders({
    "Content-Type": 'application/json',
    "Accept": 'application/json',
    "Cache-Control": 'no-cache',
    "Pragma": 'no-cache'
});

export const HTTP_GET_OPTIONS_CACHE: HttpOptionInterface = {
    headers: new HttpHeaders({
        "Content-Type": 'application/json',
        "Accept": 'application/json',
        "Cache-Control": 'no-cache',
        "Pragma": 'no-cache',
    }),
    responseType: 'json'
};

export const FILE_UPLOAD_OPTION: HttpOptionInterface = {
    headers: new HttpHeaders({
        "Content-Type": 'multipart/form-data',
    }),
    responseType: 'json'
};

/**
 * Build http request options
 *
 **
 *  ** deprecated param {RequestQueryParams} params
 * returns {RequestOptions}
 */
export function buildHttpRequestOptions(params: RequestQueryParams): HttpOptionInterface {
    let reqOptions: HttpOptionInterface = {
        headers: new HttpHeaders({
            "Content-Type": 'application/json',
            "Accept": 'application/json',
            "Cache-Control": 'no-cache',
            "Pragma": 'no-cache'
        }),
        responseType: 'json',
    };
    if (params) {
        reqOptions.params = params;
    }

    return reqOptions;
}
export function buildHttpRequestOptionsWithObserveResponse(params: RequestQueryParams): HttpOptionInterface {
    let reqOptions: HttpOptionInterface = {
        headers: new HttpHeaders({
            "Content-Type": 'application/json',
            "Accept": 'application/json',
            "Cache-Control": 'no-cache',
            "Pragma": 'no-cache'
        }),
        responseType: 'json',
        observe: 'response' as 'body'
    };
    if (params) {
        reqOptions.params = params;
    }
    return reqOptions;
}



/** Button events to pass to `DebugElement.triggerEventHandler` for RouterLink event handler */
export const ButtonClickEvents = {
    left: { button: 0 },
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

    compare(a: { [key: string]: any | any[] }, b: { [key: string]: any | any[] }) {
        let comp = 0;
        if (a && b) {
            let fieldA, fieldB;
            for (let key of Object.keys(a)) {
                if (key === this.fieldName) {
                    fieldA = a[key];
                    fieldB = b[key];
                    break;
                } else if (typeof a[key] === 'object') {
                    let insideObject = a[key];
                    for (let insideKey in insideObject) {
                        if (insideKey === this.fieldName) {
                            fieldA = insideObject[insideKey];
                            fieldB = b[key][insideKey];
                            break;
                        }
                    }
                }
            }
            switch (this.type) {
                case "number":
                    comp = fieldB - fieldA;
                    break;
                case "date":
                    comp = new Date(fieldB).getTime() - new Date(fieldA).getTime();
                    break;
                case "string":
                    comp = fieldB.localeCompare(fieldA);
                    break;
            }
        }
        return comp;
    }
}

/**
 * The default page size
 */
export const DEFAULT_PAGE_SIZE: number = 15;

/**
 * The state of vulnerability scanning
 */
export const VULNERABILITY_SCAN_STATUS = {
    unknown: "n/a",
    pending: "pending",
    running: "running",
    error: "error",
    stopped: "stopped",
    finished: "finished"
};

/**
 * Calculate page number by state
 */
export function calculatePage(state: State): number {
    if (!state || !state.page) {
        return 1;
    }

    return Math.ceil((state.page.to + 1) / state.page.size);
}

/**
 * Filter columns via RegExp
 *
 **
 *  ** deprecated param {State} state
 * returns {void}
 */
export function doFiltering<T extends { [key: string]: any | any[] }>(items: T[], state: State): T[] {
    if (!items || items.length === 0) {
        return items;
    }

    if (!state || !state.filters || state.filters.length === 0) {
        return items;
    }

    state.filters.forEach((filter: {
        property: string;
        value: string;
    }) => {
        items = items.filter(item => {
            if (filter['property'].indexOf('.') !== -1) {
                let arr = filter['property'].split('.');
                if (Array.isArray(item[arr[0]]) && item[arr[0]].length) {
                    return item[arr[0]].some((data: any) => {
                        return filter['value'] === data[arr[1]];
                    });
                }
            } else {
                return regexpFilter(filter['value'], item[filter['property']]);
            }
        });
    });

    return items;
}

/**
 * Match items via RegExp
 *
 **
 *  ** deprecated param {string} terms
 *  ** deprecated param {*} testedValue
 * returns {boolean}
 */
export function regexpFilter(terms: string, testedValue: any): boolean {
    let reg = new RegExp('.*' + terms + '.*', 'i');
    return reg.test(testedValue);
}

/**
 * Sorting the data by column
 *
 **
 * template T
 *  ** deprecated param {T[]} items
 *  ** deprecated param {State} state
 * returns {T[]}
 */
export function doSorting<T extends { [key: string]: any | any[] }>(items: T[], state: State): T[] {
    if (!items || items.length === 0) {
        return items;
    }
    if (!state || !state.sort) {
        return items;
    }

    return items.sort((a: T, b: T) => {
        let comp: number = 0;
        if (typeof state.sort.by !== "string") {
            comp = state.sort.by.compare(a, b);
        } else {
            let propA = a[state.sort.by.toString()], propB = b[state.sort.by.toString()];
            if (typeof propA === "string") {
                comp = propA.localeCompare(propB);
            } else {
                if (propA > propB) {
                    comp = 1;
                } else if (propA < propB) {
                    comp = -1;
                }
            }
        }

        if (state.sort.reverse) {
            comp = -comp;
        }

        return comp;
    });
}

/**
 * Compare the two objects to adjust if they're equal
 *
 **
 *  ** deprecated param {*} a
 *  ** deprecated param {*} b
 * returns {boolean}
 */
export function compareValue(a: any, b: any): boolean {
    if ((a && !b) || (!a && b)) { return false; }
    if (!a && !b) { return true; }

    return JSON.stringify(a) === JSON.stringify(b);
}

/**
 * Check if the object is null or empty '{}'
 *
 **
 *  ** deprecated param {*} obj
 * returns {boolean}
 */
export function isEmptyObject(obj: any): boolean {
    return !obj || JSON.stringify(obj) === "{}";
}

/**
 * Deeper clone all
 *
 **
 *  ** deprecated param {*} srcObj
 * returns {*}
 */
export function clone(srcObj: any): any {
    if (!srcObj) { return null; }
    return JSON.parse(JSON.stringify(srcObj));
}

export function isEmpty(obj: any): boolean {
    return !obj || JSON.stringify(obj) === '{}';
}

export function downloadFile(fileData) {
    let url = window.URL.createObjectURL(fileData.data);
    let a = document.createElement("a");
    document.body.appendChild(a);
    a.setAttribute("style", "display: none");
    a.href = url;
    a.download = fileData.filename;
    a.click();
    window.URL.revokeObjectURL(url);
    a.remove();
}

export function getChanges(original: any, afterChange: any): { [key: string]: any | any[] } {
    let changes: { [key: string]: any | any[] } = {};
    if (!afterChange || !original) {
        return changes;
    }
    for (let prop of Object.keys(afterChange)) {
        let field = original[prop];
        if (field && field.editable) {
            if (!compareValue(field.value, afterChange[prop].value)) {
                changes[prop] = afterChange[prop].value;
                // Number
                if (typeof field.value === 'number') {
                    changes[prop] = +changes[prop];
                }

                // Trim string value
                if (typeof field.value === "string") {
                    changes[prop] = ('' + changes[prop]).trim();
                }
            }
        }
    }
    return changes;
}

export function cronRegex(testValue: any): boolean {
    const regSecond = "^((([0-9])*|(\\*))(\\-|\\,|\\/)?([0-9])*)*\\s+";
    const regMinute = "((([0-9])*|(\\*))(\\-|\\,|\\/)?([0-9])*)*\\s+";
    const regHour = "((([0-9])*|(\\*))(\\-|\\,|\\/)?([0-9])*)*\\s+";
    const regDay = "((([0-9])*|(\\*|\\?))(\\-|\\,|\\/)?([0-9])*)*\\s+";
    const regMonth = "((([0-9a-zA-Z])*|(\\*))(\\-|\\,|\\/)?([0-9a-zA-Z])*)*\\s+";
    const regWeek = "(((([0-9a-zA-Z])*|(\\*|\\?))(\\-|\\,|\\/)?([0-9a-zA-Z])*))*(|\\s)+";
    const regYear = "((([0-9])*|(\\*|\\?))(\\-|\\,|\\/)?([0-9])*)$";
    const regEx = regSecond + regMinute + regHour + regDay + regMonth + regWeek + regYear;
    let reg = new RegExp(regEx, "i");
    return reg.test(testValue.trim());
}

/**
 * Keep decimal digits
 * @param count number
 * @param decimals number 1、2、3 ···
 */
export const roundDecimals = (count, decimals = 0) => {
    return Number(`${Math.round(+`${count}e${decimals}`)}e-${decimals}`)
}
