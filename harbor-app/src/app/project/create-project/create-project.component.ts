import { Component, EventEmitter, Output } from '@angular/core';
import { Response } from '@angular/http';

import { Project } from '../project';
import { ProjectService } from '../project.service';

import { MessageService } from '../../global-message/message.service';
import { AlertType } from '../../shared/shared.const';

@Component({
  selector: 'create-project',
  templateUrl: 'create-project.component.html',
  styleUrls: [ 'create-project.css' ]
})
export class CreateProjectComponent {
  
  project: Project = new Project();
  createProjectOpened: boolean;
  
  errorMessage: string;
  hasError: boolean;
  
  @Output() create = new EventEmitter<boolean>();
  
  constructor(private projectService: ProjectService, private messageService: MessageService) {}

  onSubmit() {
    this.hasError = false;
    this.projectService
        .createProject(this.project.name, this.project.public ? 1 : 0)
        .subscribe(
          status=>{
            this.create.emit(true);
            this.createProjectOpened = false;
          },
          error=>{
            this.hasError = true;
            if (error instanceof Response) { 
              switch(error.status) {
              case 409:
                this.errorMessage = 'Project name already exists.'; 
                break;
              case 400:
                this.errorMessage = 'Project name is illegal.'; 
                break;
              default:
                this.errorMessage = 'Unknown error for project name.';
                this.messageService.announceMessage(error.status, this.errorMessage, AlertType.DANGER);
              }
            }
          }); 
  }

  newProject() {
    this.hasError = false;
    this.project = new Project();
    this.createProjectOpened = true;
  }
}

