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
  ChangeDetectorRef,
  Input,
  EventEmitter,
  Output
} from "@angular/core";
import { Filter, ReplicationRule, Endpoint } from "../service/interface";
import { Subject, Subscription, Observable, zip } from "rxjs";
import { debounceTime, distinctUntilChanged } from "rxjs/operators";
import { FormArray, FormBuilder, FormGroup, Validators, FormControl } from "@angular/forms";
import { clone, compareValue, isEmptyObject } from "../utils";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import { ReplicationService } from "../service/replication.service";
import { ErrorHandler } from "../error-handler/error-handler";
import { TranslateService } from "@ngx-translate/core";
import { EndpointService } from "../service/endpoint.service";
import { cronRegex } from "../utils";


@Component({
  selector: "hbr-create-edit-rule",
  templateUrl: "./create-edit-rule.component.html",
  styleUrls: ["./create-edit-rule.component.scss"]
})
export class CreateEditRuleComponent implements OnInit, OnDestroy {
  sourceList: Endpoint[] = [];
  targetList: Endpoint[] = [];
  noEndpointInfo = "";
  isPushMode = true;
  noSelectedEndpoint = true;
  TRIGGER_TYPES = {
    MANUAL: "manual",
    SCHEDULED: "scheduled",
    EVENT_BASED: "event_based"
  };

  ruleNameTooltip = "REPLICATION.NAME_TOOLTIP";
  headerTitle = "REPLICATION.ADD_POLICY";

  createEditRuleOpened: boolean;
  inProgress = false;
  onGoing = false;
  inNameChecking = false;
  isRuleNameValid = true;
  nameChecker: Subject<string> = new Subject<string>();
  policyId: number;
  confirmSub: Subscription;
  ruleForm: FormGroup;
  copyUpdateForm: ReplicationRule;
  cronString: string;
  supportedTriggers: string[];
  supportedFilters: Filter[];
  @Input() withAdmiral: boolean;

  @Output() goToRegistry = new EventEmitter<any>();
  @Output() reload = new EventEmitter<boolean>();

  @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
  constructor(
    private fb: FormBuilder,
    private repService: ReplicationService,
    private endpointService: EndpointService,
    private errorHandler: ErrorHandler,
    private translateService: TranslateService,
    private ref: ChangeDetectorRef
  ) {
    this.createForm();
  }

  initRegistryInfo(id: number): void {
    this.onGoing = true;
    this.repService.getRegistryInfo(id).subscribe(adapter => {
      this.supportedFilters = adapter.supported_resource_filters;
      this.supportedFilters.forEach(element => {
        this.filters.push(this.initFilter(element.type));
      });
      this.onGoing = false;
      this.supportedTriggers = adapter.supported_triggers;
      this.ruleForm.get("trigger").get("type").setValue(this.supportedTriggers[0]);
    });
  }

  ngOnInit(): void {
    this.endpointService.getEndpoints().subscribe(endPoints => {
      this.targetList = endPoints || [];
      this.sourceList = endPoints || [];
    }, error => {
      this.errorHandler.error(error);
    });
    this.nameChecker
      .pipe(debounceTime(300))
      .pipe(distinctUntilChanged())
      .subscribe((ruleName: string) => {
        let cont = this.ruleForm.controls["name"];
        if (cont) {
          this.isRuleNameValid = cont.valid;
          if (this.isRuleNameValid) {
            this.inNameChecking = true;
            this.repService.getReplicationRules(0, ruleName)
              .subscribe(response => {
                if (response.some(rule => rule.name === ruleName)) {
                  this.ruleNameTooltip = "TOOLTIP.RULE_USER_EXISTING";
                  this.isRuleNameValid = false;
                }
                this.inNameChecking = false;
              }, () => {
                this.inNameChecking = false;
              });
          } else {
            this.ruleNameTooltip = "REPLICATION.NAME_TOOLTIP";
          }
        }
      });
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
    let controlName = !!this.ruleForm.controls["name"].value;
    let sourceRegistry = !!this.ruleForm.controls["src_registry"].value;
    let destRegistry = !!this.ruleForm.controls["dest_registry"].value;
    let triggerMode = !!this.ruleForm.controls["trigger"].value.type;
    let cron = !!this.ruleForm.value.trigger.trigger_settings.cron;
    return !(!controlName ||
      !triggerMode ||
      !this.isRuleNameValid ||
      (!this.isPushMode && !sourceRegistry ||
        this.isPushMode && !destRegistry)
      || !(!this.isNotSchedule() && cron && this.cronInputValid(this.ruleForm.value.trigger.trigger_settings.cron || '')
        || this.isNotSchedule()));
  }

  createForm() {
    this.ruleForm = this.fb.group({
      name: ["", Validators.required],
      description: "",
      src_registry: new FormControl(),
      dest_registry: new FormControl(),
      dest_namespace: "",
      trigger: this.fb.group({
        type: '',
        trigger_settings: this.fb.group({
          cron: ""
        })
      }),
      filters: this.fb.array([]),
      deletion: false,
      enabled: true,
      override: true
    });
  }


  isNotSchedule(): boolean {
    return this.ruleForm.get("trigger").get("type").value !== this.TRIGGER_TYPES.SCHEDULED;
  }

  isNotEventBased(): boolean {
    return this.ruleForm.get("trigger").get("type").value !== this.TRIGGER_TYPES.EVENT_BASED;
  }

  formReset(): void {
    this.ruleForm.reset({
      name: "",
      description: "",
      trigger: {
        type: '',
        trigger_settings: {
          cron: ""
        }
      },
      deletion: false,
      enabled: true,
      override: true
    });
    this.isPushMode = true;
  }


  updateRuleFormAndCopyUpdateForm(rule: ReplicationRule): void {
    if (rule.dest_registry.id === 0) {
      this.isPushMode = false;
    } else {
      this.isPushMode = true;
    }
    setTimeout(() => {
      // There is no trigger_setting type when the harbor is upgraded from the old version.
      rule.trigger.trigger_settings = rule.trigger.trigger_settings ? rule.trigger.trigger_settings : {cron: ''};
      this.ruleForm.reset({
        name: rule.name,
        description: rule.description,
        dest_namespace: rule.dest_namespace,
        src_registry: rule.src_registry,
        dest_registry: rule.dest_registry,
        trigger: rule.trigger,
        deletion: rule.deletion,
        enabled: rule.enabled,
        override: rule.override
      });
      let filtersArray = this.getFilterArray(rule);

      this.noSelectedEndpoint = false;
      this.setFilter(filtersArray);
      this.copyUpdateForm = clone(this.ruleForm.value);
      // keep trigger same value
      this.copyUpdateForm.trigger = clone(rule.trigger);
      this.copyUpdateForm.filters = this.copyUpdateForm.filters === null ? [] : this.copyUpdateForm.filters;
      // set filter value is [] if callback filter value is null.
    }, 100);
    // end of reset the filter list.
  }

  get filters(): FormArray {
    return this.ruleForm.get("filters") as FormArray;
  }
  setFilter(filters: Filter[]) {
    const filterFGs = filters.map(filter => this.fb.group(filter));
    const filterFormArray = this.fb.array(filterFGs);
    this.ruleForm.setControl("filters", filterFormArray);
  }

  initFilter(name: string) {
    return this.fb.group({
      type: name,
      value: ''
    });
  }

  targetChange($event: any) {
    if ($event && $event.target) {
      if ($event.target["value"] === "-1") {
        this.noSelectedEndpoint = true;
        return;
      }
      this.noSelectedEndpoint = false;
    }
  }

  checkRuleName(): void {
    let ruleName: string = this.ruleForm.controls["name"].value;
    if (ruleName) {
      this.nameChecker.next(ruleName);
    } else {
      this.ruleNameTooltip = "REPLICATION.NAME_TOOLTIP";
    }
  }

  public hasFormChange(): boolean {
    return !isEmptyObject(this.hasChanges());
  }

  onSubmit() {
    if (this.ruleForm.value.trigger.type !== 'scheduled') {
      this.ruleForm.get("trigger").get("trigger_settings").get('cron').setValue('');
    }
    // add new Replication rule
    this.inProgress = true;
    let copyRuleForm: ReplicationRule = this.ruleForm.value;
    if (this.isPushMode) {
      copyRuleForm.src_registry = null;
    } else {
      copyRuleForm.dest_registry = null;
    }
    let filters: any = copyRuleForm.filters;
    // remove the filters which user not set.
    for (let i = filters.length - 1; i >= 0; i--) {
      if (filters[i].value === "") {
        copyRuleForm.filters.splice(i, 1);
      }
    }

    if (this.policyId < 0) {
      this.repService
        .createReplicationRule(copyRuleForm)
        .subscribe(() => {
          this.translateService
            .get("REPLICATION.CREATED_SUCCESS")
            .subscribe(res => this.errorHandler.info(res));
          this.inProgress = false;
          this.reload.emit(true);
          this.createForm();
          this.close();
        }, (error: any) => {
          this.inProgress = false;
          this.inlineAlert.showInlineError(error);
        });
    } else {
      this.repService
        .updateReplicationRule(this.policyId, this.ruleForm.value)
        .subscribe(() => {
          this.translateService
            .get("REPLICATION.UPDATED_SUCCESS")
            .subscribe(res => this.errorHandler.info(res));
          this.inProgress = false;
          this.reload.emit(true);
          this.close();
        }, (error: any) => {
          this.inProgress = false;
          this.inlineAlert.showInlineError(error);
        });
    }
  }
  openCreateEditRule(ruleId?: number | string): void {
    this.formReset();
    this.copyUpdateForm = clone(this.ruleForm.value);
    this.inlineAlert.close();
    this.noSelectedEndpoint = true;
    this.isRuleNameValid = true;

    this.policyId = -1;
    this.createEditRuleOpened = true;
    this.noEndpointInfo = "";
    if (this.targetList.length === 0) {
      this.noEndpointInfo = "REPLICATION.NO_ENDPOINT_INFO";
    }
    if (ruleId) {
      this.policyId = +ruleId;
      this.headerTitle = "REPLICATION.EDIT_POLICY_TITLE";
      this.repService.getReplicationRule(ruleId)
        .subscribe((ruleInfo) => {
          let srcRegistryId = ruleInfo.src_registry.id;
          this.repService.getRegistryInfo(srcRegistryId)
            .subscribe(adapter => {
              this.setFilterAndTrigger(adapter);
              this.updateRuleFormAndCopyUpdateForm(ruleInfo);
            }, (error: any) => {
              this.inlineAlert.showInlineError(error);
            });
        }, (error: any) => {
          this.inlineAlert.showInlineError(error);
        });
    } else {
      this.onGoing = true;
      let registryObs = this.repService.getRegistryInfo(0);
      registryObs.subscribe(adapter => {
        this.setFilterAndTrigger(adapter);
        this.copyUpdateForm = clone(this.ruleForm.value);
        this.onGoing = false;
      });
      this.headerTitle = "REPLICATION.ADD_POLICY";
      this.copyUpdateForm = clone(this.ruleForm.value);
    }
  }

  setFilterAndTrigger(adapter) {
    this.supportedFilters = adapter.supported_resource_filters;
    this.setFilter([]);
    this.supportedFilters.forEach(element => {
      this.filters.push(this.initFilter(element.type));
    });

    this.supportedTriggers = adapter.supported_triggers;
    this.ruleForm.get("trigger").get("type").setValue(this.supportedTriggers[0]);
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
        message: "ALERT.FORM_CHANGE_CONFIRMATION"
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
    let formValue = clone(this.ruleForm.value);
    let initValue = clone(this.copyUpdateForm);
    let initValueCopy: any = {};
    for (let key of Object.keys(formValue)) {
      initValueCopy[key] = initValue[key];
    }

    if (!compareValue(formValue, initValueCopy)) {
      return true;
    }
    return false;
  }


  getFilterArray(rule): Array<any> {
    let filtersArray = [];
    for (let i = 0; i < this.supportedFilters.length; i++) {
      let findTag: boolean = false;
      if (rule.filters) {
        rule.filters.forEach((ruleItem) => {
          if (this.supportedFilters[i].type === ruleItem.type) {
            filtersArray.push(ruleItem);
            findTag = true;
          }
        });
      }

      if (!findTag) {
        filtersArray.push({ type: this.supportedFilters[i].type, value: "" });
      }

    }
    return filtersArray;
  }
  cronInputValid(cronValue): boolean {
    return cronRegex(cronValue);
  }
  get cronTouched(): boolean {
    let triggerControl = this.ruleForm.controls.trigger as FormGroup;
    if (!triggerControl) {
      return false;
    }
    let trigger_settingsControls = triggerControl.controls.trigger_settings as FormGroup;
    if (!trigger_settingsControls) {
      return false;
    }
    return trigger_settingsControls.controls.cron.touched || trigger_settingsControls.controls.cron.dirty;
  }
}
