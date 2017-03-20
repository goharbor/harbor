import { Injectable } from '@angular/core';
import { Headers, Http, RequestOptions } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { AppConfig } from './app-config';

export const systemInfoEndpoint = "/api/systeminfo";
/**
 * Declare service to handle the bootstrap options
 * 
 * 
 * @export
 * @class GlobalSearchService
 */
@Injectable()
export class AppConfigService {
    private headers = new Headers({
        "Content-Type": 'application/json'
    });
    private options = new RequestOptions({
        headers: this.headers
    });

    //Store the application configuration
    private configurations: AppConfig = new AppConfig();

    constructor(private http: Http) { }

    public load(): Promise<AppConfig> {
        return this.http.get(systemInfoEndpoint, this.options).toPromise()
        .then(response => this.configurations = response.json() as AppConfig)
        .catch(error => {
            //Catch the error
            console.error("Failed to load bootstrap options with error: ", error);
        });
    }

    public getConfig(): AppConfig {
        return this.configurations;
    }

    public isIntegrationMode(): boolean {
        return this.configurations && 
        this.configurations.with_admiral && 
        this.configurations.admiral_endpoint.trim() != "";
    }
}