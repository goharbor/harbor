// Copyright Project Harbor Authors
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
import { HttpClient } from '@angular/common/http';
import { CookieService } from 'ngx-cookie';
import { AppConfig } from './app-config';
import { maintainUrlQueryParmas } from '../shared/units/shared.utils';
import { map } from 'rxjs/operators';
import { Observable } from 'rxjs';
import { CURRENT_BASE_HREF, HTTP_GET_OPTIONS } from '../shared/units/utils';
import {
    CONFIG_AUTH_MODE,
    CookieKeyOfAdmiral,
    HarborQueryParamKey,
} from '../shared/entities/shared.const';
export const systemInfoEndpoint = CURRENT_BASE_HREF + '/systeminfo';
/**
 * Declare service to handle the bootstrap options
 *
 *
 **
 * class GlobalSearchService
 */
@Injectable()
export class AppConfigService {
    // Store the application configuration
    configurations: AppConfig = new AppConfig();

    private _bannerMessageClosed: boolean = false;

    constructor(private http: HttpClient, private cookie: CookieService) {}

    setBannerMessageClosed(v: boolean) {
        this._bannerMessageClosed = v;
    }

    getBannerMessageClosed(): boolean {
        return this._bannerMessageClosed;
    }

    public load(): Observable<AppConfig> {
        return this.http.get(systemInfoEndpoint, HTTP_GET_OPTIONS).pipe(
            map(response => {
                this.configurations = response as AppConfig;

                // Read admiral endpoint from cookie if existing
                let admiralUrlFromCookie: string =
                    this.cookie.get(CookieKeyOfAdmiral);
                if (admiralUrlFromCookie) {
                    // Override the endpoint from configuration file
                    this.configurations.admiral_endpoint =
                        decodeURIComponent(admiralUrlFromCookie);
                }

                return this.configurations;
            })
        );
    }

    public getConfig(): AppConfig {
        return this.configurations;
    }

    public isLdapMode(): boolean {
        return (
            this.configurations &&
            this.configurations.auth_mode === CONFIG_AUTH_MODE.LDAP_AUTH
        );
    }
    public isHttpAuthMode(): boolean {
        return (
            this.configurations &&
            this.configurations.auth_mode === CONFIG_AUTH_MODE.HTTP_AUTH
        );
    }
    public isOidcMode(): boolean {
        return (
            this.configurations &&
            this.configurations.auth_mode === CONFIG_AUTH_MODE.OIDC_AUTH
        );
    }

    // Return the reconstructed admiral url
    public getAdmiralEndpoint(currentHref: string): string {
        let admiralUrl: string = this.configurations.admiral_endpoint;
        if (admiralUrl.trim() === '' || currentHref.trim() === '') {
            return '#';
        }

        return maintainUrlQueryParmas(
            admiralUrl,
            HarborQueryParamKey,
            encodeURIComponent(currentHref)
        );
    }

    public saveAdmiralEndpoint(endpoint: string): void {
        if (!endpoint.trim()) {
            return;
        }

        // Save back to cookie
        this.cookie.put(CookieKeyOfAdmiral, endpoint);
        this.configurations.admiral_endpoint = endpoint;
    }
}
