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
import { ActivatedRoute, Router } from '@angular/router';
import { SessionUser } from '../shared/session-user';
import { MessageHandlerService } from '../shared/message-handler/message-handler.service';
import { ProjectService } from "../../../ng-swagger-gen/services/project.service";
import { AuditLog } from "../../../ng-swagger-gen/models/audit-log";
import { Project } from "../project/project";
import { finalize } from "rxjs/operators";

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
  projectName: string;
  queryUsername: string;
  queryStartTime: string;
  queryEndTime: string;
  queryOperation: string[] = [];
  auditLogs: AuditLog[];
  loading: boolean = true;

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
    private auditLogService: ProjectService,
    private messageHandlerService: MessageHandlerService) {
    // Get current user from registered resolver.
    this.route.data.subscribe(data => this.currentUser = <SessionUser>data['auditLogResolver']);
  }

  ngOnInit(): void {
    const resolverData = this.route.parent.snapshot.data;
    if (resolverData) {
      const pro: Project = <Project>resolverData['projectResolver'];
      this.projectName = pro.name;
    }
  }

  retrieve() {
    const arr: string[] = [];
    if (this.queryUsername) {
      arr.push(`username=~${this.queryUsername}`);
    }
    if (this.queryStartTime && this.queryEndTime) {
      arr.push(`op_time=[${this.queryStartTime}~${this.queryEndTime}]`);
    } else {
      if (this.queryStartTime) {
        arr.push(`op_time=[${this.queryStartTime}~]`);
      }
      if (this.queryEndTime) {
        arr.push(`op_time=[~${this.queryEndTime}]`);
      }
    }
    if (this.queryOperation && this.queryOperation.length > 0) {
      arr.push(`operation={${this.queryOperation.join(' ')}}`);
    }

    const param: ProjectService.GetLogsParams = {
      projectName: this.projectName,
      pageSize: this.pageSize,
      page: this.currentPage,
    };
    if (arr && arr.length > 0) {
      param.q = encodeURIComponent(arr.join(','));
    }
    this.loading = true;
    this.auditLogService
      .getLogsResponse(param)
      .pipe(finalize(() => this.loading = false))
      .subscribe(
        response => {
          // Get total count
          if (response.headers) {
            let xHeader: string = response.headers.get("x-total-count");
            if (xHeader) {
              this.totalRecordCount = Number.parseInt(xHeader);
            }
          }
          this.auditLogs = response.body;
        },
        error => {
          this.messageHandlerService.handleError(error);
        }
      );
  }
  doSearchAuditLogs(searchUsername: string): void {
    this.queryUsername = searchUsername;
    this.retrieve();
  }

  doSearchByStartTime(fromTimestamp: string): void {
    this.queryStartTime = fromTimestamp;
    this.retrieve();
  }

  doSearchByEndTime(toTimestamp: string): void {
    this.queryEndTime = toTimestamp;
    this.retrieve();
  }

  doSearchByOptions() {
    let selectAll = true;
    let operationFilter: string[] = [];
    for (let filterOption of this.filterOptions) {
      if (filterOption.checked) {
        operationFilter.push(filterOption.key);
      } else {
        selectAll = false;
      }
    }
    if (selectAll) {
      operationFilter = [];
    }
    this.queryOperation = operationFilter;
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
