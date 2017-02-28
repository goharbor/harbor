import { Component, Input } from '@angular/core';
import { Job } from '../job';

@Component({
  selector: 'list-job',
  templateUrl: 'list-job.component.html'
})
export class ListJobComponent {
  @Input() jobs: Job[];
}