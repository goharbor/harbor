import { Directive } from '@angular/core';
import { ValidatorFn, AbstractControl, Validator, NG_VALIDATORS, Validators } from '@angular/forms';

export const portNumbers = /[\d]+/;

export function portValidator(): ValidatorFn {
    return (control: AbstractControl): { [key: string]: any } => {
        const value: string = control.value
        if (!value) {
            return { 'port': 65535 };
        }

        const regExp = new RegExp(portNumbers, 'i');
        if(!regExp.test(value)){
            return { 'port': 65535 };
        }else{
            const portV = parseInt(value);
            if(portV <=0 || portV >65535){
                return { 'port': 65535 };
            }
        }
        return null;
    }
}

@Directive({
    selector: '[port]',
    providers: [{ provide: NG_VALIDATORS, useExisting: PortValidatorDirective, multi: true }]
})

export class PortValidatorDirective implements Validator {
    private valFn = portValidator();

    validate(control: AbstractControl): { [key: string]: any } {
        return this.valFn(control);
    }
}