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
    ValidatorFn,
    AbstractControl,
    Validator,
    NG_VALIDATORS,
    Validators,
} from '@angular/forms';

export const assiiChars = /[\u4e00-\u9fa5]/;

export function maxLengthExtValidator(length: number): ValidatorFn {
    return (control: AbstractControl): { [key: string]: any } => {
        const value: string = control.value;
        if (!value || value.trim() === '') {
            return null;
        }

        const regExp = new RegExp(assiiChars, 'i');
        let count = 0;
        let len = value.length;

        for (let i = 0; i < len; i++) {
            if (regExp.test(value[i])) {
                count += 3;
            } else {
                count++;
            }
        }
        return count > length ? { maxLengthExt: count } : null;
    };
}

@Directive({
    selector: '[maxLengthExt]',
    providers: [
        {
            provide: NG_VALIDATORS,
            useExisting: MaxLengthExtValidatorDirective,
            multi: true,
        },
    ],
})
export class MaxLengthExtValidatorDirective implements Validator, OnChanges {
    @Input() maxLengthExt: number;
    valFn = Validators.nullValidator;

    ngOnChanges(changes: SimpleChanges): void {
        const change = changes['maxLengthExt'];
        if (change) {
            const val: number = change.currentValue;
            this.valFn = maxLengthExtValidator(val);
        } else {
            this.valFn = Validators.nullValidator;
        }
    }

    validate(control: AbstractControl): { [key: string]: any } {
        return this.valFn(control);
    }
}
