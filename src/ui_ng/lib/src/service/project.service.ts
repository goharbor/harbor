import { Observable } from 'rxjs/Observable';
import { Injectable, Inject } from '@angular/core';
import 'rxjs/add/observable/of';
import { Http, Headers, RequestOptions } from '@angular/http';
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';

import { Project } from '../project-policy-config/project';
import { ProjectPolicy } from '../project-policy-config/project-policy-config.component';
import {HTTP_JSON_OPTIONS, HTTP_GET_OPTIONS} from "../utils";

/**
 * Define the service methods to handle the Prject related things.
 *
 * @export
 * @abstract
 * @class ProjectService
 */
export abstract class ProjectService {
  /**
     * Get Infomations of a specific Project.
     *
     * @abstract
     * @param {string|number} [projectId]
     * @returns {(Observable<Project> | Promise<Project> | Project)}
     *
     * @memberOf ProjectService
     */
    abstract getProject(projectId: number | string): Observable<Project> | Promise<Project> | Project;

    /**
     * Update the specified project.
     *
     * @abstract
     * @param {(number | string)} projectId
     * @param {ProjectPolicy} projectPolicy
     * @returns {(Observable<any> | Promise<any> | any)}
     *
     * @memberOf EndpointService
     */
    abstract updateProjectPolicy(projectId: number | string,  projectPolicy: ProjectPolicy): Observable<any> | Promise<any> | any;
}

/**
 * Implement default service for project.
 *
 * @export
 * @class ProjectDefaultService
 * @extends {ProjectService}
 */
@Injectable()
export class ProjectDefaultService extends ProjectService {

  constructor(
      private http: Http,
      @Inject(SERVICE_CONFIG) private config: IServiceConfig
  ) {
      super();
  }

  public getProject(projectId: number | string): Observable<Project> | Promise<Project> | Project {
    if (!projectId) {
      return Promise.reject('Bad argument');
    }

    return this.http
                .get(`/api/projects/${projectId}`, HTTP_GET_OPTIONS)
                .map(response => response.json())
                .catch(error => Observable.throw(error));
  }

  public updateProjectPolicy(projectId: number | string, projectPolicy: ProjectPolicy): any {
    return this.http
              .put(`/api/projects/${projectId}`, { 'metadata': {
                'public': projectPolicy.Public ? 'true' : 'false',
                'enable_content_trust': projectPolicy.ContentTrust ? 'true' : 'false',
                'prevent_vul': projectPolicy.PreventVulImg ? 'true' : 'false',
                'severity': projectPolicy.PreventVulImgSeverity,
                'auto_scan': projectPolicy.ScanImgOnPush ? 'true' : 'false'
              } }, HTTP_JSON_OPTIONS)
              .map(response => response.status)
              .catch(error => Observable.throw(error));
  }
}
