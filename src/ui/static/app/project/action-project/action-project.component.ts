import { Component, EventEmitter, Input, Output } from '@angular/core';
import { Project } from '../project';
import { ProjectService } from '../project.service';

@Component({
  selector: 'action-project',
  templateUrl: 'action-project.component.html'
})
export class ActionProjectComponent {

  @Output() togglePublic = new EventEmitter<Project>();
  @Output() deleteProject = new EventEmitter<Project>();

  @Input() project: Project;

  constructor(private projectService: ProjectService) {}

  toggle() {
    if(this.project) {
      this.project.public === 0 ? this.project.public = 1 : this.project.public = 0;
      this.togglePublic.emit(this.project);
    }
  }

  delete() {
    if(this.project) {
      this.deleteProject.emit(this.project);
    }
  }
}