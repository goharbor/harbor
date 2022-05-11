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
    Component,
    OnInit,
    OnDestroy,
    ViewChild,
    Input,
    EventEmitter,
    Output,
} from '@angular/core';
import { Filter } from '../../../../../shared/services';
import { forkJoin, Observable, Subject, Subscription } from 'rxjs';
import { debounceTime, distinctUntilChanged, finalize } from 'rxjs/operators';
import {
    FormArray,
    FormBuilder,
    FormGroup,
    Validators,
    FormControl,
} from '@angular/forms';
import {
    clone,
    isEmptyObject,
    isSameObject,
} from '../../../../../shared/units/utils';
import { InlineAlertComponent } from '../../../../../shared/components/inline-alert/inline-alert.component';
import { ReplicationService } from '../../../../../shared/services';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { TranslateService } from '@ngx-translate/core';
import { cronRegex } from '../../../../../shared/units/utils';
import { FilterType } from '../../../../../shared/entities/shared.const';
import { RegistryService } from '../../../../../../../ng-swagger-gen/services/registry.service';
import { Registry } from '../../../../../../../ng-swagger-gen/models/registry';
import { Label } from '../../../../../../../ng-swagger-gen/models/label';
import { LabelService } from '../../../../../../../ng-swagger-gen/services/label.service';
import {
    BandwidthUnit,
    Decoration,
    Flatten_I18n_MAP,
    Flatten_Level,
} from '../../replication';
import { errorHandler as errorHandlerFn } from '../../../../../shared/units/shared.utils';
import { ReplicationPolicy } from '../../../../../../../ng-swagger-gen/models/replication-policy';

const PREFIX: string = '0 ';
const PAGE_SIZE: number = 100;
export const KB_TO_MB: number = 1024;

@Component({
    selector: 'hbr-create-edit-rule',
    templateUrl: './create-edit-rule.component.html',
    styleUrls: ['./create-edit-rule.component.scss'],
})
export class CreateEditRuleComponent implements OnInit, OnDestroy {
    sourceList: Registry[] = [];
    targetList: Registry[] = [];
    noEndpointInfo = '';
    isPushMode = true;
    noSelectedEndpoint = true;
    TRIGGER_TYPES = {
        MANUAL: 'manual',
        SCHEDULED: 'scheduled',
        EVENT_BASED: 'event_based',
    };

    ruleNameTooltip = 'REPLICATION.NAME_TOOLTIP';
    headerTitle = 'REPLICATION.ADD_POLICY';

    createEditRuleOpened: boolean;
    inProgress = false;
    onGoing = false;
    inNameChecking = false;
    isRuleNameValid = true;
    nameChecker: Subject<string> = new Subject<string>();
    policyId: number;
    confirmSub: Subscription;
    ruleForm: FormGroup;
    copyUpdateForm: ReplicationPolicy;
    cronString: string;
    supportedTriggers: string[];
    supportedFilters: Filter[];
    supportedFilterLabels: {
        name: string;
        color: string;
        select: boolean;
        scope: string;
    }[] = [];

    @Input() withAdmiral: boolean;

    @Output() goToRegistry = new EventEmitter<any>();
    @Output() reload = new EventEmitter<boolean>();

    @ViewChild(InlineAlertComponent, { static: true })
    inlineAlert: InlineAlertComponent;
    flattenLevelMap = Flatten_I18n_MAP;
    speedUnits = [
        {
            UNIT: BandwidthUnit.KB,
        },
        {
            UNIT: BandwidthUnit.MB,
        },
    ];
    selectedUnit: string = BandwidthUnit.KB;
    copySpeedUnit: string = BandwidthUnit.KB;
    constructor(
        private fb: FormBuilder,
        private repService: ReplicationService,
        private endpointService: RegistryService,
        private errorHandler: ErrorHandler,
        private translateService: TranslateService,
        private labelService: LabelService
    ) {
        this.createForm();
    }

    initRegistryInfo(id: number): void {
        this.onGoing = true;
        this.repService
            .getRegistryInfo(id)
            .pipe(finalize(() => (this.onGoing = false)))
            .subscribe(
                adapter => {
                    this.setFilterAndTrigger(adapter);
                },
                (error: any) => {
                    this.inlineAlert.showInlineError(error);
                }
            );
    }
    getAllRegistries() {
        this.endpointService
            .listRegistriesResponse({
                page: 1,
                pageSize: PAGE_SIZE,
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
                            this.targetList = result.body || [];
                            this.sourceList = result.body || [];
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
                                    })
                                );
                            }
                            forkJoin(observableList).subscribe(res => {
                                if (res && res.length) {
                                    res.forEach(item => {
                                        arr = arr.concat(item);
                                    });
                                    this.sourceList = arr;
                                    this.targetList = arr;
                                }
                            });
                        }
                    }
                },
                error => {
                    this.errorHandler.error(error);
                }
            );
    }
    ngOnInit(): void {
        this.getAllLabels();
        this.getAllRegistries();
        this.nameChecker
            .pipe(debounceTime(300))
            .pipe(distinctUntilChanged())
            .subscribe((ruleName: string) => {
                let cont = this.ruleForm.controls['name'];
                if (cont) {
                    this.isRuleNameValid = cont.valid;
                    if (this.isRuleNameValid) {
                        this.inNameChecking = true;
                        this.repService
                            .getReplicationRules(0, ruleName)
                            .subscribe(
                                response => {
                                    if (
                                        response.some(
                                            rule =>
                                                rule.name === ruleName &&
                                                rule.id !== this.policyId
                                        )
                                    ) {
                                        this.ruleNameTooltip =
                                            'TOOLTIP.RULE_USER_EXISTING';
                                        this.isRuleNameValid = false;
                                    }
                                    this.inNameChecking = false;
                                },
                                () => {
                                    this.inNameChecking = false;
                                }
                            );
                    } else {
                        this.ruleNameTooltip = 'REPLICATION.NAME_TOOLTIP';
                    }
                }
            });
    }
    trimText(event) {
        if (event.target.value) {
            event.target.value = event.target.value.replace(/\s+/g, '');
        }
    }
    equals(c1: any, c2: any): boolean {
        return c1 && c2 ? c1.id === c2.id : c1 === c2;
    }
    pushModeChange(): void {
        this.setFilter([]);
        this.initRegistryInfo(0);
    }

    pullModeChange(): void {
        let selectId = this.ruleForm.get('src_registry').value;
        if (selectId) {
            this.setFilter([]);
            this.initRegistryInfo(selectId.id);
        }
    }

    sourceChange($event): void {
        this.noSelectedEndpoint = false;
        let selectId = this.ruleForm.get('src_registry').value;
        this.setFilter([]);
        this.initRegistryInfo(selectId.id);
    }

    ngOnDestroy(): void {
        if (this.confirmSub) {
            this.confirmSub.unsubscribe();
        }
        if (this.nameChecker) {
            this.nameChecker.unsubscribe();
        }
    }

    get isValid() {
        if (this.ruleForm.controls['dest_namespace'].value) {
            if (this.ruleForm.controls['dest_namespace'].invalid) {
                return false;
            }
        }
        if (this.ruleForm.controls['speed'].invalid) {
            return false;
        }
        let controlName = !!this.ruleForm.controls['name'].value;
        let sourceRegistry = !!this.ruleForm.controls['src_registry'].value;
        let destRegistry = !!this.ruleForm.controls['dest_registry'].value;
        let triggerMode = !!this.ruleForm.controls['trigger'].value.type;
        let cron = !!this.ruleForm.value.trigger.trigger_settings.cron;
        return !(
            !controlName ||
            !triggerMode ||
            !this.isRuleNameValid ||
            (!this.isPushMode && !sourceRegistry) ||
            (this.isPushMode && !destRegistry) ||
            !(
                (!this.isNotSchedule() &&
                    cron &&
                    cronRegex(
                        this.ruleForm.value.trigger.trigger_settings.cron || ''
                    )) ||
                this.isNotSchedule()
            )
        );
    }

    createForm() {
        this.ruleForm = this.fb.group({
            name: ['', Validators.required],
            description: '',
            src_registry: new FormControl(),
            dest_registry: new FormControl(),
            dest_namespace: '',
            dest_namespace_replace_count: -1,
            trigger: this.fb.group({
                type: '',
                trigger_settings: this.fb.group({
                    cron: '',
                }),
            }),
            filters: this.fb.array([]),
            enabled: true,
            deletion: false,
            override: true,
            speed: -1,
        });
    }

    isNotSchedule(): boolean {
        return (
            this.ruleForm.get('trigger').get('type').value !==
            this.TRIGGER_TYPES.SCHEDULED
        );
    }

    isNotEventBased(): boolean {
        return (
            this.ruleForm.get('trigger').get('type').value !==
            this.TRIGGER_TYPES.EVENT_BASED
        );
    }

    formReset(): void {
        this.ruleForm.reset({
            name: '',
            description: '',
            trigger: {
                type: '',
                trigger_settings: {
                    cron: '',
                },
            },
            deletion: false,
            enabled: true,
            override: true,
            dest_namespace_replace_count: Flatten_Level.FLATTEN_LEVEl_1,
            speed: -1,
        });
        this.isPushMode = true;
        this.selectedUnit = BandwidthUnit.KB;
    }

    updateRuleFormAndCopyUpdateForm(rule: ReplicationPolicy): void {
        this.isPushMode = rule.dest_registry.id !== 0;
        setTimeout(() => {
            // convert speed unit to KB or MB
            let speed: number = this.convertToInputValue(rule.speed);
            // There is no trigger_setting type when the harbor is upgraded from the old version.
            rule.trigger.trigger_settings = rule.trigger.trigger_settings
                ? rule.trigger.trigger_settings
                : { cron: '' };
            this.ruleForm.reset({
                name: rule.name,
                description: rule.description,
                dest_namespace: rule.dest_namespace,
                dest_namespace_replace_count: rule.dest_namespace_replace_count,
                src_registry: rule.src_registry,
                dest_registry: rule.dest_registry,
                trigger: rule.trigger,
                deletion: rule.deletion,
                enabled: rule.enabled,
                override: rule.override,
                speed: speed,
            });
            let filtersArray = this.getFilterArray(rule);
            this.noSelectedEndpoint = false;
            this.setFilter(filtersArray);
            this.copyUpdateForm = clone(this.ruleForm.value);
            this.copySpeedUnit = this.selectedUnit;
            // keep trigger same value
            this.copyUpdateForm.trigger = clone(rule.trigger);
            this.copyUpdateForm.filters =
                this.copyUpdateForm.filters === null
                    ? []
                    : this.copyUpdateForm.filters;
            // set filter value is [] if callback filter value is null.
        }, 100);
        // end of reset the filter list.
    }

    get filters(): FormArray {
        return this.ruleForm.get('filters') as FormArray;
    }
    setFilter(filters: Filter[]) {
        const filterFGs = filters.map(filter => {
            if (filter.type === FilterType.LABEL) {
                let fbLabel = this.fb.group({
                    type: FilterType.LABEL,
                    decoration: filter.decoration || Decoration.MATCHES,
                });
                let filterLabel = this.fb.array(filter.value);
                fbLabel.setControl('value', filterLabel);
                return fbLabel;
            }
            if (filter.type === FilterType.TAG) {
                return this.fb.group({
                    type: FilterType.TAG,
                    decoration: filter.decoration || Decoration.MATCHES,
                    value: filter.value,
                });
            }
            return this.fb.group(filter);
        });
        const filterFormArray = this.fb.array(filterFGs);
        this.ruleForm.setControl('filters', filterFormArray);
    }

    initFilter(name: string) {
        if (name === FilterType.LABEL) {
            const labelArray = this.fb.array([]);
            const labelControl = this.fb.group({
                type: name,
                decoration: Decoration.MATCHES,
            });
            labelControl.setControl('value', labelArray);
            return labelControl;
        }
        if (name === FilterType.TAG) {
            return this.fb.group({
                type: name,
                decoration: Decoration.MATCHES,
                value: '',
            });
        }
        return this.fb.group({
            type: name,
            value: '',
        });
    }

    targetChange($event: any) {
        if ($event && $event.target) {
            if ($event.target['value'] === '-1') {
                this.noSelectedEndpoint = true;
                return;
            }
            this.noSelectedEndpoint = false;
        }
    }

    checkRuleName(): void {
        let ruleName: string = this.ruleForm.controls['name'].value;
        if (ruleName) {
            this.nameChecker.next(ruleName);
        } else {
            this.ruleNameTooltip = 'REPLICATION.NAME_TOOLTIP';
        }
    }

    public hasFormChange(): boolean {
        if (this.copySpeedUnit !== this.selectedUnit) {
            // speed unit has been changed
            return true;
        }
        return !isEmptyObject(this.hasChanges());
    }

    onSubmit() {
        if (this.ruleForm.value.trigger.type !== 'scheduled') {
            this.ruleForm
                .get('trigger')
                .get('trigger_settings')
                .get('cron')
                .setValue('');
        }
        // add new Replication rule
        this.inProgress = true;
        let copyRuleForm: ReplicationPolicy = this.ruleForm.value;
        // need to convert unit to KB for speed property
        copyRuleForm.speed = this.convertToKB(copyRuleForm.speed);
        copyRuleForm.dest_namespace_replace_count =
            copyRuleForm.dest_namespace_replace_count
                ? parseInt(
                      copyRuleForm.dest_namespace_replace_count.toString(),
                      10
                  )
                : 0;
        if (this.isPushMode) {
            copyRuleForm.src_registry = null;
        } else {
            copyRuleForm.dest_registry = null;
        }
        let filters: any = copyRuleForm.filters;
        // remove the filters which user not set.
        for (let i = filters.length - 1; i >= 0; i--) {
            if (
                !filters[i].value ||
                (filters[i].value instanceof Array &&
                    filters[i].value.length === 0)
            ) {
                copyRuleForm.filters.splice(i, 1);
            }
        }

        if (this.policyId < 0) {
            this.repService.createReplicationRule(copyRuleForm).subscribe(
                () => {
                    this.translateService
                        .get('REPLICATION.CREATED_SUCCESS')
                        .subscribe(res => this.errorHandler.info(res));
                    this.inProgress = false;
                    this.reload.emit(true);
                    this.createForm();
                    this.close();
                },
                (error: any) => {
                    this.inProgress = false;
                    this.inlineAlert.showInlineError(error);
                }
            );
        } else {
            this.repService
                .updateReplicationRule(this.policyId, this.ruleForm.value)
                .subscribe(
                    () => {
                        this.translateService
                            .get('REPLICATION.UPDATED_SUCCESS')
                            .subscribe(res => this.errorHandler.info(res));
                        this.inProgress = false;
                        this.reload.emit(true);
                        this.close();
                    },
                    (error: any) => {
                        this.inProgress = false;
                        this.inlineAlert.showInlineError(error);
                    }
                );
        }
    }
    openCreateEditRule(rule?: ReplicationPolicy): void {
        this.formReset();
        this.copyUpdateForm = clone(this.ruleForm.value);
        this.inlineAlert.close();
        this.noSelectedEndpoint = true;
        this.isRuleNameValid = true;
        this.policyId = -1;
        this.createEditRuleOpened = true;
        this.noEndpointInfo = '';
        if (this.targetList.length === 0) {
            this.noEndpointInfo = 'REPLICATION.NO_ENDPOINT_INFO';
        }
        if (rule) {
            if (
                this.supportedFilterLabels &&
                this.supportedFilterLabels.length
            ) {
                this.supportedFilterLabels.forEach((label, index) => {
                    if (rule.filters && rule.filters.length) {
                        rule.filters.forEach(f => {
                            if (
                                f.type === FilterType.LABEL &&
                                f.value &&
                                (f.value as any).length
                            ) {
                                (f.value as any).forEach(name => {
                                    if (label.name === name) {
                                        this.supportedFilterLabels[
                                            index
                                        ].select = true;
                                    }
                                });
                            }
                        });
                    }
                });
            }
            this.onGoing = true;
            this.policyId = +rule.id;
            this.headerTitle = 'REPLICATION.EDIT_POLICY_TITLE';
            this.repService.getReplicationRule(rule.id).subscribe(
                ruleInfo => {
                    let srcRegistryId = ruleInfo.src_registry.id;
                    this.repService
                        .getRegistryInfo(srcRegistryId)
                        .pipe(finalize(() => (this.onGoing = false)))
                        .subscribe(
                            adapter => {
                                this.setFilterAndTrigger(adapter);
                                this.updateRuleFormAndCopyUpdateForm(ruleInfo);
                            },
                            (error: any) => {
                                this.translateService
                                    .get(
                                        'REPLICATION.UNREACHABLE_SOURCE_REGISTRY',
                                        {
                                            error: errorHandlerFn(error),
                                        }
                                    )
                                    .subscribe(translatedResponse => {
                                        this.inlineAlert.showInlineError(
                                            translatedResponse
                                        );
                                    });
                            }
                        );
                },
                (error: any) => {
                    this.onGoing = false;
                    this.inlineAlert.showInlineError(error);
                }
            );
        } else {
            this.onGoing = true;
            let registryObs = this.repService.getRegistryInfo(0);
            registryObs.pipe(finalize(() => (this.onGoing = false))).subscribe(
                adapter => {
                    this.setFilterAndTrigger(adapter);
                    this.copyUpdateForm = clone(this.ruleForm.value);
                    if (
                        this.supportedFilterLabels &&
                        this.supportedFilterLabels.length
                    ) {
                        this.supportedFilterLabels.forEach((label, index) => {
                            label.select = false;
                        });
                    }
                },
                (error: any) => {
                    this.inlineAlert.showInlineError(error);
                }
            );
            this.headerTitle = 'REPLICATION.ADD_POLICY';
            this.copyUpdateForm = clone(this.ruleForm.value);
        }
    }

    setFilterAndTrigger(adapter) {
        this.supportedFilters = adapter.supported_resource_filters;
        this.supportedFilters.forEach(element => {
            this.filters.push(this.initFilter(element.type));
        });

        this.supportedTriggers = adapter.supported_triggers;
        this.ruleForm
            .get('trigger')
            .get('type')
            .setValue(this.supportedTriggers[0]);
    }
    close(): void {
        this.createEditRuleOpened = false;
    }

    confirmCancel(confirmed: boolean) {
        this.inlineAlert.close();
        this.createForm();
        this.close();
    }

    onCancel(): void {
        if (this.hasFormChange()) {
            this.inlineAlert.showInlineConfirmation({
                message: 'ALERT.FORM_CHANGE_CONFIRMATION',
            });
        } else {
            this.createForm();
            this.close();
        }
    }

    goRegistry(): void {
        this.goToRegistry.emit();
    }

    hasChanges(): boolean {
        const formValue = clone(this.ruleForm.value);
        const initValue = clone(this.copyUpdateForm);
        return !isSameObject(formValue, initValue);
    }

    getFilterArray(rule): Array<any> {
        let filtersArray = [];
        for (let i = 0; i < this.supportedFilters.length; i++) {
            let findTag: boolean = false;
            if (rule.filters) {
                rule.filters.forEach(ruleItem => {
                    if (this.supportedFilters[i].type === ruleItem.type) {
                        filtersArray.push(ruleItem);
                        findTag = true;
                    }
                });
            }

            if (!findTag) {
                if (this.supportedFilters[i].type === FilterType.LABEL) {
                    filtersArray.push({
                        type: this.supportedFilters[i].type,
                        value: [],
                    });
                } else {
                    filtersArray.push({
                        type: this.supportedFilters[i].type,
                        value: '',
                    });
                }
            }
        }
        return filtersArray;
    }
    cronInputShouldShowError(): boolean {
        return (
            this.ruleForm &&
            this.ruleForm.get('trigger') &&
            this.ruleForm.get('trigger').get('trigger_settings') &&
            this.ruleForm.get('trigger').get('trigger_settings').get('cron') &&
            (this.ruleForm.get('trigger').get('trigger_settings').get('cron')
                .touched ||
                this.ruleForm.get('trigger').get('trigger_settings').get('cron')
                    .dirty) &&
            this.ruleForm.get('trigger').get('trigger_settings').get('cron')
                .value &&
            !cronRegex(
                this.ruleForm.get('trigger').get('trigger_settings').get('cron')
                    .value
            )
        );
    }
    stickLabel(value, index) {
        value.select = !value.select;
        let filters = this.ruleForm.get('filters') as FormArray;
        let fromIndex = filters.controls[index] as FormGroup;
        let labelValue = this.supportedFilterLabels.reduce(
            (cumulatedSelectedArrs, currentValue) => {
                if (currentValue.select) {
                    if (!cumulatedSelectedArrs.length) {
                        return [currentValue.name];
                    }
                    return [...cumulatedSelectedArrs, currentValue.name];
                }
                return cumulatedSelectedArrs;
            },
            []
        );

        fromIndex.setControl('value', this.fb.array(labelValue));
    }
    // set prefix '0 ', so user can not set item of 'seconds'
    inputInvalid(e: any) {
        if (this.headerTitle === 'REPLICATION.ADD_POLICY') {
            // adding model
            if (e && e.target) {
                if (
                    !e.target.value ||
                    (e.target.value && e.target.value.indexOf(PREFIX)) !== 0
                ) {
                    e.target.value = PREFIX;
                }
                e.target.value = e.target.value.replace(/\s+/g, ' ');
                if (e.target.value && e.target.value.split(/\s+/g).length > 6) {
                    e.target.value = e.target.value.trim();
                }
            }
        }
    }
    // when trigger type is scheduled, should set cron prefix to '0 '
    changeTrigger(e: any) {
        if (this.headerTitle === 'REPLICATION.ADD_POLICY') {
            // adding model
            if (
                e &&
                e.target &&
                e.target.value === this.TRIGGER_TYPES.SCHEDULED
            ) {
                this.ruleForm
                    .get('trigger')
                    .get('trigger_settings')
                    .get('cron')
                    .setValue(PREFIX);
            }
        }
    }
    getAllLabels(): void {
        // get all global labels
        this.labelService
            .ListLabelsResponse({
                pageSize: PAGE_SIZE,
                page: 1,
                scope: 'g',
            })
            .subscribe(res => {
                if (res.headers) {
                    const xHeader: string = res.headers.get('X-Total-Count');
                    const totalCount = parseInt(xHeader, 0);
                    let arr = res.body || [];
                    if (totalCount <= PAGE_SIZE) {
                        // already gotten all global labels
                        if (arr && arr.length) {
                            arr.forEach(data => {
                                this.supportedFilterLabels.push({
                                    name: data.name,
                                    color: data.color ? data.color : '#FFFFFF',
                                    select: false,
                                    scope: 'g',
                                });
                            });
                        }
                    } else {
                        // get all the global labels in specified times
                        const times: number = Math.ceil(totalCount / PAGE_SIZE);
                        const observableList: Observable<Label[]>[] = [];
                        for (let i = 2; i <= times; i++) {
                            observableList.push(
                                this.labelService.ListLabels({
                                    page: i,
                                    pageSize: PAGE_SIZE,
                                    scope: 'g',
                                })
                            );
                        }
                        forkJoin(observableList).subscribe(response => {
                            if (response && response.length) {
                                response.forEach(item => {
                                    arr = arr.concat(item);
                                });
                                arr.forEach(data => {
                                    this.supportedFilterLabels.push({
                                        name: data.name,
                                        color: data.color
                                            ? data.color
                                            : '#FFFFFF',
                                        select: false,
                                        scope: 'g',
                                    });
                                });
                            }
                        });
                    }
                }
            });
    }
    // convert 'MB' or 'KB' to 'KB'
    convertToKB(inputValue: number): number {
        if (!inputValue) {
            // return default value
            return -1;
        }
        if (this.selectedUnit === BandwidthUnit.KB) {
            return +inputValue;
        }
        return inputValue * KB_TO_MB;
    }
    // convert 'KB' to 'MB' or 'KB'
    convertToInputValue(realSpeed: number): number {
        if (realSpeed >= KB_TO_MB && realSpeed % KB_TO_MB === 0) {
            this.selectedUnit = BandwidthUnit.MB;
            return Math.ceil(realSpeed / KB_TO_MB);
        } else {
            this.selectedUnit = BandwidthUnit.KB;
            return realSpeed ? realSpeed : -1;
        }
    }
    getLabel(labelName: string): Label {
        if (this.supportedFilterLabels?.length) {
            let label: Label;
            this.supportedFilterLabels.forEach(item => {
                if (item.name === labelName) {
                    label = item;
                }
            });
            return label;
        }
    }
}
