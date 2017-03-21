import { Injectable } from '@angular/core';
import { Headers, Http, RequestOptions } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { AppConfig } from './app-config';
import { CookieService } from 'angular2-cookie/core';
import { CookieKeyOfAdmiral, HarborQueryParamKey } from './shared/shared.const';
import { maintainUrlQueryParmas } from './shared/shared.utils';


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

    constructor(
        private http: Http,
        private cookie: CookieService) { }

    public load(): Promise<AppConfig> {
        return this.http.get(systemInfoEndpoint, this.options).toPromise()
            .then(response => {
                this.configurations = response.json() as AppConfig;

                //Read admiral endpoint from cookie if existing
                let admiralUrlFromCookie: string = this.cookie.get(CookieKeyOfAdmiral);
                if(admiralUrlFromCookie){
                    //Override the endpoint from configuration file
                    this.configurations.admiral_endpoint = decodeURIComponent(admiralUrlFromCookie);
                }

                return this.configurations;
            })
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

    //Return the reconstructed admiral url
    public getAdmiralEndpoint(currentHref: string): string {
        let admiralUrl:string = this.configurations.admiral_endpoint;
        if(admiralUrl.trim() === "" || currentHref.trim() === ""){
            return "#";
        }

        return maintainUrlQueryParmas(admiralUrl, HarborQueryParamKey, encodeURIComponent(currentHref));
    }

    public saveAdmiralEndpoint(endpoint: string): void {
        if(!(endpoint.trim())){
            return;
        }

        //Save back to cookie
        this.cookie.put(CookieKeyOfAdmiral, endpoint);
        this.configurations.admiral_endpoint = endpoint;
    }
}