
import { Injectable, Inject } from '@angular/core';
import { Http } from '@angular/http';
import { throwError as observableThrowError, Observable } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { SERVICE_CONFIG, IServiceConfig } from "../../service.config";

export abstract class GcApiRepository {
    abstract postSchedule(param): Observable<any>;

    abstract putSchedule(param): Observable<any>;

    abstract getSchedule(): Observable<any>;

    abstract getLog(id): Observable<any>;

    abstract getStatus(id): Observable<any>;

    abstract getJobs(): Observable<any>;

    abstract getLogLink(id): string;
}

@Injectable()
export class GcApiDefaultRepository extends GcApiRepository {
    constructor(
        private http: Http,
        @Inject(SERVICE_CONFIG) private config: IServiceConfig
    ) {
        super();
    }

    public postSchedule(param): Observable<any> {
        return this.http.post(`${this.config.gcEndpoint}/schedule`, param)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public putSchedule(param): Observable<any> {
        return this.http.put(`${this.config.gcEndpoint}/schedule`, param)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public getSchedule(): Observable<any> {
        return this.http.get(`${this.config.gcEndpoint}/schedule`)
            .pipe(catchError(error => observableThrowError(error)))
            .pipe(map(response => response.json()));
    }

    public getLog(id): Observable<any> {
        return this.http.get(`${this.config.gcEndpoint}/${id}/log`)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public getStatus(id): Observable<any> {
        return this.http.get(`${this.config.gcEndpoint}/id`)
            .pipe(catchError(error => observableThrowError(error)))
            .pipe(map(response => response.json()));
    }

    public getJobs(): Observable<any> {
        return this.http.get(`${this.config.gcEndpoint}`)
            .pipe(catchError(error => observableThrowError(error)))
            .pipe(map(response => response.json()));
    }

    public getLogLink(id) {
        return `${this.config.gcEndpoint}/${id}/log`;
    }

}
