import { Component, EventEmitter, Output, ViewChild } from '@angular/core';
import { Response } from '@angular/http';

import { Project } from '../project';
import { ProjectService } from '../project.service';


import { MessageService } from '../../global-message/message.service';
import { AlertType } from '../../shared/shared.const';

import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';

import { TranslateService } from '@ngx-translate/core';

@Component({
  selector: 'create-project',
  templateUrl: 'create-project.component.html',
  styleUrls: [ 'create-project.css' ]
})
export class CreateProjectComponent {
  
  project: Project = new Project();
  createProjectOpened: boolean;
  
  errorMessageOpened: boolean;
  errorMessage: string;

  @Output() create = new EventEmitter<boolean>();
  @ViewChild(InlineAlertComponent)
  private inlineAlert: InlineAlertComponent;

  constructor(private projectService: ProjectService, 
              private messageService: MessageService,
              private translateService: TranslateService) {}

  onSubmit() {
    this.projectService
        .createProject(this.project.name, this.project.public ? 1 : 0)
        .subscribe(
          status=>{
            this.create.emit(true);
            this.createProjectOpened = false;
          },
          error=>{
            this.errorMessageOpened = true;
            if (error instanceof Response) { 
              switch(error.status) {
              case 409:
                this.translateService.get('PROJECT.NAME_ALREADY_EXISTS').subscribe(res=>this.errorMessage = res);
                break;
              case 400:
                this.translateService.get('PROJECT.NAME_IS_ILLEGAL').subscribe(res=>this.errorMessage = res); 
                break;
              default:
                this.translateService.get('PROJECT.UNKNOWN_ERROR').subscribe(res=>{
                  this.errorMessage = res;
                  this.messageService.announceMessage(error.status, this.errorMessage, AlertType.DANGER);
                });
              }
              this.inlineAlert.showInlineError(this.errorMessage);
            }
          }); 
  }

  newProject() {
    this.project = new Project();
    this.createProjectOpened = true;
    this.errorMessageOpened = false;
    this.errorMessage = '';
  }

  onErrorMessageClose(): void {
    this.errorMessageOpened = false;
    this.errorMessage = '';
  }

  confirmCancel(event: boolean): void {
    this.errorMessageOpened = false;
  }

}

