/**
 * Declare interface for error handling
 *
 **
 * @abstract
 * class ErrorHandler
 */
export abstract class ErrorHandler {
    /**
     * Send message with error level
     *
     * @abstract
     *  ** deprecated param {*} error
     *
     * @memberOf ErrorHandler
     */
    abstract error(error: any): void;

    /**
     * Send message with warning level
     *
     * @abstract
     *  ** deprecated param {*} warning
     *
     * @memberOf ErrorHandler
     */
    abstract warning(warning: any): void;

    /**
     * Send message with info level
     *
     * @abstract
     *  ** deprecated param {*} info
     *
     * @memberOf ErrorHandler
     */
    abstract info(info: any): void;

    /**
     * Handle log message
     *
     * @abstract
     *  ** deprecated param {*} log
     *
     * @memberOf ErrorHandler
     */
    abstract log(log: any): void;

    abstract handleErrorPopupUnauthorized(error: any): void;
}
