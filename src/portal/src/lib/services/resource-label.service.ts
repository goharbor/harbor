import { Label } from "./interface";
import { Inject, Injectable } from "@angular/core";
import { HttpClient } from "@angular/common/http";
import { IServiceConfig, SERVICE_CONFIG } from "../entities/service.config";
import { buildHttpRequestOptions, CURRENT_BASE_HREF, HTTP_JSON_OPTIONS } from "../utils/utils";
import { RequestQueryParams } from "./RequestQueryParams";
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";

export abstract class LabelService {
  abstract getLabels(
    scope: string,
    projectId?: number,
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]>;

  abstract createLabel(
    label: Label
  ): Observable<Label>;

  abstract getLabel(id: number): Observable<Label>;

  abstract updateLabel(
    id: number,
    param: Label
  ): Observable<any>;

  abstract deleteLabel(id: number): Observable<any>;
}

@Injectable()
export class LabelDefaultService extends LabelService {
  _labelUrl: string;

  constructor(
    @Inject(SERVICE_CONFIG) config: IServiceConfig,
    private http: HttpClient
  ) {
    super();
    this._labelUrl = config.labelEndpoint
      ? config.labelEndpoint
      : CURRENT_BASE_HREF + "/labels";
  }

  getLabels(
    scope: string,
    projectId?: number,
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]> {
    if (!queryParams) {
      queryParams = new RequestQueryParams();
    }
    if (scope) {
      queryParams = queryParams.set("scope", scope);
    }
    if (projectId) {
      queryParams = queryParams.set("project_id", "" + projectId);
    }
    if (name) {
      queryParams = queryParams.set("name", "" + name);
    }
    return this.http
      .get<Label[]>(this._labelUrl, buildHttpRequestOptions(queryParams))
      .pipe(catchError(error => observableThrowError(error)));
  }

  getGLabels(
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]> {
    if (!queryParams) {
      queryParams = new RequestQueryParams();
    }
    queryParams = queryParams.set("scope", "g");

    if (name) {
      queryParams = queryParams.set("name", "" + name);
    }
    return this.http
      .get<Label[]>(this._labelUrl, buildHttpRequestOptions(queryParams))
      .pipe(catchError(error => observableThrowError(error)));
  }

  getPLabels(
    projectId: number,
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]> {
    if (!queryParams) {
      queryParams = new RequestQueryParams();
    }
    queryParams = queryParams.set("scope", "p");
    if (projectId) {
      queryParams = queryParams.set("project_id", "" + projectId);
    }
    if (name) {
      queryParams = queryParams.set("name", "" + name);
    }
    return this.http
      .get<Label[]>(this._labelUrl, buildHttpRequestOptions(queryParams))
      .pipe(catchError(error => observableThrowError(error)));
  }

  createLabel(label: Label): Observable<any> {
    if (!label) {
      return observableThrowError("Invalid label.");
    }
    return this.http
      .post(this._labelUrl, JSON.stringify(label), HTTP_JSON_OPTIONS)
      .pipe(catchError(error => observableThrowError(error)));
  }

  getLabel(id: number): Observable<Label> {
    if (!id || id <= 0) {
      return observableThrowError("Bad request argument.");
    }
    let reqUrl = `${this._labelUrl}/${id}`;
    return this.http
      .get<any>(reqUrl)
      .pipe(catchError(error => observableThrowError(error)));
  }

  updateLabel(id: number, label: Label): Observable<any> {
    if (!id || id <= 0) {
      return observableThrowError("Bad request argument.");
    }
    if (!label) {
      return observableThrowError("Invalid endpoint.");
    }
    let reqUrl = `${this._labelUrl}/${id}`;
    return this.http
      .put(reqUrl, JSON.stringify(label), HTTP_JSON_OPTIONS)
      .pipe(catchError(error => observableThrowError(error)));
  }
  deleteLabel(id: number): Observable<any> {
    if (!id || id <= 0) {
      return observableThrowError("Bad request argument.");
    }
    let reqUrl = `${this._labelUrl}/${id}`;
    return this.http
      .delete(reqUrl)
      .pipe(catchError(error => observableThrowError(error)));
  }
}
