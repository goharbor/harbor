import { Component, Input, Output, EventEmitter, ChangeDetectorRef, ChangeDetectionStrategy } from '@angular/core';

import { State } from 'clarity-angular';

import { Repository } from '../service/interface';
import { LIST_REPOSITORY_TEMPLATE } from './list-repository.component.html';

@Component({
  selector: 'hbr-list-repository',
  template: LIST_REPOSITORY_TEMPLATE,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListRepositoryComponent {
  @Input() projectId: number;
  @Input() repositories: Repository[];

  @Output() delete = new EventEmitter<string>();
  @Output() paginate = new EventEmitter<State>();

  @Input() hasProjectAdminRole: boolean;

  pageOffset: number = 1;

  constructor(
    private ref: ChangeDetectorRef) { 
    let hnd = setInterval(()=>ref.markForCheck(), 100);
    setTimeout(()=>clearInterval(hnd), 1000);
  }

  ngOnInit() { }

  deleteRepo(repoName: string) {
    this.delete.emit(repoName);
  }

  refresh(state: State) {
    if (this.repositories) {
      this.paginate.emit(state);
    }
  }  
}