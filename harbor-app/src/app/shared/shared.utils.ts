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