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
