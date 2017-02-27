import { Injectable } from '@angular/core';
import { Router, Resolve, RouterStateSnapshot, ActivatedRouteSnapshot } from '@angular/router';

import { Project } from './project';
import { ProjectService } from './project.service';

@Injectable()
export class ProjectRoutingResolver implements Resolve<Project>{

  constructor(private projectService: ProjectService, private router: Router) {}

  resolve(route: ActivatedRouteSnapshot, state: RouterStateSnapshot): Promise<Project> {
    let projectId = route.params['id'];
    return this.projectService
               .getProject(projectId)
               .then(project=> {
                 if(project) {
                   return project;
                 } else {
                   this.router.navigate(['/harbor', 'projects']);
                   return null;
                 }
               });
  } 
}