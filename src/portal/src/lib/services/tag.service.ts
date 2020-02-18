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
  abstract newTag(
    projectName: string,
    repositoryName: string,
    digest: string,
    tagName: {name: string}
  ): Observable<any>;

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

  public newTag(
    projectName: string,
    repositoryName: string,
    digest: string,
    tagName: {name: string}
  ): Observable<any> {
    if (!projectName || !repositoryName || !digest || !tagName) {
      return observableThrowError("Bad argument");
    }
    let url: string = `/api/v2.0/projects/${projectName}/repositories/${repositoryName}/artifacts/${digest}/tags`;
    return this.http
      .post(url, tagName, HTTP_JSON_OPTIONS)
      .pipe(map(response => response)
        , catchError(error => observableThrowError(error)));
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
}
