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
import { Component, Input, OnInit } from '@angular/core';
import { AdditionsService } from '../additions.service';
import { AdditionLink } from '../../../../../../../../ng-swagger-gen/models/addition-link';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { finalize } from 'rxjs/operators';
import { Artifact } from 'ng-swagger-gen/models/artifact';

@Component({
    selector: 'hbr-artifact-summary',
    templateUrl: './summary.component.html',
    styleUrls: ['./summary.component.scss'],
})
export class SummaryComponent implements OnInit {
    @Input() summaryLink: AdditionLink;
    @Input() artifactDetails: Artifact;
    readme: string;
    type: string;
    loading: boolean = false;
    constructor(
        private errorHandler: ErrorHandler,
        private additionsService: AdditionsService
    ) {}

    ngOnInit(): void {
        this.getReadme();
        if (this.artifactDetails) {
            this.type = this.artifactDetails.type;
        }
    }
    getReadme() {
        if (
            this.summaryLink &&
            !this.summaryLink.absolute &&
            this.summaryLink.href
        ) {
            this.loading = true;
            this.additionsService
                .getDetailByLink(this.summaryLink.href, false, true)
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    res => {
                        this.readme = this.removeFrontMatter(res);
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }

    removeFrontMatter(content: string): string {
        return content.replace(/^---[\s\S]*?---\s*/, '');
    }
}
