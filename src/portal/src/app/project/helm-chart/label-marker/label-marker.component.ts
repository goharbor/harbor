import { Component, Input, Output, OnInit, EventEmitter, ChangeDetectorRef, ViewChild, ElementRef } from '@angular/core';
import { fromEvent, Subject } from 'rxjs';
import { debounceTime, finalize } from 'rxjs/operators';

import { HelmChartVersion } from '../helm-chart.interface.service';
import { Label, LabelService, ErrorHandler } from '@harbor/ui';
import { ResourceType } from '../../../shared/shared.const';

@Component({
    selector: 'hbr-resource-label-marker',
    templateUrl: './label-marker.component.html',
    styleUrls: ['./label-marker.component.scss']
})

export class LabelMarkerComponent implements OnInit {

    @Input() labels: Label[] = [];
    @Input() projectName: string;
    @Input() resource: HelmChartVersion;
    @Input() resourceType: ResourceType;
    @Input() addLabelHeaders: string;
    @Output() changeEvt = new EventEmitter<any>();

    labelFilter = '';
    markedMap: Map<number, boolean> = new Map<number, boolean>();
    markingMap: Map<number, boolean> = new Map<number, boolean>();
    sortedLabels: Label[] = [];

    loading = false;

    labelChangeDebouncer: Subject<any> = new Subject();

    @ViewChild('filterInput', {static: true}) filterInputRef: ElementRef;

    ngOnInit(): void {
        this.sortedLabels = this.labels;
        this.refresh();
        fromEvent(this.filterInputRef.nativeElement, 'keyup')
            .pipe(debounceTime(500))
            .subscribe(() => this.refresh());

        this.labelChangeDebouncer.pipe(debounceTime(1000)).subscribe(() => this.changeEvt.emit());
    }

    constructor(
        private labelService: LabelService,
        private errorHandler: ErrorHandler,
        private cdr: ChangeDetectorRef) { }

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
                .subscribe(chartVersionLabels => {
                    for (let label of chartVersionLabels) {
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
                    this.labelChangeDebouncer.next();
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
                    this.labelChangeDebouncer.next();
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
        return this.labels.filter(l => l.name.includes(this.labelFilter))
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
