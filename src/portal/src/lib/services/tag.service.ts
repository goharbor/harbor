import { Injectable, Inject } from "@angular/core";
import { HttpClient } from "@angular/common/http";

import { SERVICE_CONFIG, IServiceConfig } from "../entities/service.config";
import {
  buildHttpRequestOptions,
  HTTP_JSON_OPTIONS,
  HTTP_GET_OPTIONS
} from "../utils/utils";
import { RequestQueryParams } from "./RequestQueryParams";
import { Tag, Manifest } from "./interface";
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";

/**
 * For getting tag signatures.
 * This is temporary, will be removed in future.
 *
 **
 * class VerifiedSignature
 */
export class VerifiedSignature {
  tag: string;
  hashes: {
    sha256: string;
  };
}

/**
 * Define the service methods to handle the repository tag related things.
 *
 **
 * @abstract
 * class TagService
 */
export abstract class TagService {
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
  // to delete
  abstract getTags(
    repositoryName: string,
    queryParams?: RequestQueryParams
  ): Observable<Tag[]>;

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
  abstract deleteTag(
    projectName: string,
    repositoryName: string,
    digest: string,
    tagName: string
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
  abstract getTag(
    repositoryName: string,
    tag: string,
    queryParams?: RequestQueryParams
  ): Observable<Tag>;

  abstract addLabelToImages(
    repoName: string,
    tagName: string,
    labelId: number
  ): Observable<any>;
  abstract deleteLabelToImages(
    repoName: string,
    tagName: string,
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
export class TagDefaultService extends TagService {
  _baseUrl: string;
  _labelUrl: string;
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

  // Private methods
  // These two methods are temporary, will be deleted in future after API refactored
  _getTags(
    repositoryName: string,
    queryParams?: RequestQueryParams
  ): Observable<Tag[]> {
    if (!queryParams) {
      queryParams = queryParams = new RequestQueryParams();
    }

    queryParams = queryParams.set("detail", "true");
    let url: string = `${this._baseUrl}/${repositoryName}/tags`;

    return this.http
      .get(url, buildHttpRequestOptions(queryParams))
      .pipe(map(response => response as Tag[])
        , catchError(error => observableThrowError(error)));
  }

  _getSignatures(repositoryName: string): Observable<VerifiedSignature[]> {
    let url: string = `${this._baseUrl}/${repositoryName}/signatures`;
    return this.http
      .get(url, HTTP_GET_OPTIONS)
      .pipe(map(response => response as VerifiedSignature[])
        , catchError(error => observableThrowError(error)));
  }

  public getTags(
    repositoryName: string,
    queryParams?: RequestQueryParams
  ): Observable<Tag[]> {
    if (!repositoryName) {
      return observableThrowError("Bad argument");
    }
    return this._getTags(repositoryName, queryParams);
  }

  public deleteTag(
    projectName: string,
    repositoryName: string,
    digest: string,
    tagName: string
  ): Observable<any> {
    if (!projectName || !repositoryName || !digest || !tagName) {
      return observableThrowError("Bad argument");
    }

    let url: string = `/api/v2.0/projects/${projectName}/repositories/${repositoryName}/artifacts/${digest}/tags/${tagName}`;
    return this.http
      .delete(url, HTTP_JSON_OPTIONS)
      .pipe(map(response => response)
        , catchError(error => observableThrowError(error)));
  }

  public getTag(
    repositoryName: string,
    tag: string,
    queryParams?: RequestQueryParams
  ): Observable<Tag> {
    if (!repositoryName || !tag) {
      return observableThrowError("Bad argument");
    }

    let url: string = `${this._baseUrl}/${repositoryName}/tags/${tag}`;
    return this.http
      .get(url, HTTP_GET_OPTIONS)
      .pipe(map(response => response as Tag)
        , catchError(error => observableThrowError(error)));
  }

  public addLabelToImages(
    repoName: string,
    tagName: string,
    labelId: number
  ): Observable<any> {
    if (!labelId || !tagName || !repoName) {
      return observableThrowError("Invalid parameters.");
    }

    let _addLabelToImageUrl = `${
      this._baseUrl
      }/${repoName}/tags/${tagName}/labels`;
    return this.http
      .post(_addLabelToImageUrl, { id: labelId }, HTTP_JSON_OPTIONS)
      .pipe(catchError(error => observableThrowError(error)));
  }

  public deleteLabelToImages(
    repoName: string,
    tagName: string,
    labelId: number
  ): Observable<any> {
    if (!labelId || !tagName || !repoName) {
      return observableThrowError("Invalid parameters.");
    }

    let _addLabelToImageUrl = `${
      this._baseUrl
      }/${repoName}/tags/${tagName}/labels/${labelId}`;
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
