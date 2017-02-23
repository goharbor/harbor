import { Component, OnInit, ViewChild } from '@angular/core';

import { Router }  from '@angular/router';

import { Project } from './project';
import { ProjectService } from './project.service';

import { CreateProjectComponent } from './create-project/create-project.component';

import { ListProjectComponent } from './list-project/list-project.component';

import { MessageService } from '../global-message/message.service';

export const types: {} = { 0: 'My Projects', 1: 'Public Projects'};

@Component({
    selector: 'project',
    templateUrl: 'project.component.html',
    styleUrls: [ 'project.css' ]
})
export class ProjectComponent implements OnInit {
    
  selected = [];
  changedProjects: Project[];
  projectTypes = types;
  
  @ViewChild(CreateProjectComponent)
  creationProject: CreateProjectComponent;

  @ViewChild(ListProjectComponent)
  listProject: ListProjectComponent;

  currentFilteredType: number = 0;
  lastFilteredType: number = 0;

  constructor(private projectService: ProjectService, private messageService: MessageService){}

  ngOnInit(): void {
    this.retrieve('', this.lastFilteredType);
  }

  retrieve(name: string, isPublic: number): void {
    this.projectService
        .listProjects(name, isPublic)
        .subscribe(
          response => this.changedProjects = response,
          error => this.messageService.announceMessage(error));
  }

  openModal(): void {
    this.creationProject.newProject();
  }
  
  createProject(created: boolean) {
    if(created) {
      this.retrieve('', this.lastFilteredType);
    }
  }

  doSearchProjects(projectName: string): void {
    console.log('Search for project name:' + projectName);
    this.retrieve(projectName, this.lastFilteredType);
  }

  doFilterProjects(filteredType: number): void {
    console.log('Filter projects with type:' + types[filteredType]);
    this.lastFilteredType = filteredType;
    this.currentFilteredType = filteredType;
    this.retrieve('', this.lastFilteredType);
  }

  toggleProject(p: Project) {
    this.projectService
        .toggleProjectPublic(p.project_id, p.public)
        .subscribe(
          response=>console.log('Successful toggled project_id:' + p.project_id),
          error=>this.messageService.announceMessage(error)
        );
  }

  deleteProject(p: Project) {
    this.projectService
        .deleteProject(p.project_id)
        .subscribe(
          response=>{
            console.log('Successful delete project_id:' + p.project_id);
            this.retrieve('', this.lastFilteredType);
          },
          error=>console.log(error)
        );
  }

}