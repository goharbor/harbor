import { Component, OnInit } from '@angular/core';
import { AuditLog } from './audit-log';

@Component({
  templateUrl: './audit-log.component.html'
})
export class AuditLogComponent implements OnInit {

  auditLogs: AuditLog[];

  ngOnInit(): void {
    this.auditLogs = [
      { username: 'Admin', repoName: 'project01', tag: '', operation: 'create', timestamp: '2016-12-23 12:05:17' },
      { username: 'Admin', repoName: 'project01/ubuntu', tag: '14.04', operation: 'push', timestamp: '2016-12-30 14:52:23' },
      { username: 'user1', repoName: 'project01/mysql', tag: '5.6', operation: 'pull', timestamp: '2016-12-30 12:12:33' }
    ];
  }
}