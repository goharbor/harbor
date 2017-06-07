import { Component, Input, Output, EventEmitter, ChangeDetectorRef, ChangeDetectionStrategy } from '@angular/core';
import { Router } from '@angular/router';

import { State, Comparator } from 'clarity-angular';
import { Repository } from '../service/interface';

import { LIST_REPOSITORY_TEMPLATE } from './list-repository.component.html';

import { CustomComparator } from '../utils';

@Component({
  selector: 'hbr-list-repository',
  template: LIST_REPOSITORY_TEMPLATE,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListRepositoryComponent {
  
  @Input() urlPrefix: string;
  @Input() projectId: number;
  @Input() repositories: Repository[];

  @Output() delete = new EventEmitter<string>();
  @Output() paginate = new EventEmitter<State>();

  @Input() hasProjectAdminRole: boolean;

  pageOffset: number = 1;

  pullCountComparator: Comparator<Repository> = new CustomComparator<Repository>('pull_count', 'number');
  
  tagsCountComparator: Comparator<Repository> = new CustomComparator<Repository>('tags_count', 'number');

  constructor(
    private router: Router,
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

  public gotoLink(projectId: number, repoName: string): void {
    let linkUrl = [this.urlPrefix, 'tags', projectId, repoName];
    this.router.navigate(linkUrl);
  }
}