// Copyright Project Harbor Authors
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
import { Component, Input, OnInit, Output, EventEmitter } from '@angular/core';
import { AdditionsService } from '../additions.service';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { finalize } from 'rxjs/operators';

@Component({
    selector: 'hbr-artifact-dockerfile',
    templateUrl: './dockerfile.component.html',
    styleUrls: ['./dockerfile.component.scss'],
    standalone: false,
})
export class DockerfileComponent implements OnInit {
    @Input() dockerfileLink: AdditionLink;
    @Input() buildHistoryLink: AdditionLink;
    @Output() switchTab = new EventEmitter<string>();

    dockerfile: string;
    loading: boolean = false;
    fileTooLargeStatus: boolean = false;
    noDockerfileStatus: boolean = false;

    constructor(
        private errorHandler: ErrorHandler,
        private additionsService: AdditionsService
    ) {}

    ngOnInit(): void {
        this.getDockerfile();
    }

    getDockerfile() {
        if (
            this.dockerfileLink &&
            !this.dockerfileLink.absolute &&
            this.dockerfileLink.href
        ) {
            this.loading = true;
            this.additionsService
                .getDetailByLink(this.dockerfileLink.href, false, true)
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    res => {
                        this.dockerfile = res;
                    },
                    error => {
                        if (error.status === 404) {
                            this.noDockerfileStatus = true;
                        } else if (error.status === 413) {
                            this.fileTooLargeStatus = true;
                        } else {
                            this.errorHandler.error(error);
                        }
                    }
                );
        }
    }
}
