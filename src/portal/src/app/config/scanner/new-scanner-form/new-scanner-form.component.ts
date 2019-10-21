import {
    AfterViewInit,
    Component,
    ElementRef,
    OnDestroy,
    OnInit,
    ViewChild
} from "@angular/core";
import { FormBuilder, FormGroup, Validators } from "@angular/forms";
import { fromEvent } from "rxjs";
import { debounceTime, distinctUntilChanged, filter, finalize, map, switchMap } from "rxjs/operators";
import { ConfigScannerService } from "../config-scanner.service";


@Component({
    selector: 'new-scanner-form',
    templateUrl: 'new-scanner-form.component.html',
    styleUrls: ['new-scanner-form.component.scss']
})
export class NewScannerFormComponent implements  OnInit, AfterViewInit, OnDestroy {
    checkOnGoing: boolean = false;
    newScannerForm: FormGroup = this.fb.group({
        name: this.fb.control("",
            [Validators.required]),
        description: this.fb.control(""),
        url: this.fb.control("",
            [Validators.required,
                Validators.pattern(/^http[s]?:\/\//)]),
        auth: this.fb.control(""),
        accessCredential: this.fb.group({
                username: this.fb.control("", Validators.required),
                password: this.fb.control("", Validators.required),
                token: this.fb.control("", Validators.required),
                apiKey: this.fb.control("", Validators.required)
        }),
        skipCertVerify: this.fb.control(false)
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
    @ViewChild('name', {static: false}) scannerName: ElementRef;
    @ViewChild('endpointUrl', {static: false}) scannerEndpointUrl: ElementRef;
    constructor(private fb: FormBuilder, private scannerService: ConfigScannerService) {
    }
    ngAfterViewInit(): void {
        if (!this.checkNameSubscribe) {
            this.checkNameSubscribe = fromEvent(this.scannerName.nativeElement, 'input').pipe(
                map((e: any) => e.target.value),
                filter(name => {
                    if (this.isEdit && this.originValue && this.originValue.name === name) {
                        return false;
                    }
                    return this.newScannerForm.get('name').valid && name.length > 0;
                }),
                debounceTime(500),
                distinctUntilChanged(),
                switchMap((name) => {
                    this.isNameExisting = false;
                    this.checkOnGoing = true;
                    return  this.scannerService.getScannersByName(name)
                        .pipe(finalize(() => this.checkOnGoing = false));
                })).subscribe(response => {
                if (response && response.length > 0) {
                    response.forEach(s => {
                        if (s.name === this.newScannerForm.get('name').value) {
                            this.isNameExisting = true;
                            return;
                        }
                    });
                }
            }, error => {
                this.isNameExisting = false;
            });
        }
        if (!this.checkEndpointUrlSubscribe) {
            this.checkEndpointUrlSubscribe = fromEvent(this.scannerEndpointUrl.nativeElement, 'input').pipe(
                map((e: any) => e.target.value),
                filter(endpointUrl => {
                    if (this.isEdit && this.originValue && this.originValue.url === endpointUrl) {
                        return false;
                    }
                    return this.newScannerForm.get('url').valid && endpointUrl.length > 6;
                }),
                debounceTime(800),
                distinctUntilChanged(),
                switchMap((endpointUrl) => {
                    this.isEndpointUrlExisting = false;
                    this.checkEndpointOnGoing = true;
                    return  this.scannerService.getScannersByEndpointUrl(endpointUrl)
                        .pipe(finalize(() => this.checkEndpointOnGoing = false));
                })).subscribe(response => {
                if (response && response.length > 0) {
                    response.forEach(s => {
                        if (s.url === this.newScannerForm.get('url').value) {
                            this.isEndpointUrlExisting = true;
                            return;
                        }
                    });
                }
            }, error => {
                this.isEndpointUrlExisting = false;
            });
        }
    }

    ngOnInit(): void {
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
        if (!(this.newScannerForm.get('name').dirty || this.newScannerForm.get('name').touched)) {
            return true;
        }
        if (this.checkOnGoing) {
            return true;
        }
        if (this.isNameExisting) {
            this.nameTooltip = 'SCANNER.NAME_EXISTS';
            return false;
        }
        if (this.newScannerForm.get('name').errors && this.newScannerForm.get('name').errors.required) {
            this.nameTooltip = 'SCANNER.NAME_REQUIRED';
            return false;
        }
        if (this.newScannerForm.get('name').errors && this.newScannerForm.get('name').errors.pattern) {
            this.nameTooltip = 'SCANNER.NAME_REX';
            return false;
        }
        return true;
    }
    get isEndpointValid(): boolean {
        if (!(this.newScannerForm.get('url').dirty || this.newScannerForm.get('url').touched)) {
            return true;
        }
        if (this.checkEndpointOnGoing) {
            return true;
        }
        if (this.isEndpointUrlExisting) {
            this.endpointTooltip = 'SCANNER.ENDPOINT_EXISTS';
            return false;
        }
        if (this.newScannerForm.get('url').errors && this.newScannerForm.get('url').errors.required) {
            this.endpointTooltip = 'SCANNER.ENDPOINT_REQUIRED';
            return false;
        }
        //  skip here, validate when onblur
        if (this.newScannerForm.get('url').errors && this.newScannerForm.get('url').errors.pattern) {
           return true;
        }
        return true;
    }
    //  validate endpointUrl when onblur
    checkEndpointUrl() {
        if (this.newScannerForm.get('url').errors && this.newScannerForm.get('url').errors.pattern) {
            this.endpointTooltip = "SCANNER.ILLEGAL_ENDPOINT";
            this.showEndpointError = true;
        }
    }
    get auth(): string {
        return this.newScannerForm.get('auth').value;
    }
    get isUserNameValid(): boolean {
        return !(this.newScannerForm.get('accessCredential').get('username').invalid
               && (this.newScannerForm.get('accessCredential').get('username').dirty
                || this.newScannerForm.get('accessCredential').get('username').touched));
    }
    get isPasswordValid(): boolean {
        return !(this.newScannerForm.get('accessCredential').get('password').invalid
            && (this.newScannerForm.get('accessCredential').get('password').dirty
                || this.newScannerForm.get('accessCredential').get('password').touched));
    }
    get isTokenValid(): boolean {
        return !(this.newScannerForm.get('accessCredential').get('token').invalid
            && (this.newScannerForm.get('accessCredential').get('token').dirty
                || this.newScannerForm.get('accessCredential').get('token').touched));
    }
    get isApiKeyValid(): boolean {
        return !(this.newScannerForm.get('accessCredential').get('apiKey').invalid
            && (this.newScannerForm.get('accessCredential').get('apiKey').dirty
                || this.newScannerForm.get('accessCredential').get('apiKey').touched));
    }
}
