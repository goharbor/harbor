import { Injectable, Inject } from '@angular/core';
import { HttpClient, HttpResponse } from '@angular/common/http';
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
import { Repository, RepositoryItem, RequestQueryParams } from "../../../lib/services";
import { IServiceConfig, SERVICE_CONFIG } from "../../../lib/entities/service.config";
import { buildHttpRequestOptionsWithObserveResponse, HTTP_JSON_OPTIONS } from "../../../lib/utils/utils";

/**
 * Define service methods for handling the repository related things.
 * Loose couple with project module.
 *
 **
 * @abstract
 * class RepositoryService
 */
export abstract class RepositoryService {
    /**
     * List all the repositories in the specified project.
     * Specify the 'repositoryName' to only return the repositories which match the name pattern.
     * If pagination needed, set the following parameters in queryParams:
     *   'page': current page,
     *   'page_size': page size.
     *
     * @abstract
     *  ** deprecated param {(number | string)} projectId
     *  ** deprecated param {string} repositoryName
     *  ** deprecated param {RequestQueryParams} [queryParams]
     * returns {(Observable<Repository>)}
     *
     * @memberOf RepositoryService
     */
    abstract getRepositories(projectId: number | string, repositoryName?: string, queryParams?: RequestQueryParams):
        Observable<Repository>;

    /**
     * Update description of specified repository.
     *
     * @abstract
     *  ** deprecated param {number | string} projectId
     *  ** deprecated param {string} repoName
     * returns {(Observable<Repository>)}
     *
     * @memberOf RepositoryService
     */
    abstract updateRepositoryDescription(repoName: string, description: string): Observable<any>;

    /**
     * DELETE the specified repository.
     *
     * @abstract
     *  ** deprecated param {string} repositoryName
     * returns {(Observable<any>)}
     *
     * @memberOf RepositoryService
     */
    abstract deleteRepository(repositoryName: string): Observable<any>;
}

/**
 * Implement default service for repository.
 *
 **
 * class RepositoryDefaultService
 * extends {RepositoryService}
 */
@Injectable()
export class RepositoryDefaultService extends RepositoryService {
    constructor(
        private http: HttpClient,
        @Inject(SERVICE_CONFIG) private config: IServiceConfig
    ) {
        super();
    }
    public getRepositories(projectId: number | string, repositoryName?: string, queryParams?: RequestQueryParams):
        Observable<Repository> {
        if (!projectId) {
            return observableThrowError('Bad argument');
        }

        if (!queryParams) {
            queryParams = new RequestQueryParams();
        }

        queryParams = queryParams.set('project_id', '' + projectId);
        if (repositoryName && repositoryName.trim() !== '') {
            queryParams = queryParams.set('q', repositoryName);
        }
        let url: string = this.config.repositoryBaseEndpoint ? this.config.repositoryBaseEndpoint : '/api/repositories';
        return this.http.get<HttpResponse<RepositoryItem[]>>(url, buildHttpRequestOptionsWithObserveResponse(queryParams))
            .pipe(map(response => {
                let result: Repository = {
                    metadata: { xTotalCount: 0 },
                    data: []
                };

                if (response && response.headers) {
                    let xHeader: string = response.headers.get('X-Total-Count');
                    if (xHeader) {
                        result.metadata.xTotalCount = parseInt(xHeader, 0);
                    }
                }

                result.data = response.body as RepositoryItem[];

                if (result.metadata.xTotalCount === 0) {
                    if (result.data && result.data.length > 0) {
                        result.metadata.xTotalCount = result.data.length;
                    }
                }

                return result;
            })
                , catchError(error => {
                    return observableThrowError(error);
                }));
    }

    public updateRepositoryDescription(repositoryName: string, description: string,
        queryParams?: RequestQueryParams): Observable<any> {

        if (!queryParams) {
            queryParams = new RequestQueryParams();
        }

        let baseUrl: string = this.config.repositoryBaseEndpoint ? this.config.repositoryBaseEndpoint : '/api/repositories';
        let url = `${baseUrl}/${repositoryName}`;
        return this.http.put(url, { 'description': description }, HTTP_JSON_OPTIONS)
            .pipe(map(response => response)
                , catchError(error => observableThrowError(error)));
    }

    public deleteRepository(repositoryName: string): Observable<any> {
        if (!repositoryName) {
            return observableThrowError('Bad argument');
        }
        let url: string = this.config.repositoryBaseEndpoint ? this.config.repositoryBaseEndpoint : '/api/repositories';
        url = `${url}/${repositoryName}`;

        return this.http.delete(url, HTTP_JSON_OPTIONS)
            .pipe(map(response => response)
                , catchError(error => observableThrowError(error)));
    }
}
