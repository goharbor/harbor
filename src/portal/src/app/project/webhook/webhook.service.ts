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
import { throwError as observableThrowError, Observable } from "rxjs";
import { map, catchError } from "rxjs/operators";
import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { Webhook, LastTrigger } from "./webhook";

@Injectable()
export class WebhookService {
  constructor(private http: HttpClient) { }

  public listWebhook(projectId: number): Observable<Webhook[]> {
    return this.http
      .get(`/api/projects/${projectId}/webhook/policies`)
      .pipe(map(response => response as Webhook[]))
      .pipe(catchError(error => observableThrowError(error)));
  }

  public listLastTrigger(projectId: number): Observable<LastTrigger[]> {
    return this.http
      .get(`/api/projects/${projectId}/webhook/lasttrigger`)
      .pipe(map(response => response as LastTrigger[]))
      .pipe(catchError(error => observableThrowError(error)));
  }

  public editWebhook(projectId: number, policyId: number, data: any): Observable<any> {
    return this.http
      .put(`/api/projects/${projectId}/webhook/policies/${policyId}`, data)
      .pipe(catchError(error => observableThrowError(error)));
  }

  public createWebhook(projectId: number, data: any): Observable<any> {
    return this.http
      .post(`/api/projects/${projectId}/webhook/policies`, data)
      .pipe(catchError(error => observableThrowError(error)));
  }


  public testEndpoint(projectId: number, param): Observable<any> {
    return this.http
      .post(`/api/projects/${projectId}/webhook/policies/test`, param)
      .pipe(catchError(error => observableThrowError(error)));
  }
}
