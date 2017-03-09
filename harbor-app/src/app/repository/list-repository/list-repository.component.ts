import { Component, Input, Output, EventEmitter } from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';
import { Repository } from '../repository';

import { SearchTriggerService } from '../../base/global-search/search-trigger.service';
import { SessionService } from '../../shared/session.service';
import { signInRoute } from '../../shared/shared.const';

@Component({
  selector: 'list-repository',
  templateUrl: 'list-repository.component.html'
})
export class ListRepositoryComponent {

  @Input() projectId: number;
  @Input() repositories: Repository[];
  @Output() delete = new EventEmitter<string>();

  constructor(
    private router: Router,
    private searchTrigger: SearchTriggerService,
    private session: SessionService) { }

  public gotoLink(projectId: number, repoName: string): void {
    this.searchTrigger.closeSearch(false);

    let linkUrl = ['harbor', 'tags', projectId, repoName];
    if (!this.session.getCurrentUser()) {
      let navigatorExtra: NavigationExtras = {
        queryParams: { "redirect_url": linkUrl.join("/") }
      };

      this.router.navigate([signInRoute], navigatorExtra);
    } else {
      this.router.navigate(linkUrl);
    }
  }

  deleteRepo(repoName: string) {
    this.delete.emit(repoName);
  }
}