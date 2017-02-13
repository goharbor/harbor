import { Component, OnInit } from '@angular/core';
import { Project } from './project';


@Component({
  selector: 'project-list',
  templateUrl: 'project-list.component.html'
})
export class ProjectListComponent implements OnInit {
   projects: Project[];

   ngOnInit(): void {
     this.projects = [
        {
          id: 1,
          name: 'Project01',
          isPublic: true,
          repoCount: 0,
          creationTime: '2016-12-29 14:58 PM',
          destination: '10.117.4.61',
          owner: 'Administrator',
          description: 'New updated by Alice.'
        },
        {
          id: 2,
          name: 'Project02',
          isPublic: false,
          repoCount: 2,
          creationTime: '2016-12-29 15:28 PM',
          destination: '10.117.4.61',
          owner: 'Administrator',
          description: '--'
        },
        {
          id: 3,
          name: 'Project03',
          isPublic: true,
          repoCount: 5,
          creationTime: '2016-12-21 10:25 AM',
          destination: '10.117.4.61',
          owner: 'Administrator',
          description: 'Deprecated'
        },
      ];
   }
}