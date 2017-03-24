import { Component } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

import { Project } from '../project';

import { SessionService } from '../../shared/session.service';
import { ProjectService } from '../../project/project.service';

import { RoleMapping } from '../../shared/shared.const';

@Component({
    selector: 'project-detail',
    templateUrl: "project-detail.component.html",
    styleUrls: [ 'project-detail.component.css' ]
})
export class ProjectDetailComponent {

  hasSignedIn: boolean;
  currentProject: Project;
  
  isMember: boolean;
  roleName: string;

  constructor(
    private route: ActivatedRoute, 
    private router: Router,
    private sessionService: SessionService,
    private projectService: ProjectService) {

    this.hasSignedIn = this.sessionService.getCurrentUser() !== null;
    this.route.data.subscribe(data=>{
      this.currentProject = <Project>data['projectResolver'];
      this.isMember = this.currentProject.is_member;
      this.roleName = RoleMapping[this.currentProject.role_name];
    });
  }

  public get isSystemAdmin(): boolean {
    let account = this.sessionService.getCurrentUser();
    return account != null && account.has_admin_role > 0;
  }

  public get isSessionValid(): boolean {
    return this.sessionService.getCurrentUser() != null;
  }

}