import { Injectable } from '@angular/core';
import { Headers, Http, RequestOptions } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { Configuration } from './config';

const configEndpoint = "/api/configurations";

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
}
