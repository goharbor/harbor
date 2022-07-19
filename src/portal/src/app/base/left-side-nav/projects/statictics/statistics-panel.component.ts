// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { Component, OnInit, OnDestroy } from '@angular/core';
import { Subscription } from 'rxjs';
import { SessionService } from '../../../../shared/services/session.service';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { StatisticHandler } from './statistic-handler.service';
import { Statistic } from '../../../../../../ng-swagger-gen/models/statistic';
import { StatisticService } from '../../../../../../ng-swagger-gen/services/statistic.service';
import { getSizeNumber, getSizeUnit } from '../../../../shared/units/utils';

@Component({
    selector: 'statistics-panel',
    templateUrl: 'statistics-panel.component.html',
    styleUrls: ['statistics-panel.component.scss'],
})
export class StatisticsPanelComponent implements OnInit, OnDestroy {
    originalCopy: Statistic;
    refreshSub: Subscription;
    constructor(
        private statistics: StatisticService,
        private msgHandler: MessageHandlerService,
        private session: SessionService,
        private statisticHandler: StatisticHandler
    ) {}

    ngOnInit(): void {
        // Refresh
        this.refreshSub = this.statisticHandler.refreshChan$.subscribe(
            clear => {
                this.getStatistics();
            }
        );

        if (this.session.getCurrentUser()) {
            this.getStatistics();
        }
    }

    ngOnDestroy() {
        if (this.refreshSub) {
            this.refreshSub.unsubscribe();
        }
    }
    getStatistics(): void {
        this.statistics.getStatistic().subscribe(
            statistics => (this.originalCopy = statistics),
            error => {
                this.msgHandler.handleError(error);
            }
        );
    }
    get isValidSession(): boolean {
        let user = this.session.getCurrentUser();
        return user && user.has_admin_role;
    }
    getSizeNumber(): number | string {
        if (this.originalCopy) {
            return getSizeNumber(this.originalCopy.total_storage_consumption);
        }
        return 0;
    }
    getSizeUnit(): number | string {
        if (this.originalCopy) {
            return getSizeUnit(this.originalCopy.total_storage_consumption);
        }
        return null;
    }
}
