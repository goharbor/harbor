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
import { Directive, OnChanges, Input, SimpleChanges } from '@angular/core';
import {
    NG_VALIDATORS,
    Validator,
    Validators,
    ValidatorFn,
    AbstractControl,
} from '@angular/forms';

@Directive({
    selector: '[dateValidator]',
    providers: [
        {
            provide: NG_VALIDATORS,
            useExisting: DateValidatorDirective,
            multi: true,
        },
    ],
})
export class DateValidatorDirective implements Validator, OnChanges {
    @Input() dateValidator: string;
    private valFn = Validators.nullValidator;

    ngOnChanges(changes: SimpleChanges): void {
        const change = changes['dateValidator'];
        if (change) {
            this.valFn = dateValidator();
        } else {
            this.valFn = Validators.nullValidator;
        }
    }
    validate(control: AbstractControl): { [key: string]: any } {
        return this.valFn(control);
    }
}

export function dateValidator(): ValidatorFn {
    return (control: AbstractControl): { [key: string]: any } => {
        let controlValue = control.value;
        let valid = true;
        if (controlValue) {
            const regYMD =
                /^(19|20)\d\d([- /.])(0[1-9]|1[012])\2(0[1-9]|[12][0-9]|3[01])$/g;
            const regDMY =
                /^(0[1-9]|[12][0-9]|3[01])[- /.](0[1-9]|1[012])[- /.](19|20)\d\d$/g;
            valid = regYMD.test(controlValue) || regDMY.test(controlValue);
        }
        return valid ? null : { dateValidator: { value: controlValue } };
    };
}
