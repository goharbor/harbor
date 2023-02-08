import {
    Component,
    ElementRef,
    EventEmitter,
    Input,
    OnDestroy,
    OnInit,
    Output,
    Renderer2,
    ViewChild,
} from '@angular/core';
import { ArtifactFilterEvent, multipleFilter } from '../../../../artifact';
import { Label } from '../../../../../../../../../../ng-swagger-gen/models/label';
import { Subject, Subscription } from 'rxjs';
import { debounceTime } from 'rxjs/operators';

@Component({
    selector: 'app-artifact-filter',
    templateUrl: './artifact-filter.component.html',
    styleUrls: ['./artifact-filter.component.scss'],
})
export class ArtifactFilterComponent implements OnInit, OnDestroy {
    @Input()
    withDivider: boolean = false;
    @ViewChild('filterArea')
    filterArea: ElementRef;
    @Input()
    projectId: number;
    opened: boolean = false;
    multipleFilter = multipleFilter;
    filterByType: string = multipleFilter[0].filterBy;
    dropdownOpened: boolean = true;
    selectedValue: string | Label;
    @Output()
    filterEvent = new EventEmitter<ArtifactFilterEvent>();
    readonly searchId: string = 'search-btn';
    readonly typeSelectId: string = 'type-select';
    inputTag: string;
    private _keyupEventSubject: Subject<string> = new Subject();
    private _keyupEventSubscription: Subscription;
    constructor(private renderer: Renderer2) {
        // click  outside, then close dropdown
        this.renderer.listen('window', 'click', (e: Event) => {
            if (
                !(
                    (e.target as any).id === this.searchId ||
                    (e.target as any).id === this.typeSelectId ||
                    this.filterArea.nativeElement.contains(e.target)
                )
            ) {
                this.dropdownOpened = false;
            }
        });
    }
    ngOnInit(): void {
        if (!this._keyupEventSubscription) {
            this._keyupEventSubscription = this._keyupEventSubject
                .pipe(debounceTime(500))
                .subscribe(inputTag => {
                    this.filterEvent.emit({
                        type: this.filterByType,
                        isLabel: false,
                        isInputTag: true,
                        stringValue: inputTag,
                    });
                });
        }
    }
    ngOnDestroy(): void {
        if (this._keyupEventSubscription) {
            this._keyupEventSubscription.unsubscribe();
            this._keyupEventSubscription = null;
        }
    }

    selectFilterType() {
        this.selectedValue = null;
        this.dropdownOpened = true;
        if (this.filterByType === this.multipleFilter[2].filterBy) {
            this.filterEvent.emit({ type: this.filterByType, isLabel: true });
        } else {
            this.filterEvent.emit({ type: this.filterByType, isLabel: false });
        }
    }

    selectValue(value: any) {
        if (this.filterByType === this.multipleFilter[2].filterBy) {
            // for labels
            if (value.isAdd) {
                this.selectedValue = value.label;
            } else {
                this.selectedValue = null;
            }
            this.filterEvent.emit({
                type: this.filterByType,
                isLabel: true,
                label: this.selectedValue as Label,
            });
        } else {
            this.selectedValue = value?.filterText;
            this.filterEvent.emit({
                type: this.filterByType,
                isLabel: false,
                stringValue: this.selectedValue as string,
            });
        }
    }
    getSelectLabel(): Label[] {
        if (
            this.filterByType === this.multipleFilter[2].filterBy &&
            this.selectedValue
        ) {
            return [this.selectedValue as Label];
        }
        return [];
    }
    searchByInputTag() {
        this._keyupEventSubject.next(this.inputTag);
    }
}
