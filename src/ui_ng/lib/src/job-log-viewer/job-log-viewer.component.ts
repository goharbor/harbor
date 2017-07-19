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
import { Component } from '@angular/core';

import { JOB_LOG_VIEWER_TEMPLATE, JOB_LOG_VIEWER_STYLES } from './job-log-viewer.component.template';
import { ReplicationService } from '../service/index';
import { ErrorHandler } from '../error-handler/index';
import { toPromise } from '../utils';

@Component({
    selector: 'job-log-viewer',
    template: JOB_LOG_VIEWER_TEMPLATE,
    styles: [JOB_LOG_VIEWER_STYLES]
})

export class JobLogViewerComponent {
    opened: boolean = false;
    log: string = '';
    onGoing: boolean = true;

    constructor(
        private replicationService: ReplicationService,
        private errorHandler: ErrorHandler
    ) { }

    open(jobId: number | string): void {
        this.opened = true;
        this.load(jobId);
    }

    close(): void {
        this.opened = false;
        this.log = "";
    }

    load(jobId: number | string): void {
        this.onGoing = true;

        toPromise<string>(this.replicationService.getJobLog(jobId))
            .then((log: string) => {
                this.onGoing = false;
                this.log = log;
            })
            .catch(error => {
                this.onGoing = false;
                this.errorHandler.error(error);
            });
    }
}