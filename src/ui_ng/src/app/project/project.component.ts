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
import { Component, OnInit, ViewChild, OnDestroy } from '@angular/core';

import { Router } from '@angular/router';

import { Project } from './project';
import { ProjectService } from './project.service';

import { CreateProjectComponent } from './create-project/create-project.component';

import { ListProjectComponent } from './list-project/list-project.component';

import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { Message } from '../global-message/message';

import { Response } from '@angular/http';

import { ConfirmationDialogService } from '../shared/confirmation-dialog/confirmation-dialog.service';
import { ConfirmationMessage } from '../shared/confirmation-dialog/confirmation-message';
import { ConfirmationTargets, ConfirmationState } from '../shared/shared.const';

import { Subscription } from 'rxjs/Subscription';

import { State } from 'clarity-angular';

import { AppConfigService } from '../app-config.service';
import { SessionService } from '../shared/session.service';
import { ProjectTypes } from '../shared/shared.const';
import { StatisticHandler } from '../shared/statictics/statistic-handler.service';

@Component({
  selector: 'project',
  templateUrl: 'project.component.html',
  styleUrls: ['./project.component.css']
})
export class ProjectComponent implements OnInit, OnDestroy {

  selected = [];
  changedProjects: Project[];
  projectTypes = ProjectTypes;

  @ViewChild(CreateProjectComponent)
  creationProject: CreateProjectComponent;

  @ViewChild(ListProjectComponent)
  listProject: ListProjectComponent;

  currentFilteredType: number = 0;

  subscription: Subscription;

  projectName: string;
  isPublic: number;

  page: number = 1;
  pageSize: number = 15;

  totalPage: number;
  totalRecordCount: number;

  constructor(
    private projectService: ProjectService,
    private messageHandlerService: MessageHandlerService,
    private appConfigService: AppConfigService,
    private sessionService: SessionService,
    private deletionDialogService: ConfirmationDialogService,
    private statisticHandler: StatisticHandler) {
    this.subscription = deletionDialogService.confirmationConfirm$.subscribe(message => {
      if (message &&
        message.state === ConfirmationState.CONFIRMED &&
        message.source === ConfirmationTargets.PROJECT) {
        let projectId = message.data;
        this.projectService
          .deleteProject(projectId)
          .subscribe(
          response => {
            this.messageHandlerService.showSuccess('PROJECT.DELETED_SUCCESS');
            this.retrieve();
            this.statisticHandler.refresh();
          },
          error =>{
            if(error && error.status === 412) {
              this.messageHandlerService.showError('PROJECT.FAILED_TO_DELETE_PROJECT', '');
            } else {
              this.messageHandlerService.handleError(error);
            }
          }
        );
      }
    });
    
  }

  ngOnInit(): void {
    this.projectName = '';
    this.isPublic = 0;
    
  }

  ngOnDestroy(): void {
    if (this.subscription) {
      this.subscription.unsubscribe();
    }
  }

  get projectCreationRestriction(): boolean {
    let account = this.sessionService.getCurrentUser();
    if(account) {
      switch(this.appConfigService.getConfig().project_creation_restriction) {
      case 'adminonly':
        return (account.has_admin_role === 1);
      case 'everyone':
        return true;
      } 
    }
    return false;
  }

  retrieve(state?: State): void {
    if (state) {
      this.page = state.page.to + 1;
    }
    this.projectService
      .listProjects(this.projectName, this.isPublic, this.page, this.pageSize)
      .subscribe(
      response => {
        this.totalRecordCount = response.headers.get('x-total-count');
        this.totalPage = Math.ceil(this.totalRecordCount / this.pageSize);
        this.changedProjects = response.json();
      },
      error => this.messageHandlerService.handleError(error)
    );
  }

  openModal(): void {
    this.creationProject.newProject();
  }

  createProject(created: boolean) {
    if (created) {
      this.projectName = '';
      this.retrieve();
      this.statisticHandler.refresh();
    }
  }

  doSearchProjects(projectName: string): void {
    this.projectName = projectName;
    this.retrieve();
  }

  doFilterProjects($event: any): void {
    if ($event && $event.target && $event.target["value"]) {
      this.currentFilteredType = $event.target["value"];
      this.isPublic = this.currentFilteredType;
      this.retrieve();
    }
  }

  toggleProject(p: Project) {
    if (p) {
      p.public === 0 ? p.public = 1 : p.public = 0;
      this.projectService
        .toggleProjectPublic(p.project_id, p.public)
        .subscribe(
        response => {
          this.messageHandlerService.showSuccess('PROJECT.TOGGLED_SUCCESS');
          this.statisticHandler.refresh();
        },
        error => this.messageHandlerService.handleError(error)
        );
    }
  }

  deleteProject(p: Project) {
    let deletionMessage = new ConfirmationMessage(
      'PROJECT.DELETION_TITLE',
      'PROJECT.DELETION_SUMMARY',
      p.name,
      p.project_id,
      ConfirmationTargets.PROJECT
    );
    this.deletionDialogService.openComfirmDialog(deletionMessage);
  }

  refresh(): void {
    this.retrieve();
    this.statisticHandler.refresh();
  }

}