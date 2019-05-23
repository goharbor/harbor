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
import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";


import { PasswordSetting } from './password-setting';

import {HTTP_FORM_OPTIONS, HTTP_JSON_OPTIONS, HTTP_GET_OPTIONS} from "@harbor/ui";

const passwordChangeEndpoint = "/api/users/:user_id/password";
const sendEmailEndpoint = "/c/sendEmail";
const resetPasswordEndpoint = "/c/reset";

@Injectable()
export class PasswordSettingService {

    constructor(private http: HttpClient) { }

    changePassword(userId: number, setting: PasswordSetting): Observable<any> {
        if (!setting || setting.new_password.trim() === "" || setting.old_password.trim() === "") {
            return observableThrowError("Invalid data");
        }

        let putUrl = passwordChangeEndpoint.replace(":user_id", userId + "");
        return this.http.put(putUrl, JSON.stringify(setting), HTTP_JSON_OPTIONS)
            .pipe(map(() => null)
            , catchError(error => observableThrowError(error)));
    }

    sendResetPasswordMail(email: string): Observable<any> {
        if (!email) {
            return observableThrowError("Invalid email");
        }

        let getUrl = sendEmailEndpoint + "?email=" + email;
        return this.http.get(getUrl, HTTP_GET_OPTIONS)
            .pipe(map(response => response)
            , catchError(error => observableThrowError(error)));
    }

    resetPassword(uuid: string, newPassword: string): Observable<any> {
        if (!uuid || !newPassword) {
            return observableThrowError("Invalid reset uuid or password");
        }

        let body: HttpParams = new HttpParams().set("reset_uuid", uuid).set("password", newPassword);
        return this.http.post(resetPasswordEndpoint, body.toString(), HTTP_FORM_OPTIONS)
            .pipe(map(response => response)
            , catchError(error => observableThrowError(error)));
    }

}
