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
    AfterViewInit,
    Component,
    ElementRef,
    OnDestroy,
    ViewChild,
} from '@angular/core';
import {
    UntypedFormBuilder,
    UntypedFormGroup,
    Validators,
} from '@angular/forms';
import { fromEvent } from 'rxjs';
import {
    debounceTime,
    distinctUntilChanged,
    filter,
    finalize,
    map,
    switchMap,
} from 'rxjs/operators';
import { OptimizerService } from '../../../../../../../ng-swagger-gen/services/optimizer.service';

@Component({
    selector: 'new-optimizer-form',
    templateUrl: 'new-optimizer-form.component.html',
    styleUrls: ['new-optimizer-form.component.scss'],
    standalone: false,
})
export class NewOptimizerFormComponent implements AfterViewInit, OnDestroy {
    checkOnGoing: boolean = false;
    newOptimizerForm: UntypedFormGroup = this.fb.group({
        name: this.fb.control('', [Validators.required]),
        description: this.fb.control(''),
        url: this.fb.control('', [
            Validators.required,
            Validators.pattern(/^http[s]?:\/\//),
        ]),
        auth: this.fb.control(''),
        accessCredential: this.fb.group({
            username: this.fb.control('', Validators.required),
            password: this.fb.control('', Validators.required),
            token: this.fb.control('', Validators.required),
            apiKey: this.fb.control('', Validators.required),
        }),
        skipCertVerify: this.fb.control(false),
        useInner: this.fb.control(false),
    });
    checkNameSubscribe: any;
    checkEndpointUrlSubscribe: any;
    nameTooltip: string;
    endpointTooltip: string;
    isNameExisting: boolean = false;
    checkEndpointOnGoing: boolean = false;
    isEndpointUrlExisting: boolean = false;
    showEndpointError: boolean = false;
    originValue: any;
    isEdit: boolean;
    @ViewChild('name') optimizerName: ElementRef;
    @ViewChild('endpointUrl') optimizerEndpointUrl: ElementRef;
    constructor(
        private fb: UntypedFormBuilder,
        private optimizerService: OptimizerService
    ) {}
    ngAfterViewInit(): void {
        if (!this.checkNameSubscribe) {
            this.checkNameSubscribe = fromEvent(
                this.optimizerName.nativeElement,
                'input'
            )
                .pipe(
                    map((e: any) => e.target.value),
                    debounceTime(500),
                    distinctUntilChanged(),
                    filter(name => {
                        if (
                            this.isEdit &&
                            this.originValue &&
                            this.originValue.name === name
                        ) {
                            return false;
                        }
                        return (
                            this.newOptimizerForm.get('name').valid &&
                            name.length > 0
                        );
                    }),
                    switchMap(name => {
                        this.isNameExisting = false;
                        this.checkOnGoing = true;
                        return this.optimizerService
                            .listOptimizers({
                                q: encodeURIComponent(`name=${name}`),
                            })
                            .pipe(finalize(() => (this.checkOnGoing = false)));
                    })
                )
                .subscribe(
                    response => {
                        if (response && response.length > 0) {
                            response.forEach(s => {
                                if (
                                    s.name ===
                                    this.newOptimizerForm.get('name').value
                                ) {
                                    this.isNameExisting = true;
                                    return;
                                }
                            });
                        }
                    },
                    error => {
                        this.isNameExisting = false;
                    }
                );
        }
        if (!this.checkEndpointUrlSubscribe) {
            this.checkEndpointUrlSubscribe = fromEvent(
                this.optimizerEndpointUrl.nativeElement,
                'input'
            )
                .pipe(
                    map((e: any) => e.target.value),
                    debounceTime(800),
                    distinctUntilChanged(),
                    filter(endpointUrl => {
                        if (
                            this.isEdit &&
                            this.originValue &&
                            this.originValue.url === endpointUrl
                        ) {
                            return false;
                        }
                        return (
                            this.newOptimizerForm.get('url').valid &&
                            endpointUrl.length > 6
                        );
                    }),
                    switchMap(endpointUrl => {
                        this.isEndpointUrlExisting = false;
                        this.checkEndpointOnGoing = true;
                        return this.optimizerService
                            .listOptimizers({
                                q: encodeURIComponent(`url=${endpointUrl}`),
                            })
                            .pipe(
                                finalize(
                                    () => (this.checkEndpointOnGoing = false)
                                )
                            );
                    })
                )
                .subscribe(
                    response => {
                        if (response && response.length > 0) {
                            response.forEach(s => {
                                if (
                                    s.url ===
                                    this.newOptimizerForm.get('url').value
                                ) {
                                    this.isEndpointUrlExisting = true;
                                    return;
                                }
                            });
                        }
                    },
                    error => {
                        this.isEndpointUrlExisting = false;
                    }
                );
        }
    }
    ngOnDestroy() {
        if (this.checkNameSubscribe) {
            this.checkNameSubscribe.unsubscribe();
            this.checkNameSubscribe = null;
        }
        if (this.checkEndpointUrlSubscribe) {
            this.checkEndpointUrlSubscribe.unsubscribe();
            this.checkEndpointUrlSubscribe = null;
        }
    }
    get isNameValid(): boolean {
        if (
            !(
                this.newOptimizerForm.get('name').dirty ||
                this.newOptimizerForm.get('name').touched
            )
        ) {
            return true;
        }
        if (this.checkOnGoing) {
            return true;
        }
        if (this.isNameExisting) {
            this.nameTooltip = 'OPTIMIZER.NAME_EXISTS';
            return false;
        }
        if (
            this.newOptimizerForm.get('name').errors &&
            this.newOptimizerForm.get('name').errors.required
        ) {
            this.nameTooltip = 'OPTIMIZER.NAME_REQUIRED';
            return false;
        }
        if (
            this.newOptimizerForm.get('name').errors &&
            this.newOptimizerForm.get('name').errors.pattern
        ) {
            this.nameTooltip = 'OPTIMIZER.NAME_REX';
            return false;
        }
        return true;
    }
    get isEndpointValid(): boolean {
        if (
            !(
                this.newOptimizerForm.get('url').dirty ||
                this.newOptimizerForm.get('url').touched
            )
        ) {
            return true;
        }
        if (this.checkEndpointOnGoing) {
            return true;
        }
        if (this.isEndpointUrlExisting) {
            this.endpointTooltip = 'OPTIMIZER.ENDPOINT_EXISTS';
            return false;
        }
        if (
            this.newOptimizerForm.get('url').errors &&
            this.newOptimizerForm.get('url').errors.required
        ) {
            this.endpointTooltip = 'OPTIMIZER.ENDPOINT_REQUIRED';
            return false;
        }
        //  skip here, validate when onblur
        if (
            this.newOptimizerForm.get('url').errors &&
            this.newOptimizerForm.get('url').errors.pattern
        ) {
            return true;
        }
        return true;
    }
    //  validate endpointUrl when onblur
    checkEndpointUrl() {
        if (
            this.newOptimizerForm.get('url').errors &&
            this.newOptimizerForm.get('url').errors.pattern
        ) {
            this.endpointTooltip = 'OPTIMIZER.ILLEGAL_ENDPOINT';
            this.showEndpointError = true;
        }
    }
    get auth(): string {
        return this.newOptimizerForm.get('auth').value;
    }
    get isUserNameValid(): boolean {
        return !(
            this.newOptimizerForm.get('accessCredential').get('username')
                .invalid &&
            (this.newOptimizerForm.get('accessCredential').get('username')
                .dirty ||
                this.newOptimizerForm.get('accessCredential').get('username')
                    .touched)
        );
    }
    get isPasswordValid(): boolean {
        return !(
            this.newOptimizerForm.get('accessCredential').get('password')
                .invalid &&
            (this.newOptimizerForm.get('accessCredential').get('password')
                .dirty ||
                this.newOptimizerForm.get('accessCredential').get('password')
                    .touched)
        );
    }
    get isTokenValid(): boolean {
        return !(
            this.newOptimizerForm.get('accessCredential').get('token').invalid &&
            (this.newOptimizerForm.get('accessCredential').get('token').dirty ||
                this.newOptimizerForm.get('accessCredential').get('token')
                    .touched)
        );
    }
    get isApiKeyValid(): boolean {
        return !(
            this.newOptimizerForm.get('accessCredential').get('apiKey').invalid &&
            (this.newOptimizerForm.get('accessCredential').get('apiKey').dirty ||
                this.newOptimizerForm.get('accessCredential').get('apiKey')
                    .touched)
        );
    }
}
