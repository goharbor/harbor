import { Injectable, Inject } from "@angular/core";
import { Http } from "@angular/http";
import { Observable } from "rxjs/Observable";
import "rxjs/add/observable/of";

import { SERVICE_CONFIG, IServiceConfig } from "../service.config";
import {
  buildHttpRequestOptions,
  HTTP_JSON_OPTIONS,
  HTTP_GET_OPTIONS
} from "../utils";
import { RequestQueryParams } from "./RequestQueryParams";
import { Tag } from "./interface";

/**
 * For getting tag signatures.
 * This is temporary, will be removed in future.
 *
 * @export
 * @class VerifiedSignature
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
 * @export
 * @abstract
 * @class TagService
 */
export abstract class TagService {
  /**
   * Get all the tags under the specified repository.
   * NOTES: If the Notary is enabled, the signatures should be included in the returned data.
   *
   * @abstract
   * @param {string} repositoryName
   * @param {RequestQueryParams} [queryParams]
   * @returns {(Observable<Tag[]> | Promise<Tag[]> | Tag[])}
   *
   * @memberOf TagService
   */
  abstract getTags(
    repositoryName: string,
    queryParams?: RequestQueryParams
  ): Observable<Tag[]> | Promise<Tag[]> | Tag[];

  /**
   * Delete the specified tag.
   *
   * @abstract
   * @param {string} repositoryName
   * @param {string} tag
   * @returns {(Observable<any> | any)}
   *
   * @memberOf TagService
   */
  abstract deleteTag(
    repositoryName: string,
    tag: string
  ): Observable<any> | Promise<any> | any;

  /**
   * Get the specified tag.
   *
   * @abstract
   * @param {string} repositoryName
   * @param {string} tag
   * @returns {(Observable<Tag> | Promise<Tag> | Tag)}
   *
   * @memberOf TagService
   */
  abstract getTag(
    repositoryName: string,
    tag: string,
    queryParams?: RequestQueryParams
  ): Observable<Tag> | Promise<Tag> | Tag;

  abstract addLabelToImages(
    repoName: string,
    tagName: string,
    labelId: number
  ): Observable<any> | Promise<any> | any;
  abstract deleteLabelToImages(
    repoName: string,
    tagName: string,
    labelId: number
  ): Observable<any> | Promise<any> | any;
}

/**
 * Implement default service for tag.
 *
 * @export
 * @class TagDefaultService
 * @extends {TagService}
 */
@Injectable()
export class TagDefaultService extends TagService {
  _baseUrl: string;
  _labelUrl: string;
  constructor(
    private http: Http,
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
  ): Promise<Tag[]> {
    if (!queryParams) {
      queryParams = new RequestQueryParams();
    }

    queryParams.set("detail", "1");
    let url: string = `${this._baseUrl}/${repositoryName}/tags`;

    return this.http
      .get(url, buildHttpRequestOptions(queryParams))
      .toPromise()
      .then(response => response.json() as Tag[])
      .catch(error => Promise.reject(error));
  }

  _getSignatures(repositoryName: string): Promise<VerifiedSignature[]> {
    let url: string = `${this._baseUrl}/${repositoryName}/signatures`;
    return this.http
      .get(url, HTTP_GET_OPTIONS)
      .toPromise()
      .then(response => response.json() as VerifiedSignature[])
      .catch(error => Promise.reject(error));
  }

  public getTags(
    repositoryName: string,
    queryParams?: RequestQueryParams
  ): Observable<Tag[]> | Promise<Tag[]> | Tag[] {
    if (!repositoryName) {
      return Promise.reject("Bad argument");
    }
    return this._getTags(repositoryName, queryParams);
  }

  public deleteTag(
    repositoryName: string,
    tag: string
  ): Observable<any> | Promise<Tag> | any {
    if (!repositoryName || !tag) {
      return Promise.reject("Bad argument");
    }

    let url: string = `${this._baseUrl}/${repositoryName}/tags/${tag}`;
    return this.http
      .delete(url, HTTP_JSON_OPTIONS)
      .toPromise()
      .then(response => response)
      .catch(error => Promise.reject(error));
  }

  public getTag(
    repositoryName: string,
    tag: string,
    queryParams?: RequestQueryParams
  ): Observable<Tag> | Promise<Tag> | Tag {
    if (!repositoryName || !tag) {
      return Promise.reject("Bad argument");
    }

    let url: string = `${this._baseUrl}/${repositoryName}/tags/${tag}`;
    return this.http
      .get(url, HTTP_GET_OPTIONS)
      .toPromise()
      .then(response => response.json() as Tag)
      .catch(error => Promise.reject(error));
  }

  public addLabelToImages(
    repoName: string,
    tagName: string,
    labelId: number
  ): Observable<any> | Promise<any> | any {
    if (!labelId || !tagName || !repoName) {
      return Promise.reject("Invalid parameters.");
    }

    let _addLabelToImageUrl = `${
      this._baseUrl
    }/${repoName}/tags/${tagName}/labels`;
    return this.http
      .post(_addLabelToImageUrl, { id: labelId }, HTTP_JSON_OPTIONS)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }

  public deleteLabelToImages(
    repoName: string,
    tagName: string,
    labelId: number
  ): Observable<any> | Promise<any> | any {
    if (!labelId || !tagName || !repoName) {
      return Promise.reject("Invalid parameters.");
    }

    let _addLabelToImageUrl = `${
      this._baseUrl
    }/${repoName}/tags/${tagName}/labels/${labelId}`;
    return this.http
      .delete(_addLabelToImageUrl)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }
}
