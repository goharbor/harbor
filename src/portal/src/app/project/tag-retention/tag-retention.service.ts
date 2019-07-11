// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import {Retention} from "./retention";
import {Observable, throwError as observableThrowError} from "rxjs";
import { map, catchError } from "rxjs/operators";
import {Project} from "../project";
@Injectable()
export class TagRetentionService {
  constructor(
    private http: HttpClient,
  ) { }
  getRetentionMetadata(): Observable<object>  {
    return this.http.get(`/api/retentions/metadatas`)
        .pipe(map(response => response as object))
        .pipe(catchError(error => observableThrowError(error)));
  }
  getRetention(retentionId): Observable<Retention> {
    return this.http.get(`/api/retentions/${retentionId}`)
        .pipe(map(response => response as Retention))
        .pipe(catchError(error => observableThrowError(error)));
  }
  createRetention(retention: Retention) {
    return this.http.post(`/api/retentions`, retention)
        .pipe(catchError(error => observableThrowError(error)));
  }
  updateRetention(retentionId, retention: Retention) {
    return this.http.put(`/api/retentions/${retentionId}`, retention)
        .pipe(catchError(error => observableThrowError(error)));
  }
  getProjectInfo(projectId) {
    return this.http.get(`/api/projects/${projectId}`)
        .pipe(map(response => response as Project))
        .pipe(catchError(error => observableThrowError(error)));
  }
  runNowTrigger(retentionId) {
    return this.http.post(`/api/retentions/${retentionId}/executions`, {dry_run: false})
        .pipe(catchError(error => observableThrowError(error)));
  }
  whatIfRunTrigger(retentionId) {
    return this.http.post(`/api/retentions/${retentionId}/executions`, {dry_run: true})
        .pipe(catchError(error => observableThrowError(error)));
  }
  AbortRun(retentionId, executionId) {
    return this.http.patch(`/api/retentions/${retentionId}/executions/${executionId}`, {action: 'stop'})
        .pipe(catchError(error => observableThrowError(error)));
  }
  getRunNowList(retentionId) {
    return this.http.get(`/api/retentions/${retentionId}/executions`)
        .pipe(map(response => response as Array<any>))
        .pipe(catchError(error => observableThrowError(error)));
  }
  getExecutionHistory(retentionId, executionId) {
    return this.http.get(`/api/retentions/${retentionId}/executions/${executionId}/tasks`)
        .pipe(map(response => response as Array<any>))
        .pipe(catchError(error => observableThrowError(error)));
  }
  seeLog(retentionId, executionId, taskId) {
    window.open(`api/retention/${retentionId}/executions/${executionId}/tasks/${taskId}`, '_blank');
  }
}
