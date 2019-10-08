import { Injectable } from '@angular/core';
import { HttpClient } from "@angular/common/http";
import { ImmutableRetentionRule, RuleMetadate } from "../tag-retention/retention";
import { Observable, throwError as observableThrowError } from "rxjs";
import { map, catchError } from "rxjs/operators";
import { Project } from "../project";
import { HTTP_JSON_OPTIONS } from "@harbor/ui";
@Injectable()
export class ImmutableTagService {
    private I18nMap: object = {
        "retain": "ACTION_RETAIN",
        "lastXDays": "RULE_NAME_1",
        "latestActiveK": "RULE_NAME_2",
        "latestPushedK": "RULE_NAME_3",
        "latestPulledN": "RULE_NAME_4",
        "always": "RULE_NAME_5",
        "nDaysSinceLastPull": "RULE_NAME_6",
        "nDaysSinceLastPush": "RULE_NAME_7",
        "the images from the last # days": "RULE_TEMPLATE_1",
        "the most recent active # images": "RULE_TEMPLATE_2",
        "the most recently pushed # images": "RULE_TEMPLATE_3",
        "the most recently pulled # images": "RULE_TEMPLATE_4",
        "pulled within the last # days": "RULE_TEMPLATE_6",
        "pushed within the last # days": "RULE_TEMPLATE_7",
        "repoMatches": "MAT",
        "repoExcludes": "EXC",
        "matches": "MAT",
        "excludes": "EXC",
        "withLabels": "WITH",
        "withoutLabels": "WITHOUT",
        "COUNT": "UNIT_COUNT",
        "DAYS": "UNIT_DAY",
        "none": "NONE",
        "nothing": "NONE",
        "Parameters nDaysSinceLastPull is too large": "DAYS_LARGE",
        "Parameters nDaysSinceLastPush is too large": "DAYS_LARGE",
        "Parameters latestPushedK is too large": "COUNT_LARGE",
        "Parameters latestPulledN is too large": "COUNT_LARGE"
    };

    constructor(
        private http: HttpClient,
    ) {
    }

    getI18nKey(str: string): string {
        if (this.I18nMap[str.trim()]) {
            return "TAG_RETENTION." + this.I18nMap[str.trim()];
        }
        return str;
    }

    getRetentionMetadata(): Observable<RuleMetadate> {
        return this.http.get(`/api/retentions/metadatas`)
            .pipe(map(response => response as RuleMetadate))
            .pipe(catchError(error => observableThrowError(error)));
    }

    getRules(projectId): Observable<ImmutableRetentionRule[]> {
        return this.http.get(`/api/projects/${projectId}/immutabletagrules`)
            .pipe(map(response => response as ImmutableRetentionRule[]))
            .pipe(catchError(error => observableThrowError(error)));
    }

    createRule(projectId: number, retention: ImmutableRetentionRule) {
        return this.http.post(`/api/projects/${projectId}/immutabletagrules`, retention)
            .pipe(catchError(error => observableThrowError(error)));
    }

    updateRule(projectId, immutabletagrule: ImmutableRetentionRule) {
        return this.http.put(`/api/projects/${projectId}/immutabletagrules/${immutabletagrule.id}`, immutabletagrule)
            .pipe(catchError(error => observableThrowError(error)));
    }
    deleteRule(projectId, ruleId) {

        return this.http.delete(`/api/projects/${projectId}/immutabletagrules/${ruleId}`, HTTP_JSON_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }

    getProjectInfo(projectId) {
        return this.http.get(`/api/projects/${projectId}`)
            .pipe(map(response => response as Project))
            .pipe(catchError(error => observableThrowError(error)));
    }
}


