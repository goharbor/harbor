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
import { Component, EventEmitter, Output, ViewChild, AfterViewChecked, HostBinding } from '@angular/core';
import { Response } from '@angular/http';
import { NgForm } from '@angular/forms';

import { Project } from '../project';
import { ProjectService } from '../project.service';

import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';
import { InlineAlertComponent } from '../../shared/inline-alert/inline-alert.component';

import { TranslateService } from '@ngx-translate/core';

@Component({
  selector: 'create-project',
  templateUrl: 'create-project.component.html',
  styleUrls: [ 'create-project.css' ]
})
export class CreateProjectComponent implements AfterViewChecked {
  
  projectForm: NgForm;

  @ViewChild('projectForm')
  currentForm: NgForm;

  project: Project = new Project();
  initVal: Project = new Project();

  createProjectOpened: boolean;
  
  hasChanged: boolean;

  staticBackdrop: boolean = true;
  closable: boolean = false;

  @Output() create = new EventEmitter<boolean>();
  @ViewChild(InlineAlertComponent)
  private inlineAlert: InlineAlertComponent;

  constructor(private projectService: ProjectService,             
              private translateService: TranslateService,
              private messageHandlerService: MessageHandlerService) {}

  onSubmit() {
    this.projectService
        .createProject(this.project.name, this.project.public ? 1 : 0)
        .subscribe(
          status=>{
            this.create.emit(true);
            this.messageHandlerService.showSuccess('PROJECT.CREATED_SUCCESS');
            this.createProjectOpened = false;
          },
          error=>{
            let errorMessage: string;
            if (error instanceof Response) { 
              switch(error.status) {
              case 409:
                this.translateService.get('PROJECT.NAME_ALREADY_EXISTS').subscribe(res=>errorMessage = res);
                break;
              case 400:
                this.translateService.get('PROJECT.NAME_IS_ILLEGAL').subscribe(res=>errorMessage = res); 
                break;
              default:
                this.translateService.get('PROJECT.UNKNOWN_ERROR').subscribe(res=>errorMessage = res);
              }
              if(this.messageHandlerService.isAppLevel(error)) {
                this.messageHandlerService.handleError(error);
                this.createProjectOpened = false;
              } else {
                this.inlineAlert.showInlineError(errorMessage);
              }
            }
          }); 
  }

  onCancel() {
    if(this.hasChanged) {
      this.inlineAlert.showInlineConfirmation({message: 'ALERT.FORM_CHANGE_CONFIRMATION'});
    } else {
      this.createProjectOpened = false;
      this.projectForm.reset();
    }
   
  }

  ngAfterViewChecked(): void {
    this.projectForm = this.currentForm;
    if(this.projectForm) {
      this.projectForm.valueChanges.subscribe(data=>{
        for(let i in data) {
          let origin = this.initVal[i];          
          let current = data[i];
          if(current && current !== origin) {
            this.hasChanged = true;
            break;
          } else {
            this.hasChanged = false;
            this.inlineAlert.close();
          }
        }
      });
    }
  }

  newProject() {
    this.project = new Project();
    this.hasChanged = false;
    this.createProjectOpened = true;
  }

  confirmCancel(event: boolean): void {
    this.createProjectOpened = false;
    this.inlineAlert.close();
    this.projectForm.reset();
  }
}

