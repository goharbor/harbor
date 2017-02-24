import { Injectable } from '@angular/core';
import { Headers, Http, RequestOptions } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { PasswordSetting } from './password-setting';

const passwordChangeEndpoint = "/api/users/:user_id/password";

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
        if(!setting || setting.new_password.trim()==="" || setting.old_password.trim()===""){
            return Promise.reject("Invalid data");
        }

        let putUrl = passwordChangeEndpoint.replace(":user_id", userId+"");
        return this.http.put(putUrl, JSON.stringify(setting), this.options)
        .toPromise()
        .then(() => null)
        .catch(error=>{
            return Promise.reject(error);
        });
    }

}
