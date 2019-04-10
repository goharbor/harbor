
import {throwError as observableThrowError,  Observable } from "rxjs";
import { Injectable, Inject } from "@angular/core";
import { Http } from "@angular/http";
import { map ,  catchError } from "rxjs/operators";

import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import { Project } from "../project-policy-config/project";
import { ProjectPolicy } from "../project-policy-config/project-policy-config.component";
import {
  HTTP_JSON_OPTIONS,
  HTTP_GET_OPTIONS,
  buildHttpRequestOptions
} from "../utils";
import { RequestQueryParams } from "./RequestQueryParams";

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
  abstract getProject(
    projectId: number | string
  ): Observable<Project> ;

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
    projectPolicy: ProjectPolicy
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
    isPublic: number,
    page?: number,
    pageSize?: number
  ): Observable<Project[]>;
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
  constructor(
    private http: Http,
    @Inject(SERVICE_CONFIG) private config: IServiceConfig
  ) {
    super();
  }

  public getProject(
    projectId: number | string
  ): Observable<Project> {
    if (!projectId) {
      return observableThrowError("Bad argument");
    }
    let baseUrl: string = this.config.projectBaseEndpoint
      ? this.config.projectBaseEndpoint
      : "/api/projects";
    return this.http
      .get(`${baseUrl}/${projectId}`, HTTP_GET_OPTIONS)
      .pipe(map(response => response.json()))
      .pipe(catchError(error => observableThrowError(error)));
  }

  public listProjects(
    name: string,
    isPublic: number,
    page?: number,
    pageSize?: number
  ): Observable<Project[]> {
    let baseUrl: string = this.config.projectBaseEndpoint
      ? this.config.projectBaseEndpoint
      : "/api/projects";
    let params = new RequestQueryParams();
    if (page && pageSize) {
      params.set("page", page + "");
      params.set("page_size", pageSize + "");
    }
    if (name && name.trim() !== "") {
      params.set("name", name);
    }
    if (isPublic !== undefined) {
      params.set("public", "" + isPublic);
    }

    // let options = new RequestOptions({ headers: this.getHeaders, search: params });
    return this.http
      .get(baseUrl, buildHttpRequestOptions(params))
      .pipe(map(response => response.json()))
      .pipe(catchError(error => observableThrowError(error)));
  }

  public updateProjectPolicy(
    projectId: number | string,
    projectPolicy: ProjectPolicy
  ): any {
    let baseUrl: string = this.config.projectBaseEndpoint
      ? this.config.projectBaseEndpoint
      : "/api/projects";
    return this.http
      .put(
        `${baseUrl}/${projectId}`,
        {
          metadata: {
            public: projectPolicy.Public ? "true" : "false",
            enable_content_trust: projectPolicy.ContentTrust ? "true" : "false",
            prevent_vul: projectPolicy.PreventVulImg ? "true" : "false",
            severity: projectPolicy.PreventVulImgSeverity,
            auto_scan: projectPolicy.ScanImgOnPush ? "true" : "false"
          }
        },
        HTTP_JSON_OPTIONS
      )
      .pipe(map(response => response.status))
      .pipe(catchError(error => observableThrowError(error)));
  }
}
