import { Component, EventEmitter, Output, Input, OnInit } from '@angular/core';
import { Router, NavigationExtras } from '@angular/router';
import { Project } from '../project';
import { ProjectService } from '../project.service';

import { SessionService } from '../../shared/session.service';
import { SessionUser } from '../../shared/session-user';
import { SearchTriggerService } from '../../base/global-search/search-trigger.service';
import { signInRoute } from '../../shared/shared.const';

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
    this.searchTrigger.closeSearch(false);
    
    let linkUrl = ['harbor', 'projects', proId, 'repository'];
    if (!this.session.getCurrentUser()) {
      let navigatorExtra: NavigationExtras = {
        queryParams: { "redirect_url": linkUrl.join("/") }
      };

      this.router.navigate([signInRoute], navigatorExtra);
    } else {
      this.router.navigate(linkUrl);

    }
  }

  toggleProject(p: Project) {
    this.toggle.emit(p);
  }

  deleteProject(p: Project) {
    this.delete.emit(p);
  }

}