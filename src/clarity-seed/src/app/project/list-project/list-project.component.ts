import { Component, EventEmitter, Output } from '@angular/core';
import { Project } from '../project';
import { ProjectService } from '../project.service';


@Component({
  selector: 'list-project',
  templateUrl: 'list-project.component.html'
})
export class ListProjectComponent {

   projects: Project[];
   errorMessage: string;

   selected = [];

   @Output() actionPerform = new EventEmitter<boolean>();

   constructor(private projectService: ProjectService) {}

   retrieve(name: string, isPublic: number): void {
     this.projectService
         .listProjects(name, isPublic)
         .subscribe(
           response => this.projects = response,
           error => this.errorMessage = <any>error);
   }

   toggleProject(p: Project) {
     this.projectService
         .toggleProjectPublic(p.project_id, p.public)
         .subscribe(
           response=>console.log(response),
           error=>console.log(error)
         );
   }

   deleteProject(p: Project) {
     this.projectService
         .deleteProject(p.project_id)
         .subscribe(
           response=>{
             console.log(response);
             this.actionPerform.emit(true);
           },
           error=>console.log(error)
         );
   }

   deleteSelectedProjects() {
     this.selected.forEach(p=>this.deleteProject(p));
   }

   onEdit(p: Project) {

   }

   onDelete(p: Project) {

   }
}