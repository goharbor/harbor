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
import { Component, Input, Output, EventEmitter, ChangeDetectionStrategy, ChangeDetectorRef } from '@angular/core';
import { ReplicationJob } from '../service/interface';
import { State } from 'clarity-angular';
import { ErrorHandler } from '../error-handler/error-handler';

import { REPLICATION_JOB_STYLE } from './list-replication-job.component.css';
import { REPLICATION_JOB_TEMPLATE } from './list-replication-job.component.html';

@Component({
  selector: 'list-replication-job',
  template: REPLICATION_JOB_TEMPLATE,
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class ListReplicationJobComponent {
  @Input() jobs: ReplicationJob[];
  @Input() totalRecordCount: number;
  @Input() totalPage: number;
  @Output() paginate = new EventEmitter<State>();

  constructor(
    private errorHandler: ErrorHandler,
    private ref: ChangeDetectorRef) {
    let hnd = setInterval(()=>ref.markForCheck(), 100);
    setTimeout(()=>clearInterval(hnd), 1000);         
  }

  pageOffset: number = 1;

  refresh(state: State) {
    if(this.jobs) {
      this.paginate.emit(state);
    }
  }
}