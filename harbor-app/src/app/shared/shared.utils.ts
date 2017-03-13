import { NgForm } from '@angular/forms';
import { httpStatusCode, AlertType } from './shared.const';
import { MessageService } from '../global-message/message.service';
/**
 * To handle the error message body
 * 
 * @export
 * @returns {string}
 */
export const errorHandler = function (error: any): string {
    if (error) {
        if (error.message) {
            return error.message;
        } else if (error._body) {
            return error._body;
        } else if (error.statusText) {
            return error.statusText;
        } else {
            return error;
        }
    }

    return "UNKNOWN_ERROR";
}

/**
 * To check if form is empty
 */
export const isEmptyForm = function (ngForm: NgForm): boolean {
    if (ngForm && ngForm.form) {
        let values = ngForm.form.value;
        if (values) {
            for (var key in values) {
                if (values[key]) {
                    return false;
                }
            }
        }

    }

    return true;
}

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
        } else if (error.status === httpStatusCode.Forbidden) {
            msgService.announceAppLevelMessage(error.status, "FORBIDDEN_ERROR", AlertType.DANGER);
            return true;
        }
    }

    return false;
}