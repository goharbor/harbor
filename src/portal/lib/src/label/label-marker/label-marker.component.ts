
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
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

import { Component, Input, Output, OnInit, EventEmitter, ChangeDetectorRef, ViewChild, ElementRef } from '@angular/core';
import { Observable, fromEvent } from 'rxjs';
import { debounceTime } from 'rxjs/operators';
import { finalize } from 'rxjs/operators';

import { RepositoryItem, HelmChartVersion } from './../../service/interface';
import {Label} from "../../service/interface";
import { ResourceType } from '../../shared/shared.const';
import { LabelService } from '../../service/label.service';
import { TranslateService } from '@ngx-translate/core';
import { ErrorHandler } from '../../error-handler/error-handler';

@Component({
    selector: 'hbr-resource-label-marker',
    templateUrl: './label-marker.component.html',
    styleUrls: ['./label-marker.component.scss']
})

export class LabelMarkerComponent implements OnInit {

    @Input() labels: Label[] = [];
    @Input() projectName: string;
    @Input() resource: RepositoryItem | HelmChartVersion;
    @Input() resourceType: ResourceType;

    labelFilter = '';
    markedMap: Map<number, boolean> = new Map<number, boolean>();
    markingMap: Map<number, boolean> = new Map<number, boolean>();
    sortedLabels: Label[] = [];

    loading = false;

    @ViewChild('filterInput') filterInputRef: ElementRef;

    ngOnInit(): void {
        this.sortedLabels = this.labels;
        this.refresh();
        fromEvent(this.filterInputRef.nativeElement, 'keyup')
        .pipe(debounceTime(500))
        .subscribe(() => this.refresh());
    }

    constructor(
        private labelService: LabelService,
        private errorHandler: ErrorHandler,
        private translateService: TranslateService,
        private cdr: ChangeDetectorRef) {}

    refresh() {
        this.loading = true;
        if (this.resourceType === ResourceType.CHART_VERSION) {
            this.labelService.getChartVersionLabels(
                this.projectName,
                this.resource.name,
                (this.resource as HelmChartVersion).version)
            .pipe(finalize(() => {
                    this.loading = false;
                    let hnd = setInterval(() => this.cdr.markForCheck(), 100);
                    setTimeout(() => clearInterval(hnd), 2000);
                  }))
            .subscribe( chartVersionLabels => {
                for (let label of chartVersionLabels) {
                    console.log('marked label', label);
                    this.markedMap.set(label.id, true);
                }
                this.sortedLabels = this.getSortedLabels();
            });
        }
    }

    markLabel(label: Label) {
        if (this.markedMap.get(label.id) || this.isMarkOngoing(label)) {
            return;
        }
        this.markingMap.set(label.id, true);
        this.labelService.markChartLabel(
            this.projectName,
            this.resource.name,
            (this.resource as HelmChartVersion).version,
            label)
            .pipe(finalize(() => {
                this.markingMap.set(label.id, false);
                let hnd = setInterval(() => this.cdr.markForCheck(), 100);
                setTimeout(() => clearInterval(hnd), 5000);
            }))
            .subscribe(
                () => {
                    this.markedMap.set(label.id, true);
                    this.refresh();
                    let hnd = setInterval(() => this.cdr.markForCheck(), 100);
                    setTimeout(() => clearInterval(hnd), 5000);
                },
                err => this.errorHandler.error(err)
            );
    }

    unmarkLabel(label: Label) {
        if (!this.isMarked(label) || this.isMarkOngoing(label)) {
            return;
        }
        this.markingMap.set(label.id, true);
        this.labelService.unmarkChartLabel(
            this.projectName,
            this.resource.name,
            (this.resource as HelmChartVersion).version,
            label)
            .pipe(finalize(() => {
                this.markingMap.set(label.id, false);
                let hnd = setInterval(() => this.cdr.markForCheck(), 100);
                setTimeout(() => clearInterval(hnd), 5000);
            }))
            .subscribe(
                () => {
                    this.markedMap.set(label.id, false);
                    this.refresh();
                    let hnd = setInterval(() => this.cdr.markForCheck(), 100);
                    setTimeout(() => clearInterval(hnd), 5000);
                },
                err => this.errorHandler.error(err)
            );
    }

    isMarked(label: Label) {
        return this.markedMap.get(label.id) ? true : false;
    }

    isMarkOngoing(label: Label) {
        return this.markingMap.get(label.id) ? true : false;
    }

    getSortedLabels(): Label[] {
        return this.labels.filter( l => l.name.includes(this.labelFilter))
        .sort((a, b) => {
            if (this.isMarked(a) && !this.isMarked(b)) {
                return -1;
            } else if (!this.isMarked(a) && this.isMarked(b)) {
                return 1;
            } else {
                return a.name < b.name ? -1 : a.name > b.name ? 1 : 0;
            }
        });
    }
}
