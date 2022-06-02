import { Observable, throwError as observableThrowError } from 'rxjs';
import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { catchError, map } from 'rxjs/operators';
import {
    HelmChartDetail,
    HelmChartItem,
    HelmChartVersion,
} from './helm-chart.interface.service';
import { RequestQueryParams } from '../../../../shared/services';
import {
    HTTP_GET_OPTIONS,
    HTTP_JSON_OPTIONS,
    V1_BASE_HREF,
} from '../../../../shared/units/utils';

/**
 * Define service methods for handling the helmchart related things.
 * Loose couple with project module.
 *
 **
 * @abstract
 * class HelmChartService
 */
export abstract class HelmChartService {
    /**
     * Get all helm charts info
     *  ** deprecated param projectName Id of the project
     *  ** deprecated param queryParams options params for query data
     */
    abstract getHelmCharts(
        projectName: string,
        queryParams?: RequestQueryParams
    ): Observable<HelmChartItem[]>;

    /**
     * Delete an helmchart
     *  ** deprecated param projectId Id of the project
     *  ** deprecated param chartId ID of helmChart in this specific project
     */
    abstract deleteHelmChart(
        projectId: number | string,
        chartName: string
    ): Observable<any>;

    /**
     * Get all the versions of helmchart
     *  ** deprecated param projectName Id of the project
     *  ** deprecated param chartName ID of the helm chart
     *  ** deprecated param queryParams option params for query
     */
    abstract getChartVersions(
        projectName: string,
        chartName: string
    ): Observable<HelmChartVersion[]>;

    /**
     * Delete a version of helmchart
     *  ** deprecated param projectName ID of the project
     *  ** deprecated param chartName ID of the chart you want to delete
     *  ** deprecated param version name of the version
     */
    abstract deleteChartVersion(
        projectName: string,
        chartName: string,
        version: string
    ): Observable<any>;

    /**
     * Get the all details of an helmchart
     *  ** deprecated param projectName ID of the project
     *  ** deprecated param chartname ID of the chart
     *  ** deprecated param version name of the chart's version
     *  ** deprecated param queryParams options
     */
    abstract getChartDetail(
        projectName: string,
        chartname: string,
        version: string
    ): Observable<HelmChartDetail>;

    /**
     * Download an specific verison
     *  ** deprecated param projectName ID of the project
     *  ** deprecated param filename ID of the helm chart
     *  ** deprecated param version Name of version
     *  ** deprecated param queryParams options
     */
    abstract downloadChart(
        projectName: string,
        filename: string
    ): Observable<any>;

    /**
     * Upload chart and prov files to chartmuseam
     *  ** deprecated param projectName Name of the project
     *  ** deprecated param chart chart file
     *  ** deprecated param prov prov file
     */
    abstract uploadChart(
        projectName: string,
        chart: File,
        prov: File
    ): Observable<any>;
}

/**
 * Implement default service for helm chart.
 */
@Injectable()
export class HelmChartDefaultService extends HelmChartService {
    constructor(private http: HttpClient) {
        super();
    }

    private handleErrorObservable(error: HttpErrorResponse) {
        return observableThrowError(error);
    }

    public getHelmCharts(projectName: string): Observable<HelmChartItem[]> {
        if (!projectName) {
            return observableThrowError(
                'Bad argument, No project id to get helm charts'
            );
        }

        return this.http
            .get<HelmChartItem[]>(
                `${V1_BASE_HREF + '/chartrepo'}/${projectName}/charts`,
                HTTP_GET_OPTIONS
            )
            .pipe(
                map(response => response || []),
                catchError(error => this.handleErrorObservable(error))
            );
    }

    public deleteHelmChart(
        projectId: number | string,
        chartName: string
    ): Observable<any> {
        if (!chartName) {
            observableThrowError('Bad argument');
        }

        return this.http
            .delete(
                `${
                    V1_BASE_HREF + '/chartrepo'
                }/${projectId}/charts/${chartName}`
            )
            .pipe(
                map(response => {
                    return response || [];
                })
            )
            .pipe(catchError(this.handleErrorObservable));
    }

    public getChartVersions(
        projectName: string,
        chartName: string
    ): Observable<HelmChartVersion[]> {
        return this.http
            .get<HelmChartVersion[]>(
                `${
                    V1_BASE_HREF + '/chartrepo'
                }/${projectName}/charts/${chartName}`,
                HTTP_GET_OPTIONS
            )
            .pipe(
                map(response => response || []),
                catchError(this.handleErrorObservable)
            );
    }

    public deleteChartVersion(
        projectName: string,
        chartName: string,
        version: string
    ): any {
        return this.http
            .delete(
                `${
                    V1_BASE_HREF + '/chartrepo'
                }/${projectName}/charts/${chartName}/${version}`,
                HTTP_JSON_OPTIONS
            )
            .pipe(
                map(response => {
                    return response || [];
                })
            )
            .pipe(catchError(this.handleErrorObservable));
    }

    public getChartDetail(
        projectName: string,
        chartName: string,
        version: string
    ): Observable<HelmChartDetail> {
        return this.http
            .get<HelmChartDetail>(
                `${
                    V1_BASE_HREF + '/chartrepo'
                }/${projectName}/charts/${chartName}/${version}`
            )
            .pipe(catchError(this.handleErrorObservable));
    }

    public downloadChart(
        projectName: string,
        filename: string
    ): Observable<any> {
        let url: string;
        let chartFileRegexPattern = new RegExp('^http.*/chartrepo/(.*)');
        if (chartFileRegexPattern.test(filename)) {
            let match = filename.match('^http.*/chartrepo/(.*)');
            url = `${DOWNLOAD_CHART_ENDPOINT}/${match[1]}`;
        } else {
            url = `${DOWNLOAD_CHART_ENDPOINT}/${projectName}/${filename}`;
        }
        return this.http
            .get(url, {
                responseType: 'blob',
            })
            .pipe(
                map(response => {
                    let parts = filename.split('/');
                    return {
                        filename: parts[parts.length - 1],
                        data: response,
                    };
                })
            )
            .pipe(catchError(this.handleErrorObservable));
    }

    public uploadChart(
        projectName: string,
        chart?: File,
        prov?: File
    ): Observable<any> {
        let formData = new FormData();
        let uploadURL = `${V1_BASE_HREF + '/chartrepo'}/${projectName}/charts`;
        if (chart) {
            formData.append('chart', chart);
        }
        if (prov) {
            formData.append('prov', prov);
            if (!chart) {
                uploadURL = `${
                    V1_BASE_HREF + '/chartrepo'
                }/${projectName}/prov`;
            }
        }
        return this.http
            .post(uploadURL, formData, {
                responseType: 'json',
            })
            .pipe(map(response => response || []))
            .pipe(catchError(this.handleErrorObservable));
    }
}

export const DOWNLOAD_CHART_ENDPOINT: string = '/chartrepo';
