import { Observable } from "rxjs";
import { Http } from "@angular/http";
import { Injectable } from '@angular/core';
import { RetagRequest } from "./interface";
import { HTTP_JSON_OPTIONS } from "../utils";

/**
 * Define the service methods to perform images retag.
 *
 **
 * @abstract
 * class RetagService
 */
export abstract class RetagService {
    /**
     * Retag an image.
     *
     * @abstract
     * param {RetagRequest} request
     * returns {(Observable<any> | Promise<any> | any)}
     *
     * @memberOf RetagService
     */
    abstract retag(request: RetagRequest): Observable<any> | Promise<any> | any;
}

/**
 * Implement default service for retag.
 *
 **
 * class RetagDefaultService
 * extends {RetagService}
 */
@Injectable()
export class RetagDefaultService extends RetagService {
    constructor(
        private http: Http
    ) {
        super();
    }

    retag(request: RetagRequest): Observable<any> | Promise<any> | any {
        return this.http
            .post(`/api/repositories/${request.targetProject}/${request.targetRepo}/tags`,
                {
                    "tag": request.targetTag,
                    "src_image": request.srcImage,
                    "override": request.override
                },
                HTTP_JSON_OPTIONS)
            .toPromise()
            .then(response => response.status)
            .catch(error => Promise.reject(error));
    };
}