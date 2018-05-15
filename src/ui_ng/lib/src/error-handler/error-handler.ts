import { Injectable } from "@angular/core";

/**
 * Declare interface for error handling
 *
 * @export
 * @abstract
 * @class ErrorHandler
 */
export abstract class ErrorHandler {
  /**
   * Send message with error level
   *
   * @abstract
   * @param {*} error
   *
   * @memberOf ErrorHandler
   */
  abstract error(error: any): void;

  /**
   * Send message with warning level
   *
   * @abstract
   * @param {*} warning
   *
   * @memberOf ErrorHandler
   */
  abstract warning(warning: any): void;

  /**
   * Send message with info level
   *
   * @abstract
   * @param {*} info
   *
   * @memberOf ErrorHandler
   */
  abstract info(info: any): void;

  /**
   * Handle log message
   *
   * @abstract
   * @param {*} log
   *
   * @memberOf ErrorHandler
   */
  abstract log(log: any): void;
}

@Injectable()
export class DefaultErrorHandler extends ErrorHandler {
  public error(error: any): void {
    console.error("[Default error handler]: ", error);
  }

  public warning(warning: any): void {
    console.warn("[Default warning handler]: ", warning);
  }

  public info(info: any): void {
    // tslint:disable-next-line:no-console
    console.info("[Default info handler]: ", info);
  }

  public log(log: any): void {
    console.log("[Default log handler]: ", log);
  }
}
