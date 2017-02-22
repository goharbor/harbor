import { Component, OnInit, ViewChild } from '@angular/core';

import { Router }  from '@angular/router';

import { ListProjectComponent } from './list-project/list-project.component';
import { CreateProjectComponent } from './create-project/create-project.component';

@Component({
    selector: 'project',
    templateUrl: 'project.component.html',
    styleUrls: [ 'project.css' ]
})
export class ProjectComponent implements OnInit {
    
    @ViewChild(ListProjectComponent)
    listProjects: ListProjectComponent;

    @ViewChild(CreateProjectComponent)
    creationProject: CreateProjectComponent;

    lastFilteredType: number = 0;

    openModal(): void {
      this.creationProject.newProject();
    }

    deleteSelectedProjects(): void {
      this.listProjects.deleteSelectedProjects();
    }

    createProject(created: boolean): void {
      console.log('Project has been created:' + created);
      this.listProjects.retrieve('', 0);
    }

    filterProjects(type: number): void {
      this.lastFilteredType = type;
      this.listProjects.retrieve('', type);
      console.log('Projects were filtered by:' + type);
      
    }

    searchProjects(projectName: string): void {
      console.log('Search for project name:' + projectName);
      this.listProjects.retrieve(projectName, this.lastFilteredType);
    }

    actionPerform(performed: boolean): void {
      this.listProjects.retrieve('', 0);
    }

    ngOnInit(): void {
      this.listProjects.retrieve('', 0); 
    }

}