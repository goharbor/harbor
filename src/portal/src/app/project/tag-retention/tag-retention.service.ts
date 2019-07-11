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
import {Retention, RuleMetadate} from "./retention";
import {Observable, throwError as observableThrowError} from "rxjs";
import { map, catchError } from "rxjs/operators";
import {Project} from "../project";
@Injectable()
export class TagRetentionService {
  private I18nMap: object = {
    "retain": "ACTION_RETAIN",
    "lastXDays": "RULE_NAME_1",
    "latestActiveK": "RULE_NAME_2",
    "latestK": "RULE_NAME_3",
    "latestPulledK": "RULE_NAME_4",
    "always": "RULE_NAME_5",
    "the images from the last # days": "RULE_TEMPLATE_1",
    "the most recent active # images": "RULE_TEMPLATE_2",
    "the most recently pushed # images": "RULE_TEMPLATE_3",
    "the most recently pulled # images": "RULE_TEMPLATE_4",
    "repoMatches": "MAT",
    "repoExcludes": "EXC",
    "matches": "MAT",
    "excludes": "EXC",
    "withLabels": "WITH",
    "withoutLabels": "WITHOUT",
    "COUNT": "UNIT_COUNT",
    "DAYS": "UNIT_DAY",
  };
  constructor(
    private http: HttpClient,
  ) { }
  getI18nKey(str: string): string {
    if (this.I18nMap[str.trim()]) {
      return "TAG_RETENTION." + this.I18nMap[str.trim()];
    }
   return str;
  }
  getRetentionMetadata(): Observable<RuleMetadate>  {
    return this.http.get(`/api/retentions/metadatas`)
        .pipe(map(response => response as RuleMetadate))
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
