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
import { MessageService } from '../global-message/message.service';
import { AlertType } from '../shared/shared.const';
import { httpStatusCode } from "../../lib/entities/shared.const";

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
 * Hanlde the 401 and 403 code
 *
 * If handled the 401 or 403, then return true otherwise false
 */
export const accessErrorHandler = function (error: any, msgService: MessageService): boolean {
    if (error && error.status && msgService) {
        if (error.status === httpStatusCode.Unauthorized) {
            msgService.announceAppLevelMessage(error.status, "UNAUTHORIZED_ERROR", AlertType.DANGER);
            return true;
        }
    }

    return false;
};

// Provide capability of reconstructing the query paramter
export const maintainUrlQueryParmas = function (uri: string, key: string, value: string): string {
    let re: RegExp = new RegExp("([?&])" + key + "=.*?(&|#|$)", "i");
    if (value === undefined) {
        if (uri.match(re)) {
            return uri.replace(re, '$1$2');
        } else {
            return uri;
        }
    } else {
        if (uri.match(re)) {
            return uri.replace(re, '$1' + key + "=" + value + '$2');
        } else {
            let hash = '';
            if (uri.indexOf('#') !== -1) {
                hash = uri.replace(/.*#/, '#');
                uri = uri.replace(/#.*/, '');
            }
            let separator = uri.indexOf('?') !== -1 ? "&" : "?";
            return uri + separator + key + "=" + value + hash;
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
    let str = "";

     let contentArray = [['0', '1', '2', '3', '4', '5', '6', '7', '8', '9'],
    ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l'
    , 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x'
    , 'y', 'z'],
    ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J'
    , 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z']];
    for (let i = 0; i < max; i++) {
        let randomNumber = getRandomInt(contentArray.length);
        str += contentArray[randomNumber][getRandomInt(contentArray[randomNumber].length)];
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

