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
import { Component, Input, Output, EventEmitter } from '@angular/core';
import { Job } from '../job';
import { State } from 'clarity-angular';
import { MessageHandlerService } from '../../shared/message-handler/message-handler.service';

@Component({
  selector: 'list-job',
  templateUrl: 'list-job.component.html'
})
export class ListJobComponent {
  @Input() jobs: Job[];
  @Input() totalRecordCount: number;
  @Input() totalPage: number;
  @Output() paginate = new EventEmitter<State>();

  constructor(private messageHandlerService: MessageHandlerService) {}

  pageOffset: number = 1;

  refresh(state: State) {
    if(this.jobs) {
      this.paginate.emit(state);
    }
  }
}