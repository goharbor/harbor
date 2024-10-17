import { Injectable } from '@angular/core';
import { JobserviceService } from '../../../../../ng-swagger-gen/services/jobservice.service';

export const HEALTHY_TIME: number = 24; //  unit hours
export const CHECK_HEALTH_INTERVAL: number = 15 * 60 * 1000; //15 minutes, unit ms

@Injectable({
    providedIn: 'root',
})
export class JobServiceDashboardHealthCheckService {
    private _hasUnhealthyQueue: boolean = false;
    private _hasManuallyClosed: boolean = false;

    constructor(private jobServiceService: JobserviceService) {}

    hasUnhealthyQueue(): boolean {
        return this._hasUnhealthyQueue;
    }

    hasManuallyClosed(): boolean {
        return this._hasManuallyClosed;
    }

    setUnHealthy(value: boolean): void {
        this._hasUnhealthyQueue = value;
    }
    setManuallyClosed(value: boolean): void {
        this._hasManuallyClosed = value;
    }

    checkHealth(): void {
        this.jobServiceService.listJobQueues().subscribe(res => {
            this._hasUnhealthyQueue = res?.some(
                item => item.latency >= HEALTHY_TIME * 60 * 60
            );
        });
    }
}
