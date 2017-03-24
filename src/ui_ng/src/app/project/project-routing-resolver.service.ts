import { Injectable } from '@angular/core';
import { Router, Resolve, RouterStateSnapshot, ActivatedRouteSnapshot } from '@angular/router';

import { Project } from './project';
import { ProjectService } from './project.service';
import { SessionService } from '../shared/session.service';
import 'rxjs/add/operator/mergeMap';

@Injectable()
export class ProjectRoutingResolver implements Resolve<Project>{

  constructor(
    private sessionService: SessionService,
    private projectService: ProjectService, 
    private router: Router) {}

  resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<Project> {
    let projectId = route.params['id'];  
    console.log('Project resolver, projectID:' + projectId);
    return this.projectService
               .getProject(projectId)
               .toPromise()
               .then((project: Project)=> {
                  if(project) {
                    let currentUser = this.sessionService.getCurrentUser();
                    let projectMembers = this.sessionService.getProjectMembers();
                    if(currentUser && projectMembers) {
                      let currentMember = projectMembers.find(m=>m.user_id === currentUser.user_id);
                      if(currentMember) {
                        project.is_member = true;
                        project.has_project_admin_role = (currentMember.role_name === 'projectAdmin') || currentUser.has_admin_role === 1;
                      } 
                    }
                    return project;
                  } else {
                    this.router.navigate(['/harbor', 'projects']);
                    return null;
                  }
               }).catch(error=>{
                 this.router.navigate(['/harbor', 'projects']);
                 return null;
               });
               
  } 
}