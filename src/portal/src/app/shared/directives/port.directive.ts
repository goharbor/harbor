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
import { Directive } from '@angular/core';
import {
    ValidatorFn,
    AbstractControl,
    Validator,
    NG_VALIDATORS,
} from '@angular/forms';

export const portNumbers = /^[\d]{1,5}$/;

export function portValidator(): ValidatorFn {
    return (control: AbstractControl): { [key: string]: any } => {
        const value: string = control.value;
        if (!value) {
            return { port: 65535 };
        }

        const regExp = new RegExp(portNumbers, 'i');
        if (!regExp.test(value)) {
            return { port: 65535 };
        } else {
            const portV = Number.parseInt(value, 10);
            if (portV <= 0 || portV > 65535) {
                return { port: 65535 };
            }
        }
        return null;
    };
}

@Directive({
    selector: '[port]',
    providers: [
        {
            provide: NG_VALIDATORS,
            useExisting: PortValidatorDirective,
            multi: true,
        },
    ],
})
export class PortValidatorDirective implements Validator {
    valFn = portValidator();

    validate(control: AbstractControl): { [key: string]: any } {
        return this.valFn(control);
    }
}
