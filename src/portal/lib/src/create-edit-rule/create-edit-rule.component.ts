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

const ONE_HOUR_SECONDS = 3600;
const ONE_DAY_SECONDS: number = 24 * ONE_HOUR_SECONDS;

@Component({
  selector: "hbr-create-edit-rule",
  templateUrl: "./create-edit-rule.component.html",
  styleUrls: ["./create-edit-rule.component.scss"]
})
export class CreateEditRuleComponent implements OnInit, OnDestroy {
  _localTime: Date = new Date();
  sourceList: Endpoint[] = [];
  targetList: Endpoint[] = [];
  selectSrcNamespaces: any[] = [];
  selectDestNamespaces: any = [];
  selectSrcNamespacesNoSrcInfo: string;
  selectDestNamespacesNoSrcInfo: string;
  disabled: boolean;
  isFilterHide = false;
  weeklySchedule: boolean;
  noEndpointInfo = "";
  isPushMode = true;
  noSelectedEndpoint = true;
  filterCount = 0;
  alertClosed = false;
  TRIGGER_TYPES = {
    MANUAL: "manual",
    SCHEDULED: "scheduled",
    EVENT_BASED: "event_based"
  };

  ruleNameTooltip = "REPLICATION.NAME_TOOLTIP";
  headerTitle = "REPLICATION.ADD_POLICY";

  createEditRuleOpened: boolean;
  filterListData: { [key: string]: any }[] = [];
  inProgress = false;
  inNameChecking = false;
  isRuleNameValid = true;
  nameChecker: Subject<string> = new Subject<string>();
  namespaceChecker: Subject<object> = new Subject<{}>();
  firstClick = 0;
  isNamespacesValid: Boolean;
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
  // the index when user input src_namespace_list ;
  src_namespace_index: number = 0;
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
  /**
   * src_namespaces or dest_namespace input
   * @param index form array index
   * @param formControlName src_namespaces / dest_namespace
   * @param selectRegistryId push or pull : dest_registry.id/0 or   0/src_registry
   * @param acceptArrayName array name of project list
   */
  handleValidation(index: number, formControlName: string, selectRegistryId = 0, acceptArrayName): void {
      let cont = this.ruleForm.controls[formControlName] as FormArray ;
      let cont1 = cont.controls[index].value;

      if (cont1) {
        this.namespaceChecker.next({value: cont1, index: index, formControlName: formControlName
          , selectRegistryId: selectRegistryId, acceptArrayName: acceptArrayName});
      }
  }

  leaveInput() {
    this.selectSrcNamespaces = [];
    this.selectDestNamespaces = [];
  }

  selectedSrcName(srcName: string, index, formControlName: string, acceptArray: any) {
    let pro: any = acceptArray.find(
      data => data.name === srcName
    );
    let pro1 = pro.name;
    this.setNamespace(pro1, index, formControlName);
    acceptArray = [];
    this.selectSrcNamespacesNoSrcInfo = "";
    this.selectDestNamespacesNoSrcInfo = "";
  }

  initRegistryInfo(id: number): void {
    this.repService.getRegistryInfo(id).subscribe(adapter => {
      this.supportedFilters = adapter.supported_resource_filters;
      this.supportedFilters.forEach(element => {
        this.filters.push(this.initFilter(element.type));
      });

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

      this.namespaceChecker
      .pipe(debounceTime(500))
      .pipe(distinctUntilChanged())
      .subscribe((resn: any) => {
        let inputValue = resn.value;
        this.selectSrcNamespacesNoSrcInfo = "";
        this.selectDestNamespacesNoSrcInfo = "";
        this[resn.acceptArrayName] = [];
        this.endpointService.listNamespaces(resn.selectRegistryId, inputValue)
        .subscribe(res => {
          if (res) {
            this[resn.acceptArrayName] = res.slice(0, 10);
            let pro = res.find((data: any) => data.name === inputValue);
            if (!pro) {
              this[`${resn.acceptArrayName}NoSrcInfo`] = "REPLICATION.NO_PROJECT_INFO";
            } else {
              pro = pro.name;
              this[`${resn.acceptArrayName}NoSrcInfo`] = "";
              this.setNamespace(pro, resn.index, resn.formControlName);
            }
          } else {
            this[`${resn.acceptArrayName}NoSrcInfo`] = "REPLICATION.NO_PROJECT_INFO";
          }
        }, error => {
          this[`${resn.acceptArrayName}NoSrcInfo`] = "REPLICATION.NO_PROJECT_INFO";
        });
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
  }

  sourceChange($event: any): void {
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
  get src_namespaces(): FormArray { return this.ruleForm.get('src_namespaces') as FormArray; }
  get dest_namespace(): FormArray { return this.ruleForm.get('dest_namespace') as FormArray; }

  setNamespace(namespaces: string, index: number, formControlName: string): void {
    let newNameForm = this.ruleForm.controls[formControlName] as FormArray;
    newNameForm.controls[index].setValue(namespaces);
  }

  get isValid() {
    let controlName = this.ruleForm.controls["name"].invalid;
    let controlSrcNamespace = this.ruleForm.controls["src_namespaces"].invalid;
    return !(
      controlName ||
      !this.isRuleNameValid ||
      this.noSelectedEndpoint ||
      controlSrcNamespace ||
      this.inProgress
    );
  }

  createForm() {
    this.ruleForm = this.fb.group({
      name: ["", Validators.required],
      description: "",
      src_registry: new FormControl(),
      src_namespaces: new FormArray([new FormControl('')], Validators.required),
      dest_registry: new FormControl(),
      dest_namespace: new FormArray([new FormControl('')]),
      trigger: this.fb.group({
        type: '',
        trigger_settings: this.fb.group({
          cron: ""
        })
      }),
      filters: this.fb.array([]),
      deletion: false,
      enabled: true
    });
  }

  selectTrigger($event: any): void {

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
      enabled: true
    });
    this.isPushMode = true;
  }

  initForm(): void {
    this.formReset();
    this.setFilter([]);
    this.initRegistryInfo(0);
    this.copyUpdateForm = clone(this.ruleForm.value);
  }

  updateForm(rule: ReplicationRule): void {
    if (rule.dest_registry.id === 0) {
      this.isPushMode = false;
    } else {
      this.isPushMode = true;
    }
    setTimeout(() => {
      this.ruleForm.reset({
        name: rule.name,
        description: rule.description,
        src_namespaces: rule.src_namespaces,
        dest_namespace: [rule.dest_namespace],
        src_registry: rule.src_registry,
        dest_registry: rule.dest_registry,
        trigger: rule.trigger,
        deletion: rule.deletion,
        enabled: rule.enabled
      });
    // reset the filter list.
    let filters = [];
    for (let i = 0; i < this.supportedFilters.length; i++) {
      let findTag: boolean = false;
      if (rule.filters) {
        rule.filters.forEach((ruleItem, j) => {
          if (this.supportedFilters[i].type === ruleItem.type) {
            filters.push(ruleItem);
            findTag = true;
          }
        });
      }

      if (!findTag) {
        filters.push({ type: this.supportedFilters[i].type, value: "" });
      }

    }

    this.noSelectedEndpoint = false;
    this.setFilter(filters);
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

  focusClear($event: any): void {
    if (this.policyId < 0 && this.firstClick === 0) {
      if ($event && $event.target && $event.target["value"]) {
        $event.target["value"] = "";
      }
      this.firstClick++;
    }
  }

  addNewNamespace(): void {
    this.src_namespaces.push(new FormControl());
  }

  deleteNamespace(index: number): void {
    this.src_namespaces.removeAt(index);
  }

  // Replication Schedule select value exchange
  selectSchedule($event: any): void {

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
    // add new Replication rule
    this.inProgress = true;
    let copyRuleForm: ReplicationRule = this.ruleForm.value;
    // for dest_namespace array type change
    copyRuleForm.dest_namespace = copyRuleForm.dest_namespace.length ? copyRuleForm.dest_namespace[0] : '';
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
    this.filterCount = 0;
    this.isFilterHide = false;
    this.filterListData = [];
    this.firstClick = 0;
    this.noSelectedEndpoint = true;
    this.isRuleNameValid = true;
    this.selectSrcNamespaces = [];
    this.selectDestNamespaces = [];

    this.weeklySchedule = false;
    this.policyId = -1;
    this.createEditRuleOpened = true;
    this.noEndpointInfo = "";
    this.selectSrcNamespacesNoSrcInfo = "";
    this.selectDestNamespacesNoSrcInfo = "";
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
              this.copyUpdateForm = clone(ruleInfo);
              // set filter value is [] if callback filter value is null.
              this.updateForm(ruleInfo);
              // keep trigger same value
              this.copyUpdateForm.trigger = clone(ruleInfo.trigger);
              this.copyUpdateForm.filters = this.copyUpdateForm.filters === null ? [] : this.copyUpdateForm.filters;
          }, (error: any) => {
            this.inlineAlert.showInlineError(error);
          });
        }, (error: any) => {
          this.inlineAlert.showInlineError(error);
        });
    } else {
      let registryObs = this.repService.getRegistryInfo(0);
      registryObs.subscribe(adapter => { this.setFilterAndTrigger(adapter); });
      this.headerTitle = "REPLICATION.ADD_POLICY";
      this.copyUpdateForm = clone(this.ruleForm.value);
    }
  }

  setFilterAndTrigger(adapter: { supported_resource_filters: Filter[]; supported_triggers: string[]; }) {
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
    this.close();
  }

  onCancel(): void {
    if (this.hasFormChange()) {
      this.inlineAlert.showInlineConfirmation({
        message: "ALERT.FORM_CHANGE_CONFIRMATION"
      });
    } else {
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
}
