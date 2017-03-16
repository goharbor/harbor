import { Component } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

import { Project } from '../project';

import { SessionService } from '../../shared/session.service';

@Component({
    selector: 'project-detail',
    templateUrl: "project-detail.component.html",
    styleUrls: [ 'project-detail.css' ]
})
export class ProjectDetailComponent {

  currentProject: Project;

  constructor(
    private route: ActivatedRoute, 
    private router: Router,
    private session: SessionService) {
    this.route.data.subscribe(data=>this.currentProject = <Project>data['projectResolver']);
  }

  public get isSessionValid(): boolean {
    return this.session.getCurrentUser() != null;
  }
}