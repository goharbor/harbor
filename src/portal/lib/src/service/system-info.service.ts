import { Inject, Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
import { SystemInfo } from './interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { HTTP_GET_OPTIONS } from "../utils";

/**
 * Get System information about current backend server.
 * @abstract
 * class
 */
export abstract class SystemInfoService {
  /**
   *  Get global system information.
   *  @abstract
   *  returns
   */
  abstract getSystemInfo(): Observable<SystemInfo>;
}

@Injectable()
export class SystemInfoDefaultService extends SystemInfoService {
  constructor(
    @Inject(SERVICE_CONFIG) private config: IServiceConfig,
    private http: HttpClient) {
    super();
  }
  getSystemInfo(): Observable<SystemInfo> {
    let url = this.config.systemInfoEndpoint ? this.config.systemInfoEndpoint : '/api/systeminfo';
    return this.http.get(url, HTTP_GET_OPTIONS)
      .pipe(map(systemInfo => systemInfo as SystemInfo)
      , catchError(error => observableThrowError(error)));
  }
}

