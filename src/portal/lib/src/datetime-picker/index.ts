import { Type } from "@angular/core";

import { DatePickerComponent } from "./datetime-picker.component";
import { DateValidatorDirective } from "./date-validator.directive";
export const DATETIME_PICKER_DIRECTIVES: Type<any>[] = [
  DatePickerComponent,
  DateValidatorDirective
];
