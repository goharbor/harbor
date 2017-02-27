import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Params, Router } from '@angular/router';

import { AuditLog } from './audit-log';
import { SessionUser } from '../shared/session-user';

import { AuditLogService } from './audit-log.service';
import { SessionService } from '../shared/session.service';
import { MessageService } from '../global-message/message.service';

export const optionalSearch: {} = {0: 'Advanced', 1: 'Simple'};


export class FilterOption {
  key: string;
  description: string;
  checked: boolean;

  constructor(private iKey: string, private iDescription: string, private iChecked: boolean) {
    this.key = iKey;
    this.description = iDescription;
    this.checked = iChecked;
  }

  toString(): string {
    return 'key:' + this.key + ', description:' + this.description + ', checked:' + this.checked + '\n';
  }
}

@Component({
  selector: 'audit-log',
  templateUrl: './audit-log.component.html',
  styleUrls: [ 'audit-log.css' ]
})
export class AuditLogComponent implements OnInit {

  currentUser: SessionUser;
  projectId: number;
  queryParam: AuditLog = new AuditLog();
  auditLogs: AuditLog[];
 
  toggleName = optionalSearch;
  currentOption: number = 0;
  filterOptions: FilterOption[] = [ 
    new FilterOption('all', 'All Operations', true),
    new FilterOption('pull', 'Pull', true),
    new FilterOption('push', 'Push', true),
    new FilterOption('create', 'Create', true),
    new FilterOption('delete', 'Delete', true),
    new FilterOption('others', 'Others', true) 
 ];

  constructor(private route: ActivatedRoute, private router: Router, private auditLogService: AuditLogService, private messageService: MessageService) {
    //Get current user from registered resolver.
    this.route.data.subscribe(data=>this.currentUser = <SessionUser>data['auditLogResolver']);    
  }

  ngOnInit(): void {
    this.projectId = +this.route.snapshot.parent.params['id'];
    console.log('Get projectId from route params snapshot:' + this.projectId);
    this.queryParam.project_id = this.projectId;
    this.retrieve(this.queryParam);
  }

  retrieve(queryParam: AuditLog): void {
    this.auditLogService
        .listAuditLogs(queryParam)
        .subscribe(
          response=>this.auditLogs = response,
          error=>{
            this.router.navigate(['/harbor', 'projects']);
            this.messageService.announceMessage('Failed to list audit logs with project ID:' + queryParam.project_id);
          }
        );
  }

  doSearchAuditLogs(searchUsername: string): void {
    this.queryParam.username = searchUsername;
    this.retrieve(this.queryParam);
  }

  doSearchByTimeRange(strDate: string, target: string): void {
    let oneDayOffset = 3600 * 24;
    switch(target) {
    case 'begin':
      this.queryParam.begin_timestamp = new Date(strDate).getTime() / 1000;
      break;
    case 'end':
      this.queryParam.end_timestamp = new Date(strDate).getTime() / 1000 + oneDayOffset;
      break;
    }
    console.log('Search audit log filtered by time range, begin: ' + this.queryParam.begin_timestamp + ', end:' + this.queryParam.end_timestamp);
    this.retrieve(this.queryParam);
  }

  doSearchByOptions() {
    let selectAll = true;
    let operationFilter: string[] = [];
    for(var i in this.filterOptions) {
      let filterOption = this.filterOptions[i];
      if(filterOption.checked) {
        operationFilter.push(this.filterOptions[i].key);
      }else{
        selectAll = false;
      }
    }
    if(selectAll) {
      operationFilter = [];
    }
    this.queryParam.keywords = operationFilter.join('/');
    this.retrieve(this.queryParam);
    console.log('Search option filter:' + operationFilter.join('/'));
  }

  toggleOptionalName(option: number): void {
    (option === 1) ? this.currentOption = 0 : this.currentOption = 1;
  }

  toggleFilterOption(option: string): void {
    let selectedOption = this.filterOptions.find(value =>(value.key === option));
    selectedOption.checked = !selectedOption.checked;
    if(selectedOption.key === 'all') {
      this.filterOptions.filter(value=> value.key !== selectedOption.key).forEach(value => value.checked = selectedOption.checked);
    } else {
      if(!selectedOption.checked) {
        this.filterOptions.find(value=>value.key === 'all').checked = false;
      }
      let selectAll = true;
      this.filterOptions.filter(value=> value.key !== 'all').forEach(value =>{
        if(!value.checked) {
          selectAll = false;
        }
      });
      this.filterOptions.find(value=>value.key === 'all').checked = selectAll;
    }
    this.doSearchByOptions();
  }
}