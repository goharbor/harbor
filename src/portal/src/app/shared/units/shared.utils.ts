// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { NgForm } from '@angular/forms';
import { MessageService } from '../components/global-message/message.service';
import {
    AlertType,
    DatetimeRendering,
    DATETIME_RENDERINGS,
    DEFAULT_DATETIME_RENDERING_LOCALSTORAGE_KEY,
    DefaultDatetimeRendering,
    httpStatusCode,
    SupportedLanguage,
    LANGUAGES,
} from '../entities/shared.const';

/**
 * To check if form is empty
 */
export const isEmptyForm = function (ngForm: NgForm): boolean {
    if (ngForm && ngForm.form) {
        let values = ngForm.form.value;
        if (values) {
            for (let key in values) {
                if (values[key]) {
                    return false;
                }
            }
        }
    }

    return true;
};

/**
 * A type guard that checks if the given value is a supported language.
 */
export function isSupportedLanguage(x: unknown): x is SupportedLanguage {
    return Object.keys(LANGUAGES).some(lang => x === lang);
}

/**
 * Hanlde the 401 and 403 code
 *
 * If handled the 401 or 403, then return true otherwise false
 */
export const accessErrorHandler = function (
    error: any,
    msgService: MessageService
): boolean {
    if (error && error.status && msgService) {
        if (error.status === httpStatusCode.Unauthorized) {
            msgService.announceAppLevelMessage(
                error.status,
                'UNAUTHORIZED_ERROR',
                AlertType.DANGER
            );
            return true;
        }
    }

    return false;
};

// Provide capability of reconstructing the query paramter
export const maintainUrlQueryParmas = function (
    uri: string,
    key: string,
    value: string
): string {
    let re: RegExp = new RegExp('([?&])' + key + '=.*?(&|#|$)', 'i');
    if (value === undefined) {
        if (uri.match(re)) {
            return uri.replace(re, '$1$2');
        } else {
            return uri;
        }
    } else {
        if (uri.match(re)) {
            return uri.replace(re, '$1' + key + '=' + value + '$2');
        } else {
            let hash = '';
            if (uri.indexOf('#') !== -1) {
                hash = uri.replace(/.*#/, '#');
                uri = uri.replace(/#.*/, '');
            }
            let separator = uri.indexOf('?') !== -1 ? '&' : '?';
            return uri + separator + key + '=' + value + hash;
        }
    }
};
/**
 * the password or secret must longer than 8 chars with at least 1 uppercase letter, 1 lowercase letter and 1 number
 * @param randomFlag
 * @param min
 * @param max
 * @returns {string}
 */

export function randomWord(max) {
    let str = '';

    let contentArray = [
        ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9'],
        [
            'a',
            'b',
            'c',
            'd',
            'e',
            'f',
            'g',
            'h',
            'i',
            'j',
            'k',
            'l',
            'm',
            'n',
            'o',
            'p',
            'q',
            'r',
            's',
            't',
            'u',
            'v',
            'w',
            'x',
            'y',
            'z',
        ],
        [
            'A',
            'B',
            'C',
            'D',
            'E',
            'F',
            'G',
            'H',
            'I',
            'J',
            'K',
            'L',
            'M',
            'N',
            'O',
            'P',
            'Q',
            'R',
            'S',
            'T',
            'U',
            'V',
            'W',
            'X',
            'Y',
            'Z',
        ],
    ];
    for (let i = 0; i < max; i++) {
        let randomNumber = getRandomInt(contentArray.length);
        str +=
            contentArray[randomNumber][
                getRandomInt(contentArray[randomNumber].length)
            ];
    }
    if (!str.match(/\d+/g)) {
        str += contentArray[0][getRandomInt(contentArray[0].length)];
    }
    if (!str.match(/[a-z]+/g)) {
        str += contentArray[1][getRandomInt(contentArray[1].length)];
    }
    if (!str.match(/[A-Z]+/g)) {
        str += contentArray[2][getRandomInt(contentArray[2].length)];
    }
    return str;
}
function getRandomInt(max) {
    return Math.floor(Math.random() * Math.floor(max));
}

/**
 *  handle docker client response error
 * {"code":412,"message":"{\"errors\":[{\"code\":\"DENIED\",\"message\":\"Failed to process request,
 * due to 'golang1:test1' is a immutable tag.\",\"detail\":\"Failed to process request,
 * due to 'golang1:test1' is a immutable tag.\"}]}\n"}
 * @param errorString string
 */
const errorHandlerForDockerClient = function (errorString: string): string {
    try {
        const errorMsgBody = JSON.parse(errorString);
        if (
            errorMsgBody.errors &&
            errorMsgBody.errors[0] &&
            errorMsgBody.errors[0].message
        ) {
            return errorMsgBody.errors[0].message;
        }
    } catch (err) {}
    return errorString;
};

/**
 * To handle the error message body
 * Standard error return format {code : number, message: string} / {error: {code: number, message: string},...}
 **
 * returns {string}
 */

export const errorHandler = function (error: any): string {
    if (!error) {
        return 'UNKNOWN_ERROR';
    }
    // oci standard
    if (error.errors && error.errors instanceof Array && error.errors.length) {
        return error.errors.reduce((preError, currentError, index) => {
            return preError
                ? `${preError},${currentError.message}`
                : currentError.message;
        }, '');
    }
    // Not a standard error return Basically not used cover unknown error
    try {
        return JSON.parse(error.error).message;
    } catch (err) {}
    // Not a standard error return Basically not used cover unknown error
    if (typeof error.error === 'string') {
        return error.error;
    }
    // oci standard
    if (
        error.error &&
        error.error.errors &&
        error.error.errors instanceof Array &&
        error.error.errors.length
    ) {
        return error.error.errors.reduce((preError, currentError, index) => {
            return preError
                ? `${preError},${currentError.message}`
                : currentError.message;
        }, '');
    }
    if (error.error && error.error.message) {
        if (typeof error.error.message === 'string') {
            // handle docker client response error
            return errorHandlerForDockerClient(error.error.message);
        }
    }
    if (error.message) {
        // handle docker client response error
        if (typeof error.message === 'string') {
            return errorHandlerForDockerClient(error.message);
        }
    }
    // Not a standard error return Basically not used cover unknown error
    if (!(error.statusCode || error.status)) {
        // treat as string message
        return '' + error;
    } else {
        switch (error.statusCode || error.status) {
            case 400:
                return 'BAD_REQUEST_ERROR';
            case 401:
                return 'UNAUTHORIZED_ERROR';
            case 403:
                return 'FORBIDDEN_ERROR';
            case 404:
                return 'NOT_FOUND_ERROR';
            case 412:
                return 'PRECONDITION_FAILED';
            case 409:
                return 'CONFLICT_ERROR';
            case 500:
                return 'SERVER_ERROR';
            default:
                return 'UNKNOWN_ERROR';
        }
    }
};

/**
 * Gets the datetime rendering setting saved by the user, or the default setting if no valid saved value is found.
 */
export function getDatetimeRendering(): DatetimeRendering {
    const savedDatetimeRendering = localStorage.getItem(
        DEFAULT_DATETIME_RENDERING_LOCALSTORAGE_KEY
    );
    if (savedDatetimeRendering && isDatetimeRendering(savedDatetimeRendering)) {
        return savedDatetimeRendering;
    }
    return DefaultDatetimeRendering;
}

function isDatetimeRendering(x: unknown): x is DatetimeRendering {
    return Object.keys(DATETIME_RENDERINGS).some(k => k === x);
}
