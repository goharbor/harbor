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
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Params, Router } from '@angular/router';

import { AuditLog } from './audit-log';
import { SessionUser } from '../shared/session-user';

import { AuditLogService } from './audit-log.service';
import { SessionService } from '../shared/session.service';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { AlertType } from '../shared/shared.const';

import { State } from 'clarity-angular';

const optionalSearch: {} = {0: 'AUDIT_LOG.ADVANCED', 1: 'AUDIT_LOG.SIMPLE'};

class FilterOption {
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
  styleUrls: [ './audit-log.component.css' ]
})
export class AuditLogComponent implements OnInit {

  currentUser: SessionUser;
  projectId: number;
  queryParam: AuditLog = new AuditLog();
  auditLogs: AuditLog[];
 
  toggleName = optionalSearch;
  currentOption: number = 0;
  filterOptions: FilterOption[] = [ 
    new FilterOption('all', 'AUDIT_LOG.ALL_OPERATIONS', true),
    new FilterOption('pull', 'AUDIT_LOG.PULL', true),
    new FilterOption('push', 'AUDIT_LOG.PUSH', true),
    new FilterOption('create', 'AUDIT_LOG.CREATE', true),
    new FilterOption('delete', 'AUDIT_LOG.DELETE', true),
    new FilterOption('others', 'AUDIT_LOG.OTHERS', true) 
 ];

  pageOffset: number = 1;
  pageSize: number = 15;
  totalRecordCount: number;
  totalPage: number;
  
  constructor(private route: ActivatedRoute, private router: Router, private auditLogService: AuditLogService, private messageHandlerService: MessageHandlerService) {
    //Get current user from registered resolver.
    this.route.data.subscribe(data=>this.currentUser = <SessionUser>data['auditLogResolver']);    
  }

  ngOnInit(): void {
    this.projectId = +this.route.snapshot.parent.params['id'];
    console.log('Get projectId from route params snapshot:' + this.projectId);
    this.queryParam.project_id = this.projectId;
    this.queryParam.page_size = this.pageSize;
  }

  retrieve(state?: State): void {
    if(state) {
      this.queryParam.page = state.page.to + 1;
    }
    this.auditLogService
        .listAuditLogs(this.queryParam)
        .subscribe(
          response=>{
            this.totalRecordCount = response.headers.get('x-total-count');
            this.totalPage = Math.ceil(this.totalRecordCount / this.pageSize);
            console.log('TotalRecordCount:' + this.totalRecordCount + ', totalPage:' + this.totalPage);
            this.auditLogs = response.json();
          },
          error=>{
            this.router.navigate(['/harbor', 'projects']);
            this.messageHandlerService.handleError(error);
          }
        );
  }

  doSearchAuditLogs(searchUsername: string): void {
    this.queryParam.username = searchUsername;
    this.retrieve();
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
    this.retrieve();
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
    this.retrieve();
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
  refresh(): void {
    this.retrieve();
  }
}