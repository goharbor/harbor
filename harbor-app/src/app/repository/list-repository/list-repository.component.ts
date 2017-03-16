import { Component, Input, Output, EventEmitter } from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';
import { Repository } from '../repository';
import { State } from 'clarity-angular';

import { SearchTriggerService } from '../../base/global-search/search-trigger.service';
import { SessionService } from '../../shared/session.service';
import { ListMode } from '../../shared/shared.const';

@Component({
  selector: 'list-repository',
  templateUrl: 'list-repository.component.html'
})
export class ListRepositoryComponent {

  @Input() projectId: number;
  @Input() repositories: Repository[];
  @Output() delete = new EventEmitter<string>();

  @Input() totalPage: number;
  @Input() totalRecordCount: number;
  @Output() paginate = new EventEmitter<State>();

  @Input() mode: string = ListMode.FULL;

  pageOffset: number = 1;

  constructor(
    private router: Router,
    private searchTrigger: SearchTriggerService,
    private session: SessionService) { }

  deleteRepo(repoName: string) {
    this.delete.emit(repoName);
  }

  refresh(state: State) {
    if (this.repositories) {
      this.paginate.emit(state);
    }
  }

  public get listFullMode(): boolean {
    return this.mode === ListMode.FULL && this.session.getCurrentUser() != null;
  }

  public gotoLink(projectId: number, repoName: string): void {
    this.searchTrigger.closeSearch(false);

    let linkUrl = ['harbor', 'tags', projectId, repoName];
    if (!this.session.getCurrentUser()) {
      let navigatorExtra: NavigationExtras = {
        queryParams: { "guest": true }
      };
      this.router.navigate(linkUrl, navigatorExtra);
    } else {
      this.router.navigate(linkUrl);
    }
  }

}