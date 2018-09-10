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
import {
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  Component,
  Input
} from "@angular/core";

import { JobLogService } from "../service/index";
import { ErrorHandler } from "../error-handler/index";
import { toPromise } from "../utils";

const supportSet: string[] = ["replication", "scan"];

@Component({
  selector: "job-log-viewer",
  templateUrl: "./job-log-viewer.component.html",
  styleUrls: ["./job-log-viewer.component.scss"],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class JobLogViewerComponent {
  _jobType: string = "replication";

  opened: boolean = false;
  log: string = "";
  onGoing: boolean = true;

  @Input()
  get jobType(): string {
    return this._jobType;
  }
  set jobType(v: string) {
    if (supportSet.find((t: string) => t === v)) {
      this._jobType = v;
    }
  }

  get title(): string {
    if (this.jobType === "scan") {
      return "VULNERABILITY.JOB_LOG_VIEWER";
    }

    return "REPLICATION.JOB_LOG_VIEWER";
  }

  constructor(
    private jobLogService: JobLogService,
    private errorHandler: ErrorHandler,
    private ref: ChangeDetectorRef
  ) {}

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

    toPromise<string>(this.jobLogService.getJobLog(this.jobType, jobId))
      .then((log: string) => {
        this.onGoing = false;
        this.log = log;
      })
      .catch(error => {
        this.onGoing = false;
        this.errorHandler.error(error);
      });

    let hnd = setInterval(() => this.ref.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 2000);
  }
}
