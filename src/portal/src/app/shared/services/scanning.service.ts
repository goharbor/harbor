import { HttpClient } from '@angular/common/http';
import { Injectable, Inject } from '@angular/core';
import { CURRENT_BASE_HREF, HTTP_JSON_OPTIONS } from '../units/utils';
import { RequestQueryParams } from './RequestQueryParams';
import { VulnerabilitySummary } from './interface';
import { map, catchError } from 'rxjs/operators';
import { Observable, of, throwError as observableThrowError } from 'rxjs';

/**
 * Get the vulnerabilities scanning results for the specified tag.
 *
 **
 * @abstract
 * class ScanningResultService
 */
export abstract class ScanningResultService {
    /**
     * Get the summary of vulnerability scanning result.
     *
     * @abstract
     *  ** deprecated param {string} tagId
     * returns {(Observable<VulnerabilitySummary>)}
     *
     * @memberOf ScanningResultService
     */
    abstract getVulnerabilityScanningSummary(
        repoName: string,
        tagId: string,
        queryParams?: RequestQueryParams
    ): Observable<VulnerabilitySummary>;
    /**
     * Start a new vulnerability scanning
     *
     * @abstract
     *  ** deprecated param {string} repoName
     *  ** deprecated param {string} tagId
     * returns {(Observable<any>)}
     *
     * @memberOf ScanningResultService
     */
    abstract startVulnerabilityScanning(
        projectName: string,
        repoName: string,
        artifactDigest: string
    ): Observable<any>;

    /**
     * Trigger the scanning all action.
     *
     * @abstract
     * returns {(Observable<any>)}
     *
     * @memberOf ScanningResultService
     */
    abstract startScanningAll(): Observable<any>;

    /**
     *  Get scanner metadata
     * @param uuid
     * @memberOf ScanningResultService
     */
    abstract getScannerMetadata(uuid: string): Observable<any>;

    /**
     *  Get project scanner
     * @param projectId
     */
    abstract getProjectScanner(projectId: number): Observable<any>;
}

@Injectable()
export class ScanningResultDefaultService extends ScanningResultService {
    _baseUrl: string = CURRENT_BASE_HREF + '/projects';

    constructor(private http: HttpClient) {
        super();
        this._baseUrl = CURRENT_BASE_HREF + '/repositories';
    }

    getVulnerabilityScanningSummary(
        repoName: string,
        tagId: string,
        queryParams?: RequestQueryParams
    ): Observable<VulnerabilitySummary> {
        if (
            !repoName ||
            repoName.trim() === '' ||
            !tagId ||
            tagId.trim() === ''
        ) {
            return observableThrowError('Bad argument');
        }

        return of({} as VulnerabilitySummary);
    }
    startVulnerabilityScanning(
        projectName: string,
        repoName: string,
        artifactDigest: string
    ): Observable<any> {
        if (
            !repoName ||
            repoName.trim() === '' ||
            !artifactDigest ||
            artifactDigest.trim() === ''
        ) {
            return observableThrowError('Bad argument');
        }

        return this.http
            .post(
                `${CURRENT_BASE_HREF}/projects//${projectName}/repositories/${repoName}/artifacts/${artifactDigest}/scan`,
                HTTP_JSON_OPTIONS
            )
            .pipe(
                map(() => {
                    return true;
                }),
                catchError(error => observableThrowError(error))
            );
    }

    startScanningAll(): Observable<any> {
        return this.http
            .post(`${this._baseUrl}/scanAll`, HTTP_JSON_OPTIONS)
            .pipe(
                map(() => {
                    return true;
                }),
                catchError(error => observableThrowError(error))
            );
    }
    getScannerMetadata(uuid: string): Observable<any> {
        return this.http
            .get(`${CURRENT_BASE_HREF}/scanners/${uuid}/metadata`)
            .pipe(map(response => response as any))
            .pipe(catchError(error => observableThrowError(error)));
    }
    getProjectScanner(projectId: number): Observable<any> {
        return this.http
            .get(`${CURRENT_BASE_HREF}/projects/${projectId}/scanner`)
            .pipe(map(response => response as any))
            .pipe(catchError(error => observableThrowError(error)));
    }
}
