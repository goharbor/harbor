import {
    Component,
    Input,
    Output,
    EventEmitter,
    ViewChild,
    OnChanges,
    LOCALE_ID,
} from '@angular/core';
import { NgModel } from '@angular/forms';
import { DEFAULT_LANG_LOCALSTORAGE_KEY } from '../../entities/shared.const';

@Component({
    selector: 'hbr-datetime',
    templateUrl: './datetime-picker.component.html',
    styleUrls: ['./datetime-picker.component.scss'],
    providers: [
        {
            provide: LOCALE_ID,
            useValue: localStorage.getItem(DEFAULT_LANG_LOCALSTORAGE_KEY),
        },
    ],
})
export class DatePickerComponent implements OnChanges {
    @Input() dateInput: string;
    @Input() oneDayOffset: boolean;

    @ViewChild('searchTime', { static: true }) searchTime: NgModel;

    @Output() search = new EventEmitter<string>();

    ngOnChanges(): void {
        this.dateInput = this.dateInput.trim();
    }

    get dateInvalid(): boolean {
        return (
            (this.searchTime.errors &&
                this.searchTime.errors.dateValidator &&
                (this.searchTime.dirty || this.searchTime.touched)) ||
            false
        );
    }
    doSearch() {
        let searchTerm: string = '';
        if (this.searchTime.valid && this.dateInput) {
            searchTerm = this.searchTime.value;
        }
        this.search.emit(searchTerm);
    }
}
