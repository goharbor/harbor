import { Component, Input, OnInit } from '@angular/core';

import { StatisticsService } from './statistics.service';
import { errorHandler } from '../../shared/shared.utils';
import { AlertType } from '../../shared/shared.const';

import { MessageService } from '../../global-message/message.service';

import { Statistics } from './statistics';

import { SessionService } from '../session.service';

@Component({
    selector: 'statistics-panel',
    templateUrl: "statistics-panel.component.html",
    styleUrls: ['statistics.component.css'],
    providers: [StatisticsService]
})

export class StatisticsPanelComponent implements OnInit {

    private originalCopy: Statistics = new Statistics();

    constructor(
        private statistics: StatisticsService,
        private msgService: MessageService,
        private session: SessionService) { }

    ngOnInit(): void {
        if (this.session.getCurrentUser()) {
            this.getStatistics();
        }
    }

    getStatistics(): void {
        this.statistics.getStatistics()
            .then(statistics => this.originalCopy = statistics)
            .catch(error => {
                this.msgService.announceMessage(error.status, errorHandler(error), AlertType.WARNING);
            })
    }

    public get isValidSession(): boolean {
        let user = this.session.getCurrentUser();
        return user && user.has_admin_role > 0;
    }
}