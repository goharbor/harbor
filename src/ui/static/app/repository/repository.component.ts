import { Component, OnInit } from '@angular/core';
import { Repo } from './repo';


@Component({
  selector: 'repository',
  templateUrl: 'repository.component.html'
})
export class RepositoryComponent implements OnInit {
  repos: Repo[];
  ngOnInit(): void {
    this.repos = [
      { name: 'ubuntu', status: 'ready', tag: '14.04', author: 'Admin', dockerVersion: '1.10.1', created: '2016-10-10', pullCommand: 'docker pull 10.117.5.61/project01/ubuntu:14.04' },
      { name: 'mysql', status: 'ready', tag: '5.6', author: 'docker', dockerVersion: '1.11.2', created: '2016-09-23', pullCommand: 'docker pull 10.117.5.61/project01/mysql:5.6' },
      { name: 'photon', status: 'ready', tag: 'latest', author: 'Admin', dockerVersion: '1.10.1', created: '2016-11-10', pullCommand: 'docker pull 10.117.5.61/project01/photon:latest' },
    ];
  }
}