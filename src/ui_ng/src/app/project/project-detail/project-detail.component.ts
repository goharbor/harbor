import { Component } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

import { Project } from '../project';

import { SessionService } from '../../shared/session.service';

@Component({
    selector: 'project-detail',
    templateUrl: "project-detail.component.html",
    styleUrls: [ 'project-detail.component.css' ]
})
export class ProjectDetailComponent {

  currentProject: Project;

  constructor(
    private route: ActivatedRoute, 
    private router: Router,
    private sessionService: SessionService) {
    this.route.data.subscribe(data=>this.currentProject = <Project>data['projectResolver']);

  }

  public get isSystemAdmin(): boolean {
    let account = this.sessionService.getCurrentUser();
    return account != null && account.has_admin_role > 0;
  }

  public get isSessionValid(): boolean {
    return this.sessionService.getCurrentUser() != null;
  }

}