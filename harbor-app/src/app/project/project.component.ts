import { Component, OnInit, ViewChild } from '@angular/core';

import { Router }  from '@angular/router';

import { ListProjectComponent } from './list-project/list-project.component';

@Component({
    selector: 'project',
    templateUrl: 'project.component.html',
    styleUrls: [ 'project.css' ]
})
export class ProjectComponent implements OnInit {
    
    @ViewChild(ListProjectComponent)
    listProjects: ListProjectComponent;
    lastFilteredType: number = 0;

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

    ngOnInit(): void {
      this.listProjects.retrieve('', 0); 
    }

}