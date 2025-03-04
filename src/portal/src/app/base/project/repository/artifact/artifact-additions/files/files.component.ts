import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ClrTreeViewModule, ClrIconModule } from '@clr/angular';
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
    imports: [CommonModule, ClrTreeViewModule, ClrIconModule],
})
export class ArtifactFilesComponent {
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
