// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
                    if(currentUser) {
                      let projectMembers = this.sessionService.getProjectMembers();
                      if(projectMembers) {
                        let currentMember = projectMembers.find(m=>m.user_id === currentUser.user_id);
                        if(currentMember) {
                          project.is_member = true;
                          project.has_project_admin_role = (currentMember.role_name === 'projectAdmin');
                          project.role_name = currentMember.role_name;
                        } 
                      }
                      if(currentUser.has_admin_role === 1) {
                        project.has_project_admin_role = true;
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