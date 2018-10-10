
import { Injectable } from '@angular/core';
import { Http } from '@angular/http';
import { throwError as observableThrowError, Observable } from 'rxjs';
import { catchError, map } from 'rxjs/operators';


@Injectable()
export class GcApiRepository {

    constructor(
        private http: Http,
    ) {
    }

    public postSchedule(param): Observable<any> {
        return this.http.post("/api/system/gc/schedule", param)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public putSchedule(param): Observable<any> {
        return this.http.put("/api/system/gc/schedule", param)
            .pipe(catchError(error => observableThrowError(error)));
    }

    public getSchedule(): Observable<any> {
        return this.http.get("/api/system/gc/schedule")
            .pipe(catchError(error => observableThrowError(error)))
            .pipe(map(response => response.json()));
    }

    public getLog(id): Observable<any> {
        return this.http.get("/api/system/gc/" + id + "/log")
            .pipe(catchError(error => observableThrowError(error)));
    }

    public getStatus(id): Observable<any> {
        return this.http.get("/api/system/gc/" + id)
            .pipe(catchError(error => observableThrowError(error)))
            .pipe(map(response => response.json()));
    }

    public getJobs(): Observable<any> {
        return this.http.get("/api/system/gc")
            .pipe(catchError(error => observableThrowError(error)))
            .pipe(map(response => response.json()));
    }

}
