import { Component } from '@angular/core';
import { Project } from '../project';
import { ProjectService } from '../project.service';


@Component({
  selector: 'list-project',
  templateUrl: 'list-project.component.html'
})
export class ListProjectComponent {

   projects: Project[];
   errorMessage: string;

   constructor(private projectService: ProjectService) {}

   retrieve(name: string, isPublic: number): void {
     this.projectService
         .listProjects(name, isPublic)
         .subscribe(
           response => this.projects = response,
           error => this.errorMessage = <any>error);
   }
}