import { Component, EventEmitter, Output, Input, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { Project } from '../project';
import { ProjectService } from '../project.service';

import { SessionService } from '../../shared/session.service';
import { SessionUser } from '../../shared/session-user';
import { SearchTriggerService } from '../../base/global-search/search-trigger.service';


@Component({
  selector: 'list-project',
  templateUrl: 'list-project.component.html'
})
export class ListProjectComponent implements OnInit {

  @Input() projects: Project[];

  @Output() toggle = new EventEmitter<Project>();
  @Output() delete = new EventEmitter<Project>();

  private currentUser: SessionUser = null;

  constructor(
    private session: SessionService,
    private router: Router,
    private searchTrigger: SearchTriggerService) { }

  ngOnInit(): void {
    this.currentUser = this.session.getCurrentUser();
  }

  public get isSessionValid(): boolean {
    return this.currentUser != null;
  }

  goToLink(proId: number): void {
    this.router.navigate(['/harbor', 'projects', proId, 'repository']);
    this.searchTrigger.closeSearch(false);
  }

  toggleProject(p: Project) {
    this.toggle.emit(p);
  }

  deleteProject(p: Project) {
    this.delete.emit(p);
  }

}