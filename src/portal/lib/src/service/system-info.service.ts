import {ElementRef, Inject, Injectable} from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
import {SystemCVEWhitelist, SystemInfo} from './interface';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import {HTTP_GET_OPTIONS, HTTP_JSON_OPTIONS} from "../utils";

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
  /**
   *  get system CEVWhitelist
   */
  abstract getSystemWhitelist(): Observable<SystemCVEWhitelist>;
  /**
   *  update systemCVEWhitelist
   * @param systemCVEWhitelist
   */
  abstract updateSystemWhitelist(systemCVEWhitelist: SystemCVEWhitelist): Observable<any>;
  /**
   *  set null to the date type input
   * @param ref
   */
  abstract resetDateInput(ref: ElementRef);
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
  public getSystemWhitelist(): Observable<SystemCVEWhitelist> {
    return this.http.get("/api/system/CVEWhitelist", HTTP_GET_OPTIONS)
        .pipe(map(systemCVEWhitelist => systemCVEWhitelist as SystemCVEWhitelist)
            , catchError(error => observableThrowError(error)));
  }
  public updateSystemWhitelist(systemCVEWhitelist: SystemCVEWhitelist): Observable<any> {
    return this.http.put("/api/system/CVEWhitelist", JSON.stringify(systemCVEWhitelist), HTTP_JSON_OPTIONS)
        .pipe(map(response => response)
            , catchError(error => observableThrowError(error)));
  }
  public resetDateInput(ref: ElementRef) {
    if (ref) {
      ref.nativeElement.value = null ;
    }
  }
}

