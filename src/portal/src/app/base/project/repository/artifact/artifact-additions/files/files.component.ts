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
import { finalize } from 'rxjs/operators';
import { AdditionLink } from 'ng-swagger-gen/models';
import { AdditionsService } from '../additions.service';
import { ErrorHandler } from '../../../../../../shared/units/error-handler';
import { FilesItem } from 'src/app/shared/services/interface';
import { formatSize } from 'src/app/shared/units/utils';

@Component({
    selector: 'hbr-artifact-files',
    templateUrl: './files.component.html',
    styleUrls: ['./files.component.scss'],
})
export class ArtifactFilesComponent implements OnInit {
    @Input() filesLink: AdditionLink;
    filesList: FilesItem[] = [];
    loading: Boolean = false;
    expandedNodes: Set<string> = new Set();
    constructor(
        private errorHandler: ErrorHandler,
        private additionsService: AdditionsService
    ) {}

    ngOnInit(): void {
        this.getFiles();
    }
    getFiles() {
        if (this.filesLink && !this.filesLink.absolute && this.filesLink.href) {
            this.loading = true;
            this.additionsService
                .getDetailByLink(this.filesLink.href, false, false)
                .pipe(finalize(() => (this.loading = false)))
                .subscribe(
                    res => {
                        if (res && res.length) {
                            this.filesList = res;
                        }
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
        }
    }

    getChildren(folder: any) {
        return folder.children || [];
    }

    sizeTransform(tagSize: string): string {
        return formatSize(tagSize);
    }

    toggleNodeExpansion(nodeName: string): void {
        if (this.expandedNodes.has(nodeName)) {
            this.expandedNodes.delete(nodeName);
        } else {
            this.expandedNodes.add(nodeName);
        }
    }
}
