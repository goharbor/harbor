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
 * To handle the error message body
 *
 * @export
 * @returns {string}
 */
export const errorHandler = function (error: any): string {
    if (!error) {
        return "UNKNOWN_ERROR";
    }
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

export class CancelablePromise<T> {

  constructor(promise: Promise<T>) {
    this.wrappedPromise = new Promise((resolve, reject) => {
      promise.then((val) =>
        this.isCanceled ? reject({isCanceled: true}) : resolve(val)
      );
      promise.catch((error) =>
        this.isCanceled ? reject({isCanceled: true}) : reject(error)
      );
    });
  }

  private wrappedPromise: Promise<T>;
  private isCanceled: boolean;
  getPromise(): Promise<T> {
    return this.wrappedPromise;
  }

  cancel() {
    this.isCanceled = true;
  }
}
