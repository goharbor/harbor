import { Component, Input, Output, EventEmitter } from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';
import { Repository } from '../../repository/repository';
import { State } from 'clarity-angular';

import { SearchTriggerService } from '../../base/global-search/search-trigger.service';

@Component({
  selector: 'list-repository-ro',
  templateUrl: 'list-repository-ro.component.html'
})
export class ListRepositoryROComponent {

  @Input() projectId: number;
  @Input() repositories: Repository[];

  @Input() totalPage: number;
  @Input() totalRecordCount: number;
  @Output() paginate = new EventEmitter<State>();
  pageOffset: number = 1;

  constructor(
    private router: Router,
    private searchTrigger: SearchTriggerService
    ) { }

  refresh(state: State) {
    if (this.repositories) {
      this.paginate.emit(state);
    }
  }

  public gotoLink(projectId: number, repoName: string): void {
    this.searchTrigger.closeSearch(true);

    let linkUrl = ['harbor', 'tags', projectId, repoName];
    this.router.navigate(linkUrl);
  }

}