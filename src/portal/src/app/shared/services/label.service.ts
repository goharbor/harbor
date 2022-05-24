import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { catchError } from 'rxjs/operators';
import { RequestQueryParams } from './RequestQueryParams';
import { Label } from './interface';
import {
    buildHttpRequestOptions,
    CURRENT_BASE_HREF,
    V1_BASE_HREF,
    HTTP_JSON_OPTIONS,
} from '../units/utils';
import { Observable, throwError as observableThrowError } from 'rxjs';

export abstract class LabelService {
    abstract getGLabels(
        name?: string,
        queryParams?: RequestQueryParams
    ): Observable<Label[]>;

    abstract getPLabels(
        projectId: number,
        name?: string,
        queryParams?: RequestQueryParams
    ): Observable<Label[]>;

    abstract getProjectLabels(
        projectId: number,
        name?: string,
        queryParams?: RequestQueryParams
    ): Observable<Label[]>;

    abstract getLabels(
        scope: string,
        projectId?: number,
        name?: string,
        queryParams?: RequestQueryParams
    ): Observable<Label[]>;

    abstract createLabel(label: Label): Observable<Label>;

    abstract getLabel(id: number): Observable<Label>;

    abstract updateLabel(id: number, param: Label): Observable<any>;

    abstract deleteLabel(id: number): Observable<any>;

    abstract getChartVersionLabels(
        projectName: string,
        chartName: string,
        version?: string
    ): Observable<Label[]>;

    abstract markChartLabel(
        projectName: string,
        chartName: string,
        version: string,
        label: Label
    ): Observable<any>;

    abstract unmarkChartLabel(
        projectName: string,
        chartName: string,
        version: string,
        label: Label
    ): Observable<any>;
}

@Injectable()
export class LabelDefaultService extends LabelService {
    labelUrl: string;
    chartUrl: string;
    chartLabelUrl: string;

    constructor(private http: HttpClient) {
        super();
        this.labelUrl = CURRENT_BASE_HREF + '/labels';
        this.chartUrl = V1_BASE_HREF + '/chartrepo';
        this.chartLabelUrl = CURRENT_BASE_HREF + '/chartrepo';
    }

    getGLabels(
        name?: string,
        queryParams?: RequestQueryParams
    ): Observable<Label[]> {
        if (!queryParams) {
            queryParams = new RequestQueryParams();
        }
        queryParams = queryParams.set('scope', 'g');

        if (name) {
            queryParams = queryParams.set('name', '' + name);
        }
        return this.http
            .get<Label[]>(this.labelUrl, buildHttpRequestOptions(queryParams))
            .pipe(catchError(error => observableThrowError(error)));
    }

    getPLabels(
        projectId: number,
        name?: string,
        queryParams?: RequestQueryParams
    ): Observable<Label[]> {
        if (!queryParams) {
            queryParams = new RequestQueryParams();
        }
        queryParams = queryParams.set('scope', 'p');
        if (projectId) {
            queryParams = queryParams.set('project_id', '' + projectId);
        }
        if (name) {
            queryParams = queryParams.set('name', '' + name);
        }
        return this.http
            .get<Label[]>(this.labelUrl, buildHttpRequestOptions(queryParams))
            .pipe(catchError(error => observableThrowError(error)));
    }

    getProjectLabels(
        projectId: number,
        name?: string,
        queryParams?: RequestQueryParams
    ): Observable<Label[]> {
        if (!queryParams) {
            queryParams = new RequestQueryParams();
        }
        queryParams = queryParams.set('scope', 'p');
        if (projectId) {
            queryParams = queryParams.set('project_id', '' + projectId);
        }
        if (name) {
            queryParams = queryParams.set('name', '' + name);
        }
        return this.http.get<Label[]>(
            this.labelUrl,
            buildHttpRequestOptions(queryParams)
        );
    }

    getLabels(
        scope: string,
        projectId?: number,
        name?: string,
        queryParams?: RequestQueryParams
    ): Observable<Label[]> {
        if (!queryParams) {
            queryParams = queryParams = new RequestQueryParams();
        }
        if (scope) {
            queryParams = queryParams.set('scope', scope);
        }
        if (projectId) {
            queryParams = queryParams.set('project_id', '' + projectId);
        }
        if (name) {
            queryParams = queryParams.set('name', '' + name);
        }
        return this.http
            .get<Label[]>(this.labelUrl, buildHttpRequestOptions(queryParams))
            .pipe(catchError(error => observableThrowError(error)));
    }

    createLabel(label: Label): Observable<any> {
        if (!label) {
            return observableThrowError('Invalid label.');
        }
        return this.http
            .post<any>(this.labelUrl, JSON.stringify(label), HTTP_JSON_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }

    getLabel(id: number): Observable<Label> {
        if (!id || id <= 0) {
            return observableThrowError('Bad request argument.');
        }
        let reqUrl = `${this.labelUrl}/${id}`;
        return this.http
            .get<Label>(reqUrl, HTTP_JSON_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }

    updateLabel(id: number, label: Label): Observable<any> {
        if (!id || id <= 0) {
            return observableThrowError('Bad request argument.');
        }
        if (!label) {
            return observableThrowError('Invalid endpoint.');
        }
        let reqUrl = `${this.labelUrl}/${id}`;
        return this.http
            .put<any>(reqUrl, JSON.stringify(label), HTTP_JSON_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }

    deleteLabel(id: number): Observable<any> {
        if (!id || id <= 0) {
            return observableThrowError('Bad request argument.');
        }
        let reqUrl = `${this.labelUrl}/${id}`;
        return this.http
            .delete<any>(reqUrl)
            .pipe(catchError(error => observableThrowError(error)));
    }

    getChartVersionLabels(
        projectName: string,
        chartName: string,
        version: string
    ): Observable<Label[]> {
        return this.http.get<Label[]>(
            `${this.chartLabelUrl}/${projectName}/charts/${chartName}/${version}/labels`
        );
    }

    markChartLabel(
        projectName: string,
        chartName: string,
        version: string,
        label: Label
    ): Observable<any> {
        return this.http.post(
            `${this.chartLabelUrl}/${projectName}/charts/${chartName}/${version}/labels`,
            JSON.stringify(label),
            HTTP_JSON_OPTIONS
        );
    }

    unmarkChartLabel(
        projectName: string,
        chartName: string,
        version: string,
        label: Label
    ): Observable<any> {
        return this.http.delete(
            `${this.chartLabelUrl}/${projectName}/charts/${chartName}/${version}/labels/${label.id}`,
            HTTP_JSON_OPTIONS
        );
    }
}
