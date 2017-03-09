import { Component, Input, Output, EventEmitter } from '@angular/core';
import { Repository } from '../repository';
import { State } from 'clarity-angular';

@Component({
  selector: 'list-repository',
  templateUrl: 'list-repository.component.html'
})
export class ListRepositoryComponent {
  
  @Input() projectId: number;
  @Input() repositories: Repository[];
  @Output() delete = new EventEmitter<string>();

  @Input() total: number;
  @Input() pageSize: number;
  @Output() paginate = new EventEmitter<State>();

  deleteRepo(repoName: string) {
    this.delete.emit(repoName);
  } 

  refresh(state: State) {
    if(this.repositories) {
      this.paginate.emit(state);
    }
  }

}