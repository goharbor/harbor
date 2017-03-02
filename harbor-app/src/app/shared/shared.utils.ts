import { NgForm } from '@angular/forms';

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