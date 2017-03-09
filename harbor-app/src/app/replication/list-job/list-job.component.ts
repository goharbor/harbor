import { Component, Input, Output, EventEmitter } from '@angular/core';
import { Job } from '../job';
import { State } from 'clarity-angular';

@Component({
  selector: 'list-job',
  templateUrl: 'list-job.component.html'
})
export class ListJobComponent {
  @Input() jobs: Job[];
  @Input() pageSize: number;
  @Output() paginate = new EventEmitter<State>();

  refresh(state: State) {
    if(this.jobs) {
      this.paginate.emit(state);
    }
  }
}