import {
  Component,
  Input,
  Output,
  EventEmitter,
  ViewChild,
  OnChanges
} from "@angular/core";
import { NgModel } from "@angular/forms";

@Component({
  selector: "hbr-datetime",
  templateUrl: "./datetime-picker.component.html",
  styleUrls: ["./datetime-picker.component.scss"]
})
export class DatePickerComponent implements OnChanges {
  @Input() dateInput: string;
  @Input() oneDayOffset: boolean;

  @ViewChild("searchTime", {static: true}) searchTime: NgModel;

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
    let searchTerm: string = "";
    if (this.searchTime.valid && this.dateInput) {
      searchTerm = this.searchTime.value;
    }
    this.search.emit(searchTerm);
  }
}
