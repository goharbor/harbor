import { Injectable } from '@angular/core';
import { Http } from '@angular/http';
import { Observable, Subscription, Subject, of } from 'rxjs';
import { catchError, map } from 'rxjs/operators';
import { GcApiRepository } from './gc.api.repository';
import { ErrorHandler } from '../../error-handler/index';
import { GcJobData } from './gcLog';


@Injectable()
export class GcRepoService {

    constructor(private http: Http,
        private gcApiRepository: GcApiRepository,
        private errorHandler: ErrorHandler) {
    }

    public manualGc(): Observable <any> {
        let param = {
            "schedule": {
                "type": "Manual"
            }
        };
        return this.gcApiRepository.postSchedule(param);
    }

    public getJobs(): Observable <GcJobData []> {
        return this.gcApiRepository.getJobs();
    }

    public getLog(id): Observable <any> {
        return this.gcApiRepository.getLog(id);
    }

    public getSchedule(): Observable <any> {
        return this.gcApiRepository.getSchedule();
    }

    public postScheduleGc(type, cron): Observable <any> {
        let param = {
            "schedule": {
                "type": type,
                "cron": cron,
            }
        };

        return this.gcApiRepository.postSchedule(param);
    }

    public putScheduleGc(type, cron): Observable <any> {
        let param = {
            "schedule": {
                "type": type,
                "cron": cron,
            }
        };

        return this.gcApiRepository.putSchedule(param);
    }

    public getLogLink(id): string  {
        return this.gcApiRepository.getLogLink(id);
    }
}
