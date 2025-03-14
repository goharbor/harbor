// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
