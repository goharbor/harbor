import { Observable} from "rxjs";
import { Label } from "./interface";
import { Inject, Injectable } from "@angular/core";
import { Http } from "@angular/http";
import { IServiceConfig, SERVICE_CONFIG } from "../service.config";
import { buildHttpRequestOptions, HTTP_JSON_OPTIONS } from "../utils";
import { RequestQueryParams } from "./RequestQueryParams";

export abstract class LabelService {
  abstract getLabels(
    scope: string,
    projectId?: number,
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]>;

  abstract createLabel(
    label: Label
  ): Observable<Label> | Promise<Label> | Label;

  abstract getLabel(id: number): Observable<Label> | Promise<Label> | Label;

  abstract updateLabel(
    id: number,
    param: Label
  ): Observable<any> | Promise<any> | any;

  abstract deleteLabel(id: number): Observable<any> | Promise<any> | any;
}

@Injectable()
export class LabelDefaultService extends LabelService {
  _labelUrl: string;

  constructor(
    @Inject(SERVICE_CONFIG) config: IServiceConfig,
    private http: Http
  ) {
    super();
    this._labelUrl = config.labelEndpoint
      ? config.labelEndpoint
      : "/api/labels";
  }

  getLabels(
    scope: string,
    projectId?: number,
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]> | Promise<Label[]> {
    if (!queryParams) {
      queryParams = new RequestQueryParams();
    }
    if (scope) {
      queryParams.set("scope", scope);
    }
    if (projectId) {
      queryParams.set("project_id", "" + projectId);
    }
    if (name) {
      queryParams.set("name", "" + name);
    }
    return this.http
      .get(this._labelUrl, buildHttpRequestOptions(queryParams))
      .toPromise()
      .then(response => response.json())
      .catch(error => Promise.reject(error));
  }

  getGLabels(
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]> | Promise<Label[]> {
    if (!queryParams) {
      queryParams = new RequestQueryParams();
    }
    queryParams.set("scope", "g");

    if (name) {
      queryParams.set("name", "" + name);
    }
    return this.http
      .get(this._labelUrl, buildHttpRequestOptions(queryParams))
      .toPromise()
      .then(response => response.json())
      .catch(error => Promise.reject(error));
  }

  getPLabels(
    projectId: number,
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]> | Promise<Label[]> {
    if (!queryParams) {
      queryParams = new RequestQueryParams();
    }
    queryParams.set("scope", "p");
    if (projectId) {
      queryParams.set("project_id", "" + projectId);
    }
    if (name) {
      queryParams.set("name", "" + name);
    }
    return this.http
      .get(this._labelUrl, buildHttpRequestOptions(queryParams))
      .toPromise()
      .then(response => response.json())
      .catch(error => Promise.reject(error));
  }

  createLabel(label: Label): Observable<any> | Promise<any> | any {
    if (!label) {
      return Promise.reject("Invalid label.");
    }
    return this.http
      .post(this._labelUrl, JSON.stringify(label), HTTP_JSON_OPTIONS)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }

  getLabel(id: number): Observable<Label> | Promise<Label> | Label {
    if (!id || id <= 0) {
      return Promise.reject("Bad request argument.");
    }
    let reqUrl = `${this._labelUrl}/${id}`;
    return this.http
      .get(reqUrl)
      .toPromise()
      .then(response => response.json())
      .catch(error => Promise.reject(error));
  }

  updateLabel(id: number, label: Label): Observable<any> | Promise<any> | any {
    if (!id || id <= 0) {
      return Promise.reject("Bad request argument.");
    }
    if (!label) {
      return Promise.reject("Invalid endpoint.");
    }
    let reqUrl = `${this._labelUrl}/${id}`;
    return this.http
      .put(reqUrl, JSON.stringify(label), HTTP_JSON_OPTIONS)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }
  deleteLabel(id: number): Observable<any> | Promise<any> | any {
    if (!id || id <= 0) {
      return Promise.reject("Bad request argument.");
    }
    let reqUrl = `${this._labelUrl}/${id}`;
    return this.http
      .delete(reqUrl)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }
}
