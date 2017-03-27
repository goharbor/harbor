import { Component, Input, OnInit } from '@angular/core';

import { StatisticsService } from './statistics.service';
import { errorHandler, accessErrorHandler } from '../../shared/shared.utils';
import { AlertType } from '../../shared/shared.const';

import { MessageService } from '../../global-message/message.service';

import { Statistics } from './statistics';

import { SessionService } from '../session.service';
import { Volumes } from './volumes';

@Component({
    selector: 'statistics-panel',
    templateUrl: "statistics-panel.component.html",
    styleUrls: ['statistics.component.css'],
    providers: [StatisticsService]
})

export class StatisticsPanelComponent implements OnInit {

    private originalCopy: Statistics = new Statistics();
    private volumesInfo: Volumes = new Volumes();

    constructor(
        private statistics: StatisticsService,
        private msgService: MessageService,
        private session: SessionService) { }

    ngOnInit(): void {
        if (this.isValidSession) {
            this.getStatistics();
            this.getVolumes();
        }
    }

    public get totalStorage(): number {
        return this.getGBFromBytes(this.volumesInfo.storage.total);
    }

    public get freeStorage(): number {
        return this.getGBFromBytes(this.volumesInfo.storage.free);
    }

    public getStatistics(): void {
        this.statistics.getStatistics()
            .then(statistics => this.originalCopy = statistics)
            .catch(error => {
                if (!accessErrorHandler(error, this.msgService)) {
                    this.msgService.announceMessage(error.status, errorHandler(error), AlertType.WARNING);
                }
            });
    }

    public getVolumes(): void {
        this.statistics.getVolumes()
            .then(volumes => this.volumesInfo = volumes)
            .catch(error => {
                if (!accessErrorHandler(error, this.msgService)) {
                    this.msgService.announceMessage(error.status, errorHandler(error), AlertType.WARNING);
                }
            });
    }

    public get isValidSession(): boolean {
        let user = this.session.getCurrentUser();
        return user && user.has_admin_role > 0;
    }

    private getGBFromBytes(bytes: number): number {
        return Math.round((bytes / (1024 * 1024 * 1024)));
    }
}