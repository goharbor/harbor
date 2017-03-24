import { Component, Input } from '@angular/core';

@Component({
    selector: 'statistics',
    templateUrl: "statistics.component.html",
    styleUrls: ['statistics.component.css']
})

export class StatisticsComponent {
    @Input() label: string;
    @Input() data: number = 0;
}