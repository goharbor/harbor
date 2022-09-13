import { Component } from '@angular/core';

@Component({
    selector: 'app-clearing-job',
    templateUrl: './clearing-job.component.html',
    styleUrls: ['./clearing-job.component.scss'],
})
export class ClearingJobComponent {
    inProgress: boolean = true;
    constructor() {}
}
