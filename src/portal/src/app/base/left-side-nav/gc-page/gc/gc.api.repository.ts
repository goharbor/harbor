import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { throwError as observableThrowError, Observable } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { CURRENT_BASE_HREF } from "../../../../shared/units/utils";

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
        private http: HttpClient
    ) {
        super();
    }

    public postSchedule(param): Observable<any> {
        return this.http.post(`${CURRENT_BASE_HREF}/system/gc/schedule`, param)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public putSchedule(param): Observable<any> {
        return this.http.put(`${CURRENT_BASE_HREF}/system/gc/schedule`, param)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public getSchedule(): Observable<any> {
        return this.http.get(`${CURRENT_BASE_HREF}/system/gc/schedule`)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public getLog(id): Observable<any> {
        return this.http.get(`${CURRENT_BASE_HREF}/system/gc/${id}/log`)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public getStatus(id): Observable<any> {
        return this.http.get(`${CURRENT_BASE_HREF}/system/gc/${id}`)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public getJobs(): Observable<any> {
        return this.http.get(`${CURRENT_BASE_HREF}/system/gc`)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public getLogLink(id) {
        return `${CURRENT_BASE_HREF}/system/gc/${id}/log`;
    }

}
