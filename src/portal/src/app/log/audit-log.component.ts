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
import { Component, OnInit, ViewChild } from '@angular/core';
import { NgModel } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';

import { AuditLog } from './audit-log';
import { SessionUser } from '../shared/session-user';

import { AuditLogService } from './audit-log.service';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';

import { State } from '../../../lib/src/service/interface';

const optionalSearch: {} = { 0: 'AUDIT_LOG.ADVANCED', 1: 'AUDIT_LOG.SIMPLE' };

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

export class SearchOption {
  startTime: string = "";
  endTime: string = "";
}

@Component({
  selector: 'audit-log',
  templateUrl: './audit-log.component.html',
  styleUrls: ['./audit-log.component.scss']
})
export class AuditLogComponent implements OnInit {
  search: SearchOption = new SearchOption();
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

  pageOffset = 1;
  pageSize = 15;
  totalRecordCount = 0;
  currentPage = 1;
  totalPage = 0;

  get showPaginationIndex(): boolean {
    return this.totalRecordCount > 0;
  }

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private auditLogService: AuditLogService,
    private messageHandlerService: MessageHandlerService) {
    // Get current user from registered resolver.
    this.route.data.subscribe(data => this.currentUser = <SessionUser>data['auditLogResolver']);
  }

  ngOnInit(): void {
    this.projectId = +this.route.snapshot.parent.params['id'];
    this.queryParam.project_id = this.projectId;
    this.queryParam.page_size = this.pageSize;

  }

  private retrieve(): void {
    this.auditLogService
      .listAuditLogs(this.queryParam)
      .subscribe(
        response => {
          this.totalRecordCount = Number.parseInt(response.headers.get('x-total-count'));
          this.auditLogs = response.body;
        },
        error => {
          this.messageHandlerService.handleError(error);
        }
      );
  }

  retrievePage(state: State) {
    if (state && state.page) {
      this.queryParam.page = Math.ceil((state.page.to + 1) / this.pageSize);
      this.currentPage = this.queryParam.page;
      this.retrieve();
    }
  }

  doSearchAuditLogs(searchUsername: string): void {
    this.queryParam.username = searchUsername;
    this.retrieve();
  }

  doSearchByStartTime(fromTimestamp: string): void {
    this.queryParam.begin_timestamp = fromTimestamp;
    this.retrieve();
  }

  doSearchByEndTime(toTimestamp: string): void {
    this.queryParam.end_timestamp = toTimestamp;
    this.retrieve();
  }

  doSearchByOptions() {
    let selectAll = true;
    let operationFilter: string[] = [];
    for (let filterOption of this.filterOptions) {
      if (filterOption.checked) {
        operationFilter.push('operation=' + filterOption.key);
      } else {
        selectAll = false;
      }
    }
    if (selectAll) {
      operationFilter = [];
    }
    this.queryParam.keywords = operationFilter.join('&');
    this.retrieve();
  }

  toggleOptionalName(option: number): void {
    (option === 1) ? this.currentOption = 0 : this.currentOption = 1;
  }

  toggleFilterOption(option: string): void {
    let selectedOption = this.filterOptions.find(value => (value.key === option));
    selectedOption.checked = !selectedOption.checked;
    if (selectedOption.key === 'all') {
      this.filterOptions.filter(value => value.key !== selectedOption.key).forEach(value => value.checked = selectedOption.checked);
    } else {
      if (!selectedOption.checked) {
        this.filterOptions.find(value => value.key === 'all').checked = false;
      }
      let selectAll = true;
      this.filterOptions.filter(value => value.key !== 'all').forEach(value => {
        if (!value.checked) {
          selectAll = false;
        }
      });
      this.filterOptions.find(value => value.key === 'all').checked = selectAll;
    }
    this.doSearchByOptions();
  }
  refresh(): void {
    this.retrieve();
  }
}
