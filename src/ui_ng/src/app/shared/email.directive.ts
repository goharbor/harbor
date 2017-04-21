// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { ValidatorFn, AbstractControl, Validator, NG_VALIDATORS, Validators } from '@angular/forms';

const emailPattern = /^(([^<>()[\]\.,;:\s@\"]+(\.[^<>()[\]\.,;:\s@\"]+)*)|(\".+\"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

export function emailValidator(): ValidatorFn {
    return (control: AbstractControl): { [key: string]: any } => {
        const value: string = control.value
        if (!value) {
            return { 'email': false };
        }

        const regExp = new RegExp(emailPattern);
        if(!regExp.test(value)){
            return { 'email': false };
        }
        
        return null;
    }
}

@Directive({
    selector: '[email]',
    providers: [{ provide: NG_VALIDATORS, useExisting: EmailValidatorDirective, multi: true }]
})

export class EmailValidatorDirective implements Validator {
    valFn = emailValidator();

    validate(control: AbstractControl): { [key: string]: any } {
        return this.valFn(control);
    }
}