import { Component, Input, Output, EventEmitter } from '@angular/core';
import { Job } from '../job';
import { State } from 'clarity-angular';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';

@Component({
  selector: 'list-job',
  templateUrl: 'list-job.component.html'
})
export class ListJobComponent {
  @Input() jobs: Job[];
  @Input() totalRecordCount: number;
  @Input() totalPage: number;
  @Output() paginate = new EventEmitter<State>();

  constructor(private messageHandlerService: MessageHandlerService) {}

  pageOffset: number = 1;

  refresh(state: State) {
    if(this.jobs) {
      for(let i = 0; i < this.jobs.length; i++) {
        let j = this.jobs[i];
        if(j.status === 'retrying' || j.status === 'error') {
          this.messageHandlerService.showError('REPLICATION.FOUND_ERROR_IN_JOBS', '');
        }
      }
      this.paginate.emit(state);
    }
  }
}