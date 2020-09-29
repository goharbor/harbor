import { Injectable } from '@angular/core';
import { HttpClient } from "@angular/common/http";
import { ImmutableRetentionRule, RuleMetadate } from "../tag-retention/retention";
import { Observable, throwError as observableThrowError } from "rxjs";
import { map, catchError } from "rxjs/operators";
import { Project } from "../../project";
import { CURRENT_BASE_HREF, HTTP_JSON_OPTIONS } from "../../../../lib/utils/utils";


@Injectable()
export class ImmutableTagService {
    private I18nMap: object = {
        "repoMatches": "MAT",
        "repoExcludes": "EXC",
        "matches": "MAT",
        "excludes": "EXC",
        "withLabels": "WITH",
        "withoutLabels": "WITHOUT",
        "none": "NONE",
        "nothing": "NONE"
    };

    constructor(
        private http: HttpClient,
    ) {
    }

    getI18nKey(str: string): string {
        if (this.I18nMap[str.trim()]) {
            return "IMMUTABLE_TAG." + this.I18nMap[str.trim()];
        }
        return str;
    }

    getRetentionMetadata(): Observable<RuleMetadate> {
        return this.http.get(`${ CURRENT_BASE_HREF }/retentions/metadatas`)
            .pipe(map(response => response as RuleMetadate))
            .pipe(catchError(error => observableThrowError(error)));
    }

    getRules(projectId): Observable<ImmutableRetentionRule[]> {
        return this.http.get(`${ CURRENT_BASE_HREF }/projects/${projectId}/immutabletagrules`)
            .pipe(map(response => response as ImmutableRetentionRule[]))
            .pipe(catchError(error => observableThrowError(error)));
    }

    createRule(projectId: number, retention: ImmutableRetentionRule) {
        return this.http.post(`${ CURRENT_BASE_HREF }/projects/${projectId}/immutabletagrules`, retention)
            .pipe(catchError(error => observableThrowError(error)));
    }

    updateRule(projectId, immutabletagrule: ImmutableRetentionRule) {
        return this.http.put(`${ CURRENT_BASE_HREF }/projects/${projectId}/immutabletagrules/${immutabletagrule.id}`, immutabletagrule)
            .pipe(catchError(error => observableThrowError(error)));
    }
    deleteRule(projectId, ruleId) {

        return this.http.delete(`${ CURRENT_BASE_HREF }/projects/${projectId}/immutabletagrules/${ruleId}`, HTTP_JSON_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }

    getProjectInfo(projectId) {
        return this.http.get(`${ CURRENT_BASE_HREF }/projects/${projectId}`)
            .pipe(map(response => response as Project))
            .pipe(catchError(error => observableThrowError(error)));
    }
}


