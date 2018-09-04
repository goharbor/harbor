import { Inject, Injectable } from '@angular/core';
import { Http } from '@angular/http';
import { Observable } from 'rxjs/Observable';
import { SystemInfo } from './interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import {HTTP_GET_OPTIONS} from "../utils";
/**
 * Get System information about current backend server.
 * @abstract
 * @class
 */
export abstract class SystemInfoService {
  /**
   *  Get global system information.
   *  @abstract
   *  @returns
   */
  abstract getSystemInfo(): Observable<SystemInfo> | Promise<SystemInfo> | SystemInfo;
}

@Injectable()
export class SystemInfoDefaultService extends SystemInfoService {
  constructor(
    @Inject(SERVICE_CONFIG) private config: IServiceConfig,
    private http: Http) {
    super();
  }
  getSystemInfo(): Observable<SystemInfo> | Promise<SystemInfo> | SystemInfo {
    let url = this.config.systemInfoEndpoint ? this.config.systemInfoEndpoint : '/api/systeminfo';
    return this.http.get(url, HTTP_GET_OPTIONS)
      .toPromise()
      .then(systemInfo => systemInfo.json() as SystemInfo)
      .catch(error => Promise.reject(error));
  }
}

