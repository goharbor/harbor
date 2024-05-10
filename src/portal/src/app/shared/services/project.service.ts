import { throwError as observableThrowError, Observable, of } from 'rxjs';
import { Injectable } from '@angular/core';
import { HttpClient, HttpParams, HttpResponse } from '@angular/common/http';
import { catchError, share } from 'rxjs/operators';
import { Project } from '../../base/project/project-config/project-policy-config/project';
import { ProjectPolicy } from '../../base/project/project-config/project-policy-config/project-policy-config.component';
import {
    HTTP_JSON_OPTIONS,
    HTTP_GET_OPTIONS,
    buildHttpRequestOptionsWithObserveResponse,
    CURRENT_BASE_HREF,
} from '../units/utils';
import { CacheObservable } from '../units/cache-util';

/**
 * Define the service methods to handle the Project related things.
 *
 **
 * @abstract
 * class ProjectService
 */
export abstract class ProjectService {
    /**
     * Get Informations about a specific Project.
     *
     * @abstract
     *  ** deprecated param {string|number} [projectId]
     * returns {(Observable<Project> )}
     *
     * @memberOf ProjectService
     */
    abstract getProject(projectId: number | string): Observable<Project>;

    /**
     * Get info about a specific Project from cache, if no cache, get it from the back end then store it in cache.
     *
     * @abstract
     *  ** deprecated param {string|number} [projectId]
     * returns {(Observable<Project> )}
     *
     * @memberOf ProjectService
     */
    abstract getProjectFromCache(
        projectId: number | string
    ): Observable<Project>;

    /**
     * Update the specified project.
     *
     * @abstract
     *  ** deprecated param {(number | string)} projectId
     *  ** deprecated param {ProjectPolicy} projectPolicy
     * returns {(Observable<any>)}
     *
     * @memberOf EndpointService
     */
    abstract updateProjectPolicy(
        projectId: number | string,
        projectPolicy: ProjectPolicy,
        reuseSysCVEVAllowlist: string,
        projectAllowlist: object
    ): Observable<any>;

    /**
     * Get all projects
     *
     * @abstract
     *  ** deprecated param {string} name
     *  ** deprecated param {number} isPublic
     *  ** deprecated param {number} page
     *  ** deprecated param {number} pageSize
     * returns {(Observable<any>)}
     *
     * @memberOf EndpointService
     */
    abstract listProjects(
        name: string,
        isPublic?: number,
        page?: number,
        pageSize?: number,
        sort?: string
    ): Observable<HttpResponse<Project[]>>;
    abstract createProject(
        name: string,
        metadata: any,
        storageLimit: number,
        registryId: number
    ): Observable<any>;
    abstract deleteProject(projectId: number): Observable<any>;
    abstract checkProjectExists(projectName: string): Observable<any>;
    abstract checkProjectMember(projectId: number): Observable<any>;
    abstract getProjectSummary(projectId: number): Observable<any>;
}

/**
 * Implement default service for project.
 *
 **
 * class ProjectDefaultService
 * extends {ProjectService}
 */
@Injectable()
export class ProjectDefaultService extends ProjectService {
    // to avoid multiple requests for one navigating action
    private _sharedProjectObservableMap: {
        [key: number | string]: Observable<Project>;
    } = {};
    constructor(private http: HttpClient) {
        super();
    }

    public getProject(projectId: number | string): Observable<Project> {
        if (!projectId) {
            return observableThrowError('Bad argument');
        }
        let baseUrl: string = CURRENT_BASE_HREF + '/projects';
        return this.http
            .get<Project>(`${baseUrl}/${projectId}`, HTTP_GET_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }

    @CacheObservable({ maxAge: 1000 * 30 })
    public getProjectFromCache(
        projectId: number | string
    ): Observable<Project> {
        const baseUrl: string = CURRENT_BASE_HREF + '/projects';
        if (this._sharedProjectObservableMap[projectId]) {
            return this._sharedProjectObservableMap[projectId];
        }
        this._sharedProjectObservableMap[projectId] = this.http
            .get<Project>(`${baseUrl}/${projectId}`, HTTP_GET_OPTIONS)
            .pipe(share());
        return this._sharedProjectObservableMap[projectId];
    }

    public updateProjectPolicy(
        projectId: number | string,
        projectPolicy: ProjectPolicy,
        reuseSysCVEVAllowlist: string,
        projectAllowlist: object
    ): any {
        let baseUrl: string = CURRENT_BASE_HREF + '/projects';
        return this.http
            .put<any>(
                `${baseUrl}/${projectId}`,
                {
                    metadata: {
                        public: projectPolicy.Public ? 'true' : 'false',
                        enable_content_trust: projectPolicy.ContentTrust
                            ? 'true'
                            : 'false',
                        enable_content_trust_cosign:
                            projectPolicy.ContentTrustCosign ? 'true' : 'false',
                        prevent_vul: projectPolicy.PreventVulImg
                            ? 'true'
                            : 'false',
                        severity: projectPolicy.PreventVulImgSeverity,
                        auto_scan: projectPolicy.ScanImgOnPush
                            ? 'true'
                            : 'false',
                        auto_sbom_generation: projectPolicy.GenerateSbomOnPush
                            ? 'true'
                            : 'false',
                        reuse_sys_cve_allowlist: reuseSysCVEVAllowlist,
                    },
                    cve_allowlist: projectAllowlist,
                },
                HTTP_JSON_OPTIONS
            )
            .pipe(catchError(error => observableThrowError(error)));
    }
    public listProjects(
        name: string,
        isPublic?: number,
        page?: number,
        pageSize?: number,
        sort?: string
    ): Observable<HttpResponse<Project[]>> {
        let params = new HttpParams();
        if (page && pageSize) {
            params = params
                .set('page', page + '')
                .set('page_size', pageSize + '');
        }
        if (name && name.trim() !== '') {
            params = params.set('name', name);
        }
        if (isPublic !== undefined) {
            params = params.set('public', '' + isPublic);
        }
        if (sort) {
            params = params.set('sort', sort);
        }
        return this.http
            .get<HttpResponse<Project[]>>(
                `${CURRENT_BASE_HREF}/projects`,
                buildHttpRequestOptionsWithObserveResponse(params)
            )
            .pipe(catchError(error => observableThrowError(error)));
    }

    public createProject(
        name: string,
        metadata: any,
        storageLimit: number,
        registryId: number
    ): Observable<any> {
        return this.http
            .post(
                `${CURRENT_BASE_HREF}/projects`,
                JSON.stringify({
                    project_name: name,
                    registry_id: +registryId,
                    metadata: {
                        public: metadata.public ? 'true' : 'false',
                    },
                    storage_limit: storageLimit,
                }),
                HTTP_JSON_OPTIONS
            )
            .pipe(catchError(error => observableThrowError(error)));
    }

    public deleteProject(projectId: number): Observable<any> {
        return this.http
            .delete(`${CURRENT_BASE_HREF}/projects/${projectId}`)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public checkProjectExists(projectName: string): Observable<any> {
        return this.http
            .head(`${CURRENT_BASE_HREF}/projects/?project_name=${projectName}`)
            .pipe(
                catchError(error => {
                    if (error && error.status === 404) {
                        return of(error);
                    }
                    return observableThrowError(error);
                })
            );
    }

    public checkProjectMember(projectId: number): Observable<any> {
        return this.http
            .get(
                `${CURRENT_BASE_HREF}/projects/${projectId}/members`,
                HTTP_GET_OPTIONS
            )
            .pipe(catchError(error => observableThrowError(error)));
    }
    public getProjectSummary(projectId: number): Observable<any> {
        return this.http
            .get(
                `${CURRENT_BASE_HREF}/projects/${projectId}/summary`,
                HTTP_GET_OPTIONS
            )
            .pipe(catchError(error => observableThrowError(error)));
    }
}
