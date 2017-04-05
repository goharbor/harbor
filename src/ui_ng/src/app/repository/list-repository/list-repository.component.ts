import { Component, Input, Output, EventEmitter, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { Repository } from '../repository';
import { State } from 'clarity-angular';

import { SearchTriggerService } from '../../base/global-search/search-trigger.service';

@Component({
  selector: 'list-repository',
  templateUrl: 'list-repository.component.html'
})
export class ListRepositoryComponent implements OnInit {

  @Input() projectId: number;
  @Input() repositories: Repository[];


  @Output() delete = new EventEmitter<string>();

  @Input() totalPage: number;
  @Input() totalRecordCount: number;
  @Output() paginate = new EventEmitter<State>();

  @Input() hasProjectAdminRole: boolean;

  pageOffset: number = 1;

  constructor(
    private router: Router,
    private searchTrigger: SearchTriggerService) { }

  ngOnInit() { }

  deleteRepo(repoName: string) {
    this.delete.emit(repoName);
  }

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