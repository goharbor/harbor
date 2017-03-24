import { Component, EventEmitter, Output, Input } from '@angular/core';
import { Router } from '@angular/router';

import { SearchTriggerService } from '../../base/global-search/search-trigger.service';

import { Project } from '../../project/project';
import { State } from 'clarity-angular';

@Component({
  moduleId: module.id,
  selector: 'list-project-ro',
  templateUrl: 'list-project-ro.component.html'
})
export class ListProjectROComponent {
  @Input() projects: Project[];

  @Input() totalPage: number;
  @Input() totalRecordCount: number;
  pageOffset: number = 1;

  @Output() paginate = new EventEmitter<State>();

  constructor(
    private searchTrigger: SearchTriggerService,
    private router: Router) { }

  goToLink(proId: number): void {
    this.searchTrigger.closeSearch(true);

    let linkUrl = ['harbor', 'projects', proId, 'repository'];
    this.router.navigate(linkUrl);
  }

  refresh(state: State) {
    this.paginate.emit(state);
  }
}