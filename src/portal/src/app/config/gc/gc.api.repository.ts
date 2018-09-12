
import { Injectable } from '@angular/core';
import { Http } from '@angular/http';
import { Observable } from 'rxjs';
import { catchError, map } from 'rxjs/operators';


@Injectable()
export class GcApiRepository {

    constructor(
        private http: Http,
    ) {
    }

    public postSchedule(param): Observable<any> {
        return this.http.post("/api/system/gc/schedule", param)
            .pipe(catchError(err => Observable.throw(err)));
    }

    public putSchedule(param): Observable<any> {
        return this.http.put("/api/system/gc/schedule", param)
            .pipe(catchError(err => Observable.throw(err)));
    }

    public getSchedule(): Observable<any> {
        return this.http.get("/api/system/gc/schedule")
            .pipe(catchError(err => Observable.throw(err)))
            .pipe(map(response => response.json()));
    }

    public getLog(id): Observable<any> {
        return this.http.get("/api/system/gc/" + id + "/log")
            .pipe(catchError(err => Observable.throw(err)));
    }

    public getStatus(id): Observable<any> {
        return this.http.get("/api/system/gc/" + id)
            .pipe(catchError(err => Observable.throw(err)))
            .pipe(map(response => response.json()));
    }

    public getJobs(): Observable<any> {
        return this.http.get("/api/system/gc")
            .pipe(catchError(err => Observable.throw(err)))
            .pipe(map(response => response.json()));
    }

}
