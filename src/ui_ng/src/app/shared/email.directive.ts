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