import { Observable } from 'rxjs/Observable';
import { RequestQueryParams } from './RequestQueryParams';
import { Repository, RepositoryItem } from './interface';
import { Injectable, Inject } from '@angular/core';
import 'rxjs/add/observable/of';
import { Http } from '@angular/http';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { buildHttpRequestOptions, HTTP_JSON_OPTIONS } from '../utils';

/**
 * Define service methods for handling the repository related things.
 * Loose couple with project module.
 *
 * @export
 * @abstract
 * @class RepositoryService
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
     * @param {(number | string)} projectId
     * @param {string} repositoryName
     * @param {RequestQueryParams} [queryParams]
     * @returns {(Observable<Repository> | Promise<Repository> | Repository)}
     *
     * @memberOf RepositoryService
     */
    abstract getRepositories(projectId: number | string, repositoryName?: string, queryParams?: RequestQueryParams):
    Observable<Repository> | Promise<Repository> | Repository;

    /**
     * Update description of specified repository.
     *
     * @abstract
     * @param {number | string} projectId
     * @param {string} repoName
     * @returns {(Observable<Repository> | Promise<Repository> | Repository)}
     *
     * @memberOf RepositoryService
     */
    abstract updateRepositoryDescription(repoName: string, description: string): Observable<any> | Promise<any> | any;

    /**
     * DELETE the specified repository.
     *
     * @abstract
     * @param {string} repositoryName
     * @returns {(Observable<any> | Promise<any> | any)}
     *
     * @memberOf RepositoryService
     */
    abstract deleteRepository(repositoryName: string): Observable<any> | Promise<any> | any;
}

/**
 * Implement default service for repository.
 *
 * @export
 * @class RepositoryDefaultService
 * @extends {RepositoryService}
 */
@Injectable()
export class RepositoryDefaultService extends RepositoryService {
    constructor(
        private http: Http,
        @Inject(SERVICE_CONFIG) private config: IServiceConfig
    ) {
        super();
    }

    public getRepositories(projectId: number | string, repositoryName?: string, queryParams?: RequestQueryParams):
    Observable<Repository> | Promise<Repository> | Repository {
        if (!projectId) {
            return Promise.reject('Bad argument');
        }

        if (!queryParams) {
            queryParams = new RequestQueryParams();
        }

        queryParams.set('project_id', '' + projectId);
        if (repositoryName && repositoryName.trim() !== '') {
            queryParams.set('q', repositoryName);
        }

        let url: string = this.config.repositoryBaseEndpoint ? this.config.repositoryBaseEndpoint : '/api/repositories';
        return this.http.get(url, buildHttpRequestOptions(queryParams)).toPromise()
            .then(response => {
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

                result.data = response.json() as RepositoryItem[];

                if (result.metadata.xTotalCount === 0) {
                    if (result.data && result.data.length > 0) {
                        result.metadata.xTotalCount = result.data.length;
                    }
                }

                return result;
            })
            .catch(error => Promise.reject(error));
    }

    public updateRepositoryDescription(repositoryName: string, description: string,
         queryParams?: RequestQueryParams): Observable<any> | Promise<any> | any {

        if (!queryParams) {
            queryParams = new RequestQueryParams();
        }

        let baseUrl: string = this.config.repositoryBaseEndpoint ? this.config.repositoryBaseEndpoint : '/api/repositories';
        let url = `${baseUrl}/${repositoryName}`;
        return this.http.put(url, {'description': description }, HTTP_JSON_OPTIONS).toPromise()
        .then(response => response)
        .catch(error => Promise.reject(error));
      }

    public deleteRepository(repositoryName: string): Observable<any> | Promise<any> | any {
        if (!repositoryName) {
            return Promise.reject('Bad argument');
        }
        let url: string = this.config.repositoryBaseEndpoint ? this.config.repositoryBaseEndpoint : '/api/repositories';
        url = `${url}/${repositoryName}`;

        return this.http.delete(url, HTTP_JSON_OPTIONS).toPromise()
            .then(response => response)
            .catch(error => {return Promise.reject(error); });
    }
}
