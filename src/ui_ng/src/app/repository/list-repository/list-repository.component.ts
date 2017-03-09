import { Component, Input, Output, EventEmitter } from '@angular/core';
import { Repository } from '../repository';

@Component({
  selector: 'list-repository',
  templateUrl: 'list-repository.component.html'
})
export class ListRepositoryComponent {
  
  @Input() projectId: number;
  @Input() repositories: Repository[];
  @Output() delete = new EventEmitter<string>();

  deleteRepo(repoName: string) {
    this.delete.emit(repoName);
  } 
}