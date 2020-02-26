import { Injectable, Inject } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";

import { SERVICE_CONFIG, IServiceConfig } from "../entities/service.config";
import { HTTP_JSON_OPTIONS, HTTP_GET_OPTIONS, CURRENT_BASE_HREF } from "../utils/utils";
import { Configuration } from "../components/config/config";

/**
 * Service used to get and save registry-related configurations.
 *
 **
 * @abstract
 * class ConfigurationService
 */
export abstract class ConfigurationService {
  /**
   * Get configurations.
   *
   * @abstract
   * returns {(Observable<Configuration>)}
   *
   * @memberOf ConfigurationService
   */
  abstract getConfigurations():
    | Observable<Configuration>;

  /**
   * Save configurations.
   *
   * @abstract
   * returns {(Observable<Configuration>)}
   *
   * @memberOf ConfigurationService
   */
  abstract saveConfigurations(
    changedConfigs: any | { [key: string]: any | any[] }
  ): Observable<any>;
}

@Injectable()
export class ConfigurationDefaultService extends ConfigurationService {
  _baseUrl: string;

  constructor(
    private http: HttpClient,
    @Inject(SERVICE_CONFIG) private config: IServiceConfig
  ) {
    super();

    this._baseUrl =
      this.config && this.config.configurationEndpoint
        ? this.config.configurationEndpoint
        : CURRENT_BASE_HREF + "/configurations";
  }

  getConfigurations():
    | Observable<Configuration> {
    return this.http
      .get(this._baseUrl, HTTP_GET_OPTIONS)
      .pipe(map(response => response as Configuration)
      , catchError(error => observableThrowError(error)));
  }

  saveConfigurations(
    changedConfigs: any | { [key: string]: any | any[] }
  ): Observable<any> {
    if (!changedConfigs) {
      return observableThrowError("Bad argument!");
    }

    return this.http
      .put(this._baseUrl, JSON.stringify(changedConfigs), HTTP_JSON_OPTIONS)
      .pipe(map(() => { })
      , catchError(error => observableThrowError(error)));
  }
}
