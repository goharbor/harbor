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
import { Headers, Http, RequestOptions, URLSearchParams } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { PasswordSetting } from './password-setting';

const passwordChangeEndpoint = "/api/users/:user_id/password";
const sendEmailEndpoint = "/sendEmail";
const resetPasswordEndpoint = "/reset";

@Injectable()
export class PasswordSettingService {
    private headers: Headers = new Headers({
        "Accept": 'application/json',
        "Content-Type": 'application/json'
    });
    private options: RequestOptions = new RequestOptions({
        'headers': this.headers
    });

    constructor(private http: Http) { }

    changePassword(userId: number, setting: PasswordSetting): Promise<any> {
        if (!setting || setting.new_password.trim() === "" || setting.old_password.trim() === "") {
            return Promise.reject("Invalid data");
        }

        let putUrl = passwordChangeEndpoint.replace(":user_id", userId + "");
        return this.http.put(putUrl, JSON.stringify(setting), this.options)
            .toPromise()
            .then(() => null)
            .catch(error => {
                return Promise.reject(error);
            });
    }

    sendResetPasswordMail(email: string): Promise<any> {
        if (!email) {
            return Promise.reject("Invalid email");
        }

        let getUrl = sendEmailEndpoint + "?email=" + email;
        return this.http.get(getUrl, this.options).toPromise()
            .then(response => response)
            .catch(error => {
                return Promise.reject(error);
            })
    }

    resetPassword(uuid: string, newPassword: string): Promise<any> {
        if (!uuid || !newPassword) {
            return Promise.reject("Invalid reset uuid or password");
        }

        let formHeaders = new Headers({
            "Content-Type": 'application/x-www-form-urlencoded'
        });
        let formOptions: RequestOptions = new RequestOptions({
            headers: formHeaders
        });
        
        let body: URLSearchParams = new URLSearchParams();
        body.set("reset_uuid", uuid);
        body.set("password", newPassword);

        return this.http.post(resetPasswordEndpoint, body.toString(), formOptions)
            .toPromise()
            .then(response => response)
            .catch(error => {
                return Promise.reject(error);
            });
    }

}
