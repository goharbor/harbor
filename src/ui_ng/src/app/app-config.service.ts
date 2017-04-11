import { Injectable } from '@angular/core';
import { Headers, Http, RequestOptions } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { AppConfig } from './app-config';
import { NgXCookies } from 'ngx-cookies';
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
    headers = new Headers({
        "Content-Type": 'application/json'
    });
    options = new RequestOptions({
        headers: this.headers
    });

    //Store the application configuration
    configurations: AppConfig = new AppConfig();

    constructor(
        private http: Http) { }

    public load(): Promise<AppConfig> {
        return this.http.get(systemInfoEndpoint, this.options).toPromise()
            .then(response => {
                this.configurations = response.json() as AppConfig;

                //Read admiral endpoint from cookie if existing
                let admiralUrlFromCookie: string = NgXCookies.getCookie(CookieKeyOfAdmiral);
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
        NgXCookies.setCookie(CookieKeyOfAdmiral, endpoint);
        this.configurations.admiral_endpoint = endpoint;
    }
}