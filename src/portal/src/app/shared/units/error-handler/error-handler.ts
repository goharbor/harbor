// Copyright Project Harbor Authors
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
