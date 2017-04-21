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
import { Headers, Http, RequestOptions } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { Configuration } from './config';

const configEndpoint = "/api/configurations";
const emailEndpoint = "/api/email/ping";
const ldapEndpoint = "/api/ldap/ping";

@Injectable()
export class ConfigurationService {
    private headers: Headers = new Headers({
        "Accept": 'application/json',
        "Content-Type": 'application/json'
    });
    private options: RequestOptions = new RequestOptions({
        'headers': this.headers
    });

    constructor(private http: Http) { }

    public getConfiguration(): Promise<Configuration> {
        return this.http.get(configEndpoint, this.options).toPromise()
        .then(response => response.json() as Configuration)
        .catch(error => Promise.reject(error));
    }

    public saveConfiguration(values: any): Promise<any> {
        return this.http.put(configEndpoint, JSON.stringify(values), this.options)
        .toPromise()
        .then(response => response)
        .catch(error => Promise.reject(error));
    }

    public testMailServer(mailSettings: any): Promise<any> {
        return this.http.post(emailEndpoint, JSON.stringify(mailSettings), this.options)
        .toPromise()
        .then(response => response)
        .catch(error => Promise.reject(error));
    }

    public testLDAPServer(ldapSettings: any): Promise<any> {
         return this.http.post(ldapEndpoint, JSON.stringify(ldapSettings), this.options)
        .toPromise()
        .then(response => response)
        .catch(error => Promise.reject(error));
    }
}
