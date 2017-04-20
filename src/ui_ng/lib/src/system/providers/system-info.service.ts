import { Injectable, Inject } from '@angular/core';
import { Http, RequestOptions, Headers } from '@angular/http';
import 'rxjs/add/operator/toPromise';

import { SERVICE_CONFIG, IServiceConfig } from '../../service.config';

@Injectable()
export class SystemInfoService {
  httpOptions = new RequestOptions({
    headers: new Headers({
      "Content-Type": 'application/json'
    })
  });

  constructor(
    private http: Http,
    @Inject(SERVICE_CONFIG) private config: IServiceConfig) { }

  getSystemInfo(): Promise<any> {
    if(this.config.systemInfoEndpoint.trim() === "") {
      return Promise.reject("500: Internal error");
    }

    return this.http.get(this.config.systemInfoEndpoint, this.httpOptions).toPromise()
    .then(response => response.json())
    .catch(error => console.error("Get systeminfo error: ", error));
  }

}