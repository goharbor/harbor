import { Injectable, Inject } from "@angular/core";
import { HttpClient, HttpResponse } from "@angular/common/http";
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError, Subject } from "rxjs";
import { Manifest, RequestQueryParams } from "../../../../lib/services";
import { IServiceConfig, SERVICE_CONFIG } from "../../../../lib/entities/service.config";
import {
  buildHttpRequestOptionsWithObserveResponse,
  HTTP_GET_OPTIONS,
  HTTP_JSON_OPTIONS
} from "../../../../lib/utils/utils";
import { Artifact } from "./artifact";


/**
 * Define the service methods to handle the repository tag related things.
 *
 **
 * @abstract
 * class TagService
 */
export abstract class ArtifactService {
  reference: string[];
  triggerUploadArtifact = new Subject<string>();
  TriggerArtifactChan$ = this.triggerUploadArtifact.asObservable();
  /**
   * Get all the tags under the specified repository.
   * NOTES: If the Notary is enabled, the signatures should be included in the returned data.
   *
   * @abstract
   *  ** deprecated param {string} repositoryName
   *  ** deprecated param {RequestQueryParams} [queryParams]
   * returns {(Observable<Tag[]>)}
   *
   * @memberOf TagService
   */
  abstract getArtifactList(
    projectName: string,
    repositoryName: string,
    queryParams?: RequestQueryParams
  ): Observable<HttpResponse<Artifact[]>>;

  /**
   * Delete the specified tag.
   *
   * @abstract
   *  ** deprecated param {string} repositoryName
   *  ** deprecated param {string} tag
   * returns {(Observable<any> | any)}
   *
   * @memberOf TagService
   */
  abstract getArtifactFromDigest(
    projectName: string,
    repositoryName: string,
    artifactDigest: string
  ): Observable<Artifact>;

  abstract deleteArtifact(
    projectName: string,
    repositoryName: string,
    digest: string
  ): Observable<any>;

  /**
   * Get the specified tag.
   *
   * @abstract
   *  ** deprecated param {string} repositoryName
   *  ** deprecated param {string} tag
   * returns {(Observable<Tag>)}
   *
   * @memberOf TagService
   */

  abstract addLabelToImages(
    projectName: string,
    repoName: string,
    digest: string,
    labelId: number
  ): Observable<any>;
  abstract deleteLabelToImages(
    projectName: string,
    repoName: string,
    digest: string,
    labelId: number
  ): Observable<any>;

  /**
   * Get manifest of tag under the specified repository.
   *
   * @abstract
   * returns {(Observable<Manifest>)}
   *
   * @memberOf TagService
   */
  abstract getManifest(
    repositoryName: string,
    tag: string
  ): Observable<Manifest>;
}

/**
 * Implement default service for tag.
 *
 **
 * class TagDefaultService
 * extends {TagService}
 */
@Injectable()
export class ArtifactDefaultService extends ArtifactService {
  _baseUrl: string;
  _labelUrl: string;
  reference: string[] = [];
  triggerUploadArtifact = new Subject<string>();
  TriggerArtifactChan$ = this.triggerUploadArtifact.asObservable();

  constructor(
    private http: HttpClient,
    @Inject(SERVICE_CONFIG) private config: IServiceConfig
  ) {
    super();
    this._baseUrl = this.config.repositoryBaseEndpoint
      ? this.config.repositoryBaseEndpoint
      : "/api/repositories";
    this._labelUrl = this.config.labelEndpoint
      ? this.config.labelEndpoint
      : "/api/labels";
  }


  _getArtifacts(
    project_id: string, repositoryName: string,
    queryParams?: RequestQueryParams
  ): Observable<HttpResponse<Artifact[]>> {
    if (!queryParams) {
      queryParams = queryParams = new RequestQueryParams();
    }

    // queryParams = queryParams.set("detail", "true");
    let url: string = `/api/v2.0/projects/${project_id}/repositories/${repositoryName}/artifacts`;
    // /api/v2/projects/{project_id}/repositories/{repositoryName}/artifacts
    return this.http
      .get<HttpResponse<Artifact[]>>(url, buildHttpRequestOptionsWithObserveResponse(queryParams))
      .pipe(map(response => response as HttpResponse<Artifact[]>)
      , catchError(error => observableThrowError(error)));
  }

  public getArtifactList(
    projectName: string,
    repositoryName: string,
    queryParams?: RequestQueryParams
  ): Observable<HttpResponse<Artifact[]>> {
    if (!repositoryName) {
      return observableThrowError("Bad argument");
    }
    return this._getArtifacts(projectName, repositoryName, queryParams);
  }
  public getArtifactFromDigest(
    projectName: string,
    repositoryName: string,
    artifactDigest: string
  ): Observable<Artifact> {
    if (!artifactDigest) {
      return observableThrowError("Bad argument");
    }
    let url = `/api/v2.0/projects/${projectName}/repositories/${repositoryName}/artifacts/${artifactDigest}`;
    return this.http.get(url).pipe(catchError(error => observableThrowError(error))) as Observable<Artifact>;
  }
  public deleteArtifact(
    projectName: string,
    repositoryName: string,
    digest: string
  ): Observable<any> {
    if (!repositoryName || !projectName || !digest) {
      return observableThrowError("Bad argument");
    }

    let url: string = `/api/v2.0/projects/${projectName}/repositories/${repositoryName}/artifacts/${digest}`;
    return this.http
      .delete(url, HTTP_JSON_OPTIONS)
      .pipe(map(response => response)
      , catchError(error => observableThrowError(error)));
  }


  public addLabelToImages(
    projectName: string,
    repoName: string,
    digest: string,
    labelId: number
  ): Observable<any> {
    if (!labelId || !digest || !repoName) {
      return observableThrowError("Invalid parameters.");
    }

    let _addLabelToImageUrl = `
    /api/v2.0/projects/${projectName}/repositories/${repoName}/artifacts/${digest}/labels`;
    return this.http
      .post(_addLabelToImageUrl, { id: labelId }, HTTP_JSON_OPTIONS)
      .pipe(catchError(error => observableThrowError(error)));
  }

  public deleteLabelToImages(
    projectName: string,
    repoName: string,
    digest: string,
    labelId: number
  ): Observable<any> {
    if (!labelId || !digest || !repoName) {
      return observableThrowError("Invalid parameters.");
    }

    let _addLabelToImageUrl = `
    /api/v2.0/projects/${projectName}/repositories/${repoName}/artifacts/${digest}/labels/${labelId}`;
    return this.http
      .delete(_addLabelToImageUrl)
      .pipe(catchError(error => observableThrowError(error)));
  }

  public getManifest(
    repositoryName: string,
    tag: string
  ): Observable<Manifest> {
    if (!repositoryName || !tag) {
      return observableThrowError("Bad argument");
    }
    let url: string = `${this._baseUrl}/${repositoryName}/tags/${tag}/manifest`;
    return this.http
      .get(url, HTTP_GET_OPTIONS)
      .pipe(map(response => response as Manifest)
      , catchError(error => observableThrowError(error)));
  }
}
