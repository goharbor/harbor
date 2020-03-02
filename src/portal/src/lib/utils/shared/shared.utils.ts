import { errorSrcWithoutHttpClient } from "ngx-markdown";

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
        if (errorMsgBody.errors && errorMsgBody.errors[0] && errorMsgBody.errors[0].message) {
            return errorMsgBody.errors[0].message;
        }
    } catch (err) { }
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
        return "UNKNOWN_ERROR";
    }
    // Not a standard error return Basically not used cover unknown error
    try {
        return JSON.parse(error.error).message;
    } catch (err) { }
    // Not a standard error return Basically not used cover unknown error
    if (typeof error.error === "string") {
        return error.error;
    }
    if (error.error && error.error.message) {
        if (typeof error.error.message === "string") {
            // handle docker client response error
            return errorHandlerForDockerClient(error.error.message);
        }
    }
    if (error.message) {
        // handle docker client response error
        if (typeof error.message === "string") {
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
                return "BAD_REQUEST_ERROR";
            case 401:
                return "UNAUTHORIZED_ERROR";
            case 403:
                return "FORBIDDEN_ERROR";
            case 404:
                return "NOT_FOUND_ERROR";
            case 412:
                return "PRECONDITION_FAILED";
            case 409:
                return "CONFLICT_ERROR";
            case 500:
                return "SERVER_ERROR";
            default:
                return "UNKNOWN_ERROR";
        }
    }
};
