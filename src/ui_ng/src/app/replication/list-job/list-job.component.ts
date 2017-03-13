import { Component, Input, Output, EventEmitter } from '@angular/core';
import { Job } from '../job';
import { State } from 'clarity-angular';

@Component({
  selector: 'list-job',
  templateUrl: 'list-job.component.html'
})
export class ListJobComponent {
  @Input() jobs: Job[];
  @Input() totalRecordCount: number;
  @Input() totalPage: number;
  @Output() paginate = new EventEmitter<State>();

  pageOffset: number = 1;

  refresh(state: State) {
    if(this.jobs) {
      this.paginate.emit(state);
    }
  }
}