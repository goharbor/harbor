import { Component, Input, OnInit } from '@angular/core';

import { StatisticsService } from './statistics.service';
import { errorHandler } from '../../shared/shared.utils';
import { AlertType } from '../../shared/shared.const';

import { MessageService } from '../../global-message/message.service';

import { Statistics } from './statistics';

@Component({
    selector: 'statistics-panel',
    templateUrl: "statistics-panel.component.html",
    styleUrls: ['statistics.component.css'],
    providers: [StatisticsService]
})

export class StatisticsPanelComponent implements OnInit{

    private originalCopy:Statistics = new Statistics();

    constructor(
        private statistics: StatisticsService,
        private msgService: MessageService) { }

    ngOnInit(): void {
        this.getStatistics();
    }

    getStatistics(): void {
        this.statistics.getStatistics()
        .then(statistics => this.originalCopy = statistics)
        .catch(error => {
            this.msgService.announceMessage(error.status, errorHandler(error), AlertType.WARNING);
        })
    }
}