import { Injectable, Inject } from "@angular/core";
import { Http } from "@angular/http";
import { Observable } from "rxjs/Observable";
import "rxjs/add/observable/of";

import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import { HTTP_JSON_OPTIONS, HTTP_GET_OPTIONS } from "../utils";
import { Configuration } from "../config/config";

/**
 * Service used to get and save registry-related configurations.
 *
 * @export
 * @abstract
 * @class ConfigurationService
 */
export abstract class ConfigurationService {
  /**
   * Get configurations.
   *
   * @abstract
   * @returns {(Observable<Configuration> | Promise<Configuration> | Configuration)}
   *
   * @memberOf ConfigurationService
   */
  abstract getConfigurations():
    | Observable<Configuration>
    | Promise<Configuration>
    | Configuration;

  /**
   * Save configurations.
   *
   * @abstract
   * @returns {(Observable<Configuration> | Promise<Configuration> | Configuration)}
   *
   * @memberOf ConfigurationService
   */
  abstract saveConfigurations(
    changedConfigs: any | { [key: string]: any | any[] }
  ): Observable<any> | Promise<any> | any;
}

@Injectable()
export class ConfigurationDefaultService extends ConfigurationService {
  _baseUrl: string;

  constructor(
    private http: Http,
    @Inject(SERVICE_CONFIG) private config: IServiceConfig
  ) {
    super();

    this._baseUrl =
      this.config && this.config.configurationEndpoint
        ? this.config.configurationEndpoint
        : "/api/configurations";
  }

  getConfigurations():
    | Observable<Configuration>
    | Promise<Configuration>
    | Configuration {
    return this.http
      .get(this._baseUrl, HTTP_GET_OPTIONS)
      .toPromise()
      .then(response => response.json() as Configuration)
      .catch(error => Promise.reject(error));
  }

  saveConfigurations(
    changedConfigs: any | { [key: string]: any | any[] }
  ): Observable<any> | Promise<any> | any {
    if (!changedConfigs) {
      return Promise.reject("Bad argument!");
    }

    return this.http
      .put(this._baseUrl, JSON.stringify(changedConfigs), HTTP_JSON_OPTIONS)
      .toPromise()
      .then(() => {})
      .catch(error => Promise.reject(error));
  }
}
