import { Component, EventEmitter, Output, Input } from '@angular/core';
import { Project } from '../project';
import { ProjectService } from '../project.service';


@Component({
  selector: 'list-project',
  templateUrl: 'list-project.component.html'
})
export class ListProjectComponent {

   @Input() projects: Project[];

   @Output() toggle = new EventEmitter<Project>();
   @Output() delete = new EventEmitter<Project>();

   toggleProject(p: Project) {
     this.toggle.emit(p);
   }

   deleteProject(p: Project) {
     this.delete.emit(p);
   }

}