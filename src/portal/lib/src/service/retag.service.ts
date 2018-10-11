import { Observable } from "rxjs";
import { Http } from "@angular/http";
import { Injectable } from "@angular/core";
import { RetagRequest } from "./interface";
import { HTTP_JSON_OPTIONS } from "../utils";
import { catchError } from "rxjs/operators";
import { throwError as observableThrowError } from "rxjs/index";

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
     * returns {Observable<any>}
     *
     * @memberOf RetagService
     */
    abstract retag(request: RetagRequest): Observable<any>;
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

    retag(request: RetagRequest): Observable<any> {
        return this.http
            .post(`/api/repositories/${request.targetProject}/${request.targetRepo}/tags`,
                {
                    "tag": request.targetTag,
                    "src_image": request.srcImage,
                    "override": request.override
                },
                HTTP_JSON_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }
}
