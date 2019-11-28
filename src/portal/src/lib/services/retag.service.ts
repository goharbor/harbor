import { Observable } from "rxjs";
import { HttpClient } from "@angular/common/http";
import { Injectable, Inject } from "@angular/core";
import { RetagRequest } from "./interface";
import { HTTP_JSON_OPTIONS } from "../utils/utils";
import { catchError } from "rxjs/operators";
import { throwError as observableThrowError } from "rxjs/index";
import { IServiceConfig, SERVICE_CONFIG } from "../entities/service.config";

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
        private http: HttpClient,
        @Inject(SERVICE_CONFIG) private config: IServiceConfig
    ) {
        super();
    }

    retag(request: RetagRequest): Observable<any> {
        let baseUrl: string = this.config.repositoryBaseEndpoint ? this.config.repositoryBaseEndpoint : '/api/repositories';
        return this.http
            .post(`${baseUrl}/${request.targetProject}/${request.targetRepo}/tags`,
                {
                    "tag": request.targetTag,
                    "src_image": request.srcImage,
                    "override": request.override
                },
                HTTP_JSON_OPTIONS)
            .pipe(catchError(error => observableThrowError(error)));
    }
}
