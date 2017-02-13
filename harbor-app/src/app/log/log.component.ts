import { Component, OnInit } from '@angular/core';
import { Log } from './log';

@Component({
  templateUrl: './log.component.html'
})
export class LogComponent implements OnInit {

  logs: Log[];

  ngOnInit(): void {
    this.logs = [
      { username: 'Admin', repoName: 'project01', tag: '', operation: 'create', timestamp: '2016-12-23 12:05:17' },
      { username: 'Admin', repoName: 'project01/ubuntu', tag: '14.04', operation: 'push', timestamp: '2016-12-30 14:52:23' },
      { username: 'user1', repoName: 'project01/mysql', tag: '5.6', operation: 'pull', timestamp: '2016-12-30 12:12:33' }
    ];
  }
}