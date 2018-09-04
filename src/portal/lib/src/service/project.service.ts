import { Observable } from "rxjs/Observable";
import { Injectable, Inject } from "@angular/core";
import { Http } from "@angular/http";
import "rxjs/add/observable/of";
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/map';
import 'rxjs/add/observable/throw';
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
  abstract getProject(
    projectId: number | string
  ): Observable<Project> | Promise<Project> | Project;

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
  abstract updateProjectPolicy(
    projectId: number | string,
    projectPolicy: ProjectPolicy
  ): Observable<any> | Promise<any> | any;

  /**
   * Get all projects
   *
   * @abstract
   * @param {string} name
   * @param {number} isPublic
   * @param {number} page
   * @param {number} pageSize
   * @returns {(Observable<any> | Promise<any> | any)}
   *
   * @memberOf EndpointService
   */
  abstract listProjects(
    name: string,
    isPublic: number,
    page?: number,
    pageSize?: number
  ): Observable<Project[]> | Promise<Project[]> | Project[];
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

  public getProject(
    projectId: number | string
  ): Observable<Project> | Promise<Project> | Project {
    if (!projectId) {
      return Promise.reject("Bad argument");
    }
    let baseUrl: string = this.config.projectBaseEndpoint
      ? this.config.projectBaseEndpoint
      : "/api/projects";
    return this.http
      .get(`${baseUrl}/${projectId}`, HTTP_GET_OPTIONS)
      .map(response => response.json())
      .catch(error => Observable.throw(error));
  }

  public listProjects(
    name: string,
    isPublic: number,
    page?: number,
    pageSize?: number
  ): Observable<Project[]> | Promise<Project[]> | Project[] {
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
      .map(response => response.json())
      .catch(error => Observable.throw(error));
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
      .map(response => response.status)
      .catch(error => Observable.throw(error));
  }
}
