import { Inject, Injectable } from "@angular/core";
import { Http } from "@angular/http";
import { Observable} from "rxjs";
import { map } from "rxjs/operators";

import { RequestQueryParams } from "./RequestQueryParams";
import { Label } from "./interface";

import { IServiceConfig, SERVICE_CONFIG } from "../service.config";
import { buildHttpRequestOptions, HTTP_JSON_OPTIONS } from "../utils";
import { extractJson } from "../shared/shared.utils";

export abstract class LabelService {
  abstract getGLabels(
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]> | Promise<Label[]>;

  abstract getPLabels(
    projectId: number,
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]> | Promise<Label[]>;

  abstract getProjectLabels(
    projectId: number,
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]>;

  abstract getLabels(
    scope: string,
    projectId?: number,
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]> | Promise<Label[]>;

  abstract createLabel(
    label: Label
  ): Observable<Label> | Promise<Label> | Label;

  abstract getLabel(id: number): Observable<Label> | Promise<Label> | Label;

  abstract updateLabel(
    id: number,
    param: Label
  ): Observable<any> | Promise<any> | any;

  abstract deleteLabel(id: number): Observable<any> | Promise<any> | any;

  abstract getChartVersionLabels(
    projectName: string,
    chartName: string,
    version?: string,
  ): Observable<Label[]>;

  abstract markChartLabel(
    projectName: string,
    chartName: string,
    version: string,
    label: Label,
  ): Observable<any>;

  abstract unmarkChartLabel(
    projectName: string,
    chartName: string,
    version: string,
    label: Label,
  ): Observable<any>;
}

@Injectable()
export class LabelDefaultService extends LabelService {
  labelUrl: string;
  chartUrl: string;

  constructor(
    @Inject(SERVICE_CONFIG) config: IServiceConfig,
    private http: Http
  ) {
    super();
    this.labelUrl = config.labelEndpoint ? config.labelEndpoint : "/api/labels";
    this.chartUrl =  config.helmChartEndpoint ? config.helmChartEndpoint : "/api/chartrepo";
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
      .get(this.labelUrl, buildHttpRequestOptions(queryParams))
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
      .get(this.labelUrl, buildHttpRequestOptions(queryParams))
      .toPromise()
      .then(response => response.json())
      .catch(error => Promise.reject(error));
  }

  getProjectLabels(
    projectId: number,
    name?: string,
    queryParams?: RequestQueryParams
  ): Observable<Label[]> {
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
    return this.http.get(this.labelUrl, buildHttpRequestOptions(queryParams))
    .pipe(map( res => extractJson(res)));
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
      .get(this.labelUrl, buildHttpRequestOptions(queryParams))
      .toPromise()
      .then(response => response.json())
      .catch(error => Promise.reject(error));
  }

  createLabel(label: Label): Observable<any> | Promise<any> | any {
    if (!label) {
      return Promise.reject("Invalid label.");
    }
    return this.http
      .post(this.labelUrl, JSON.stringify(label), HTTP_JSON_OPTIONS)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }

  getLabel(id: number): Observable<Label> | Promise<Label> | Label {
    if (!id || id <= 0) {
      return Promise.reject("Bad request argument.");
    }
    let reqUrl = `${this.labelUrl}/${id}`;
    return this.http
      .get(reqUrl, HTTP_JSON_OPTIONS)
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
    let reqUrl = `${this.labelUrl}/${id}`;
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
    let reqUrl = `${this.labelUrl}/${id}`;
    return this.http
      .delete(reqUrl)
      .toPromise()
      .then(response => response.status)
      .catch(error => Promise.reject(error));
  }

  getChartVersionLabels(
    projectName: string,
    chartName: string,
    version: string
  ): Observable<Label[]> {
    return this.http.get(`${this.chartUrl}/${projectName}/charts/${chartName}/${version}/labels`)
    .pipe(map(res => extractJson(res)));
  }

  markChartLabel(
    projectName: string,
    chartName: string,
    version: string,
    label: Label,
  ): Observable<any> {
    return this.http.post(`${this.chartUrl}/${projectName}/charts/${chartName}/${version}/labels`,
    JSON.stringify(label), HTTP_JSON_OPTIONS)
    .pipe(map(res => extractJson(res)));
  }

  unmarkChartLabel(
    projectName: string,
    chartName: string,
    version: string,
    label: Label,
  ): Observable<any> {
    return this.http.delete(`${this.chartUrl}/${projectName}/charts/${chartName}/${version}/labels/${label.id}`, HTTP_JSON_OPTIONS)
    .pipe(map(res => extractJson(res)));
  }

}
