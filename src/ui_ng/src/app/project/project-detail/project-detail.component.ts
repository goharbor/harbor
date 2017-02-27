import { Component } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

import { Project } from '../project';

@Component({
    selector: 'project-detail',
    templateUrl: "project-detail.component.html",
    styleUrls: [ 'project-detail.css' ]
})
export class ProjectDetailComponent {

  currentProject: Project;

  constructor(private route: ActivatedRoute, private router: Router) {
    this.route.data.subscribe(data=>this.currentProject = <Project>data['projectResolver']);
  }
}