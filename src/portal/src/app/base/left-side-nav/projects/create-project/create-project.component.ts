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
import {
    debounceTime,
    distinctUntilChanged,
    filter,
    map,
    switchMap,
} from 'rxjs/operators';
import {
    Component,
    EventEmitter,
    Output,
    ViewChild,
    OnInit,
    OnDestroy,
    Input,
    OnChanges,
    SimpleChanges,
    AfterViewInit,
    ElementRef,
} from '@angular/core';
import { NgForm, Validators } from '@angular/forms';
import { forkJoin, fromEvent, Observable, Subscription } from 'rxjs';
import { TranslateService } from '@ngx-translate/core';
import { MessageHandlerService } from '../../../../shared/services/message-handler.service';
import { Project } from '../../../project/project';
import {
    QuotaUnits,
    QuotaUnlimited,
} from '../../../../shared/entities/shared.const';
import { QuotaHardInterface } from '../../../../shared/services';
import {
    clone,
    getByte,
    GetIntegerAndUnit,
    validateLimit,
} from '../../../../shared/units/utils';
import { InlineAlertComponent } from '../../../../shared/components/inline-alert/inline-alert.component';
import { Registry } from '../../../../../../ng-swagger-gen/models/registry';
import { RegistryService } from '../../../../../../ng-swagger-gen/services/registry.service';
import { ProjectService } from '../../../../../../ng-swagger-gen/services/project.service';

const PAGE_SIZE: number = 100;
@Component({
    selector: 'create-project',
    templateUrl: 'create-project.component.html',
    styleUrls: ['create-project.scss'],
})
export class CreateProjectComponent
    implements OnInit, AfterViewInit, OnChanges, OnDestroy
{
    projectForm: NgForm;

    @ViewChild('projectForm', { static: true })
    currentForm: NgForm;
    quotaUnits = QuotaUnits;
    project: Project = new Project();
    storageLimit: number;
    storageLimitUnit: string = QuotaUnits[3].UNIT;
    storageDefaultLimit: number;
    storageDefaultLimitUnit: string;
    initVal: Project = new Project();

    createProjectOpened: boolean;

    hasChanged: boolean;
    isSubmitOnGoing = false;

    staticBackdrop = true;
    closable = false;
    isNameExisted: boolean = false;
    nameTooltipText = 'PROJECT.NAME_TOOLTIP';
    checkOnGoing = false;
    enableProxyCache: boolean = false;
    endpoint: string = '';
    @Output() create = new EventEmitter<boolean>();
    @Input() quotaObj: QuotaHardInterface;
    @Input() isSystemAdmin: boolean;
    @ViewChild(InlineAlertComponent, { static: true })
    inlineAlert: InlineAlertComponent;
    @ViewChild('projectName') projectNameInput: ElementRef;
    checkNameSubscribe: Subscription;

    registries: Registry[] = [];
    supportedRegistryTypeQueryString: string =
        'type={docker-hub harbor azure-acr aws-ecr google-gcr quay docker-registry}';

    constructor(
        private projectService: ProjectService,
        private translateService: TranslateService,
        private messageHandlerService: MessageHandlerService,
        private endpointService: RegistryService
    ) {}

    ngOnInit(): void {
        if (this.isSystemAdmin) {
            this.getRegistries();
        }
    }

    getRegistries() {
        this.endpointService
            .listRegistriesResponse({
                page: 1,
                pageSize: PAGE_SIZE,
                q: this.supportedRegistryTypeQueryString,
            })
            .subscribe(
                result => {
                    // Get total count
                    if (result.headers) {
                        const xHeader: string =
                            result.headers.get('X-Total-Count');
                        const totalCount = parseInt(xHeader, 0);
                        let arr = result.body || [];
                        if (totalCount <= PAGE_SIZE) {
                            // already gotten all Registries
                            this.registries = result.body || [];
                        } else {
                            // get all the registries in specified times
                            const times: number = Math.ceil(
                                totalCount / PAGE_SIZE
                            );
                            const observableList: Observable<Registry[]>[] = [];
                            for (let i = 2; i <= times; i++) {
                                observableList.push(
                                    this.endpointService.listRegistries({
                                        page: i,
                                        pageSize: PAGE_SIZE,
                                        q: this
                                            .supportedRegistryTypeQueryString,
                                    })
                                );
                            }
                            forkJoin(observableList).subscribe(res => {
                                if (res && res.length) {
                                    res.forEach(item => {
                                        arr = arr.concat(item);
                                    });
                                    this.registries = arr;
                                }
                            });
                        }
                    }
                },
                error => {
                    this.messageHandlerService.error(error);
                }
            );
    }

    ngAfterViewInit(): void {
        if (!this.checkNameSubscribe) {
            this.checkNameSubscribe = fromEvent(
                this.projectNameInput.nativeElement,
                'input'
            )
                .pipe(
                    map((e: any) => e.target.value),
                    debounceTime(300),
                    distinctUntilChanged(),
                    filter(name => {
                        return (
                            this.currentForm.controls['create_project_name']
                                .valid && name.length > 0
                        );
                    }),
                    switchMap(name => {
                        // Check exiting from backend
                        this.checkOnGoing = true;
                        this.isNameExisted = false;
                        return this.projectService.listProjects({
                            q: encodeURIComponent(`name=${name}`),
                        });
                    })
                )
                .subscribe(
                    response => {
                        // Project existing
                        if (response && response.length) {
                            this.isNameExisted = true;
                        }
                        this.checkOnGoing = false;
                    },
                    error => {
                        this.checkOnGoing = false;
                        this.isNameExisted = false;
                    }
                );
        }
    }
    get isNameValid(): boolean {
        if (
            !this.currentForm ||
            !this.currentForm.controls ||
            !this.currentForm.controls['create_project_name']
        ) {
            return true;
        }
        if (
            !(
                this.currentForm.controls['create_project_name'].dirty ||
                this.currentForm.controls['create_project_name'].touched
            )
        ) {
            return true;
        }
        if (this.checkOnGoing) {
            return true;
        }
        if (this.currentForm.controls['create_project_name'].errors) {
            this.nameTooltipText = 'PROJECT.NAME_TOOLTIP';
            return false;
        }
        if (this.isNameExisted) {
            this.nameTooltipText = 'PROJECT.NAME_ALREADY_EXISTS';
            return false;
        }
        return true;
    }
    ngOnChanges(changes: SimpleChanges): void {
        if (
            changes &&
            changes['quotaObj'] &&
            changes['quotaObj'].currentValue
        ) {
            this.storageLimit = GetIntegerAndUnit(
                this.quotaObj.storage_per_project,
                clone(QuotaUnits),
                0,
                clone(QuotaUnits)
            ).partNumberHard;
            this.storageLimitUnit =
                this.storageLimit === QuotaUnlimited
                    ? QuotaUnits[3].UNIT
                    : GetIntegerAndUnit(
                          this.quotaObj.storage_per_project,
                          clone(QuotaUnits),
                          0,
                          clone(QuotaUnits)
                      ).partCharacterHard;

            this.storageDefaultLimit = this.storageLimit;
            this.storageDefaultLimitUnit = this.storageLimitUnit;
            if (this.isSystemAdmin) {
                this.currentForm.form.controls[
                    'create_project_storage_limit'
                ].setValidators([
                    Validators.required,
                    Validators.pattern('(^-1$)|(^([1-9]+)([0-9]+)*$)'),
                    validateLimit(
                        this.currentForm.form.controls[
                            'create_project_storage_limit_unit'
                        ]
                    ),
                ]);
            }
            this.currentForm.form.valueChanges
                .pipe(
                    distinctUntilChanged(
                        (a, b) => JSON.stringify(a) === JSON.stringify(b)
                    )
                )
                .subscribe(data => {
                    [
                        'create_project_storage_limit',
                        'create_project_storage_limit_unit',
                    ].forEach(fieldName => {
                        if (
                            this.currentForm.form.get(fieldName) &&
                            this.currentForm.form.get(fieldName).value !== null
                        ) {
                            this.currentForm.form
                                .get(fieldName)
                                .updateValueAndValidity();
                        }
                    });
                });
        }
    }
    ngOnDestroy(): void {
        if (this.checkNameSubscribe) {
            this.checkNameSubscribe.unsubscribe();
            this.checkNameSubscribe = null;
        }
    }

    onSubmit() {
        if (this.isSubmitOnGoing) {
            return;
        }
        this.isSubmitOnGoing = true;
        const storageByte =
            +this.storageLimit === QuotaUnlimited
                ? this.storageLimit
                : getByte(+this.storageLimit, this.storageLimitUnit);
        this.projectService
            .createProject({
                project: {
                    project_name: this.project.name,
                    metadata: {
                        public: this.project.metadata.public ? 'true' : 'false',
                    },
                    storage_limit: +storageByte,
                    registry_id: +this.project.registry_id,
                },
            })
            .subscribe(
                status => {
                    this.isSubmitOnGoing = false;

                    this.create.emit(true);
                    this.messageHandlerService.showSuccess(
                        'PROJECT.CREATED_SUCCESS'
                    );
                    this.createProjectOpened = false;
                },
                error => {
                    this.isSubmitOnGoing = false;
                    this.inlineAlert.showInlineError(error);
                }
            );
    }

    onCancel() {
        this.createProjectOpened = false;
    }

    newProject() {
        this.project = new Project();
        this.hasChanged = false;
        this.createProjectOpened = true;
        this.enableProxyCache = false;
        this.endpoint = '';
        if (
            this.currentForm &&
            this.currentForm.controls &&
            this.currentForm.controls['create_project_name']
        ) {
            this.currentForm.controls['create_project_name'].reset();
        }
        this.inlineAlert.close();
        this.storageLimit = this.storageDefaultLimit;
        this.storageLimitUnit = this.storageDefaultLimitUnit;
    }

    public get isValid(): boolean {
        return (
            this.currentForm &&
            this.currentForm.valid &&
            !this.isSubmitOnGoing &&
            this.isNameValid &&
            !this.checkOnGoing
        );
    }

    getEndpoint(): string {
        if (
            this.registries &&
            this.registries.length &&
            this.project.registry_id
        ) {
            for (let i = 0; i < this.registries.length; i++) {
                if (+this.registries[i].id === +this.project.registry_id) {
                    return this.registries[i].url;
                }
            }
        }
        return '';
    }
}
