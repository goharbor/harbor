import { OnInit, Input, EventEmitter, Component, ViewChild, ElementRef, ChangeDetectorRef } from '@angular/core';
import {ClrDatagridFilterInterface} from "@clr/angular";
import { fromEvent } from 'rxjs';
import { debounceTime } from 'rxjs/operators';

import { Label, Tag } from '@harbor/ui';
import { HelmChartVersion } from '../helm-chart.interface.service';
import { ResourceType } from '../../../shared/shared.const';

@Component({
    selector: "hbr-chart-version-label-filter",
    templateUrl: './label-filter.component.html',
    styleUrls: ['./label-filter.component.scss']
})
export class LabelFilterComponent implements ClrDatagridFilterInterface<any>, OnInit {

    @Input() labels: Label[] = [];
    @Input() resourceType: ResourceType;

    @ViewChild('filterInput', {static: false}) filterInputRef: ElementRef;

    selectedLabels: Map<number, boolean> = new Map<number, boolean>();

    changes: EventEmitter<any> = new EventEmitter<any>(false);

    labelFilter = '';

    ngOnInit(): void {
        fromEvent(this.filterInputRef.nativeElement, 'keyup')
        .pipe(debounceTime(500))
        .subscribe(() => {
            let hnd = setInterval(() => this.cdr.markForCheck(), 100);
            setTimeout(() => clearInterval(hnd), 2000);
        });
    }
    constructor(private cdr: ChangeDetectorRef) {}

    get filteredLabels() {
        return this.labels.filter(label => label.name.includes(this.labelFilter));
    }

    isActive(): boolean {
        return this.selectedLabels.size > 0;
     }

    accepts(cv: any): boolean {
        if (this.resourceType === ResourceType.CHART_VERSION) {
            return (cv as HelmChartVersion).labels.some(label => this.selectedLabels.get(label.id));
        } else if (this.resourceType === ResourceType.REPOSITORY_TAG) {
            return (cv as Tag).labels.some(label => this.selectedLabels.get(label.id));
        } else {
            return true;
        }
    }

    selectLabel(label: Label) {
        this.selectedLabels.set(label.id, true);
        this.changes.emit();
    }

    unselectLabel(label: Label) {
        this.selectedLabels.delete(label.id);
        this.changes.emit(true);
    }

    isSelected(label: Label) {
        return this.selectedLabels.has(label.id);
    }
}
