import { Component } from '@angular/core';

@Component({
    selector: 'app-logs',
    templateUrl: './logs.component.html',
    styleUrls: ['./logs.component.scss'],
})
export class LogsComponent {
    inProgress: boolean = true;
    constructor() {}
}
