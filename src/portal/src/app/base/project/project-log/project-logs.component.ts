import { Component } from '@angular/core';

@Component({
    selector: 'project-logs',
    templateUrl: './project-logs.component.html',
    styleUrls: ['./project-logs.component.scss'],
})
export class ProjectLogsComponent {
    inProgress: boolean = true;
    constructor() {}
}
