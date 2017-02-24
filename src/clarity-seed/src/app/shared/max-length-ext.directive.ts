import { Directive, OnChanges, Input, SimpleChanges } from '@angular/core';
import { ValidatorFn, AbstractControl, Validator, NG_VALIDATORS, Validators } from '@angular/forms';

export const assiiChars = /[\u4e00-\u9fa5]/;

export function maxLengthExtValidator(length: number): ValidatorFn {
    return (control: AbstractControl): { [key: string]: any } => {
        const value: string = control.value
        if (!value || value.trim() === "") {
            return { 'maxLengthExt': 0 };
        }

        const regExp = new RegExp(assiiChars, 'i');
        let count = 0;
        let len = value.length;

        for (var i = 0; i < len; i++) {
            if (regExp.test(value[i])) {
                count += 2;
            } else {
                count++;
            }
        }
        return count > length ? { 'maxLengthExt': count } : null;
    }
}

@Directive({
    selector: '[maxLengthExt]',
    providers: [{ provide: NG_VALIDATORS, useExisting: MaxLengthExtValidatorDirective, multi: true }]
})

export class MaxLengthExtValidatorDirective implements Validator, OnChanges {
    @Input() maxLengthExt: number;
    private valFn = Validators.nullValidator;

    ngOnChanges(changes: SimpleChanges): void {
        const change = changes['maxLengthExt'];
        if (change) {
            const val: number = change.currentValue;
            this.valFn = maxLengthExtValidator(val);
        } else {
            this.valFn = Validators.nullValidator;
        }
        console.info(changes, this.maxLengthExt);
    }

    validate(control: AbstractControl): { [key: string]: any } {
        return this.valFn(control);
    }
}