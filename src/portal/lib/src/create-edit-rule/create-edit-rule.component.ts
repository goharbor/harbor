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
import { Filter, ReplicationRule, Endpoint, Label } from "../service/interface";
import { Subject, Subscription } from "rxjs";
import { debounceTime, distinctUntilChanged } from "rxjs/operators";
import { FormArray, FormBuilder, FormGroup, Validators, FormControl } from "@angular/forms";
import { clone, compareValue, isEmptyObject } from "../utils";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import { ReplicationService } from "../service/replication.service";
import { ErrorHandler } from "../error-handler/error-handler";
import { TranslateService } from "@ngx-translate/core";
import { EndpointService } from "../service/endpoint.service";
import { ProjectService } from "../service/project.service";
import { Project } from "../project-policy-config/project";
import { LabelState } from "../tag/tag.component";

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
  projectList: Project[] = [];
  selectedProjectList: Project[] = [];
  isFilterHide = false;
  weeklySchedule: boolean;
  isScheduleOpt: boolean;
  isImmediate = false;
  noProjectInfo = "";
  noEndpointInfo = "";
  isPushMode = true;
  noSelectedProject = true;
  noSelectedEndpoint = true;
  filterCount = 0;
  alertClosed = false;
  triggerNames: string[] = ["Manual", "Immediate", "Scheduled"];
  filterSelect: string[] = ["type", "repository", "tag", "label"];
  ruleNameTooltip = "REPLICATION.NAME_TOOLTIP";
  headerTitle = "REPLICATION.ADD_POLICY";

  createEditRuleOpened: boolean;
  filterListData: { [key: string]: any }[] = [];
  inProgress = false;
  inNameChecking = false;
  isRuleNameValid = true;
  nameChecker: Subject<string> = new Subject<string>();
  firstClick = 0;
  policyId: number;
  labelInputVal = '';
  filterLabelInfo: Label[] = [];  // store filter selected labels` id
  deletedLabelCount = 0;
  deletedLabelInfo: string;
  confirmSub: Subscription;
  ruleForm: FormGroup;
  formArrayLabel: FormArray;
  copyUpdateForm: ReplicationRule;
  cronString: string;

  @Input() projectId: number;
  @Input() projectName: string;
  @Input() withAdmiral: boolean;

  @Output() goToRegistry = new EventEmitter<any>();
  @Output() reload = new EventEmitter<boolean>();

  @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
  constructor(
    private fb: FormBuilder,
    private repService: ReplicationService,
    private endpointService: EndpointService,
    private errorHandler: ErrorHandler,
    private proService: ProjectService,
    private translateService: TranslateService,
    private ref: ChangeDetectorRef
  ) {
    this.createForm();
  }

  baseFilterData(name: string, option: string[], state: boolean) {
    return {
      name: name,
      options: option,
      state: state,
      isValid: true,
      isOpen: false  // label list
    };
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


  sourceChange($event): void {
    if ($event && $event.target) {
      if ($event.target["value"] === "-1") {
        this.noSelectedEndpoint = true;
        return;
      }
      let selecedTarget: Endpoint = this.sourceList.find(
        source => source.id === +$event.target["value"]
      );
      this.noSelectedEndpoint = false;
    }

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

  get isValid() {
    return !(
      !this.isRuleNameValid ||
      this.noSelectedEndpoint ||
      this.inProgress
    );
  }

  createForm() {
    this.formArrayLabel = this.fb.array([]);
    this.ruleForm = this.fb.group({
      name: ["", Validators.required],
      description: "",
      src_registry_id: new FormControl(),
      src_namespaces: new FormArray([new FormControl('')], Validators.required),
      dest_registry_id: new FormControl(),
      dest_namespace: "",
      trigger: this.fb.group({
        kind: this.triggerNames[0],
        schedule_param: this.fb.group({
          cron: ""
        })
      }),
      filters: this.fb.array([]),
      deletion: false
    });
  }

  initForm(): void {
    this.ruleForm.reset({
      name: "",
      description: "",
      trigger: {
        kind: this.triggerNames[0],
        schedule_param: {
          cron: ""
        }
      },
      deletion: false
    });
    this.setFilter([]);

    this.copyUpdateForm = clone(this.ruleForm.value);
  }

  updateForm(rule: ReplicationRule): void {
    this.ruleForm.reset({
      name: rule.name,
      description: rule.description,
      src_namespaces: rule.src_namespaces,
      dest_namespace: rule.dest_namespace,
      trigger: rule.trigger,
      deletion: rule.deletion
    });

    this.noSelectedProject = false;
    this.noSelectedEndpoint = false;

    if (rule.filters) {
      this.reOrganizeLabel(rule.filters);
      this.setFilter(rule.filters);
      this.updateFilter(rule.filters);
    }

    // Force refresh view
    let hnd = setInterval(() => this.ref.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 2000);
  }

  // reorganize filter structure
  reOrganizeLabel(filterLabels: any[]): void {
    let count = 0;
    if (filterLabels.length) {
      this.filterLabelInfo = [];

      let delLabel = '';
      filterLabels.forEach((data: any) => {
        if (data.kind === this.filterSelect[3]) {
          if (!data.value.deleted) {
            count++;
            this.filterLabelInfo.push(data.value);
          } else {
            this.deletedLabelCount++;
            delLabel += data.value.name + ',';
          }
        }
      });

      this.translateService.get('REPLICATION.DELETED_LABEL_INFO', {
        param: delLabel
      }).subscribe((res: string) => {
        this.deletedLabelInfo = res;
        this.alertClosed = false;
      });

      // delete api return label info, replace with label count
      if (delLabel || count) {
        let len = filterLabels.length;
        for (let i = 0; i < len; i++) {
          let lab = filterLabels.find(data => data.kind === this.filterSelect[3]);
          if (lab) { filterLabels.splice(filterLabels.indexOf(lab), 1); }
        }
        filterLabels.push({ kind: 'label', value: count + ' labels' });
        this.labelInputVal = count.toString();
      }
    }
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
      kind: name,
      value: ''
    });
  }

  filterChange($event: any, selectedValue: string) {
    if ($event && $event.target["value"]) {
      let id: number = $event.target.id;
      let name: string = $event.target.name;
      let value: string = $event.target["value"];

      const controlArray = <FormArray>this.ruleForm.get('filters');
      this.filterListData.forEach((data, index) => {
        if (index === +id) {
          data.name = $event.target.name = value;
        } else {
          data.options.splice(data.options.indexOf(value), 1);
        }
        if (data.options.indexOf(name) === -1) {
          data.options.push(name);
        }

        // if before select, $event is label
        if (!this.withAdmiral && name === this.filterSelect[3] && data.name === value) {
          this.labelInputVal = controlArray.controls[index].get('value').value.split(' ')[0];
          data.isOpen = false;
          controlArray.controls[index].get('value').setValue('');
        }
        // if before select, $event is  not label
        if (!this.withAdmiral && data.name === this.filterSelect[2]) {
          if (this.labelInputVal) {
            controlArray.controls[index].get('value').setValue(this.labelInputVal + ' labels');
          } else {
            controlArray.controls[index].get('value').setValue('');
          }

          // this.labelInputVal = '';
          data.isOpen = false;
        }

      });
    }
  }

  // when input value is label, then open label panel
  openLabelList(labelTag: string, indexId: number, $event: any) {
    if (!this.withAdmiral && labelTag === this.filterSelect[3]) {
      this.filterListData.forEach((data, index) => {
        if (index === indexId) {
          data.isOpen = true;
        } else {
          data.isOpen = false;
        }
      });
    }
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

  leaveInput() {
    this.selectedProjectList = [];
  }

  addNewNamespace(): void {
    this.src_namespaces.push(new FormControl());
  }

  deleteNamespace(index: number): void {
    this.src_namespaces.removeAt(index);
  }

  addNewFilter(): void {
    const controlArray = <FormArray>this.ruleForm.get('filters');
    if (this.filterCount === 0) {
      this.filterListData.push(
        this.baseFilterData(
          this.filterSelect[0],
          this.filterSelect.slice(),
          true,
        )
      );
      controlArray.push(this.initFilter(this.filterSelect[0]));
    } else {
      let nameArr: string[] = this.filterSelect.slice();
      this.filterListData.forEach(data => {
        nameArr.splice(nameArr.indexOf(data.name), 1);
      });
      // when add a new filter,the filterListData should change the options
      this.filterListData.filter(data => {
        data.options.splice(data.options.indexOf(nameArr[0]), 1);
      });
      this.filterListData.push(this.baseFilterData(nameArr[0], nameArr, true));
      controlArray.push(this.initFilter(nameArr[0]));
    }
    this.filterCount += 1;
    if (this.filterCount >= this.filterSelect.length) {
      this.isFilterHide = true;
    }
    if (controlArray.controls[this.filterCount - 1].get('kind').value === this.filterSelect[3] && this.labelInputVal) {
      controlArray.controls[this.filterCount - 1].get('value').setValue(this.labelInputVal + ' labels');
    }

  }

  // delete a filter
  deleteFilter(i: number): void {
    if (i >= 0) {
      let delFilter = this.filterListData.splice(i, 1)[0];
      if (this.filterCount === this.filterSelect.length) {
        this.isFilterHide = false;
      }
      this.filterCount -= 1;
      if (this.filterListData.length) {
        let optionVal = delFilter.name;
        this.filterListData.filter(data => {
          if (data.options.indexOf(optionVal) === -1) {
            data.options.push(optionVal);
          }
        });
      }
      const control = <FormArray>this.ruleForm.get('filters');
      if (control.controls[i].get('kind').value === this.filterSelect[2]) {
        this.filterLabelInfo = [];
        this.labelInputVal = "";
      }
      control.removeAt(i);
      this.setFilter(control.value);
    }
  }

  selectTrigger($event: any): void {
    if ($event && $event.target && $event.target["value"]) {
      let val: string = $event.target["value"];
      if (val === this.triggerNames[2]) {
        this.isScheduleOpt = true;
        this.isImmediate = false;
      }
      if (val === this.triggerNames[1]) {
        this.isScheduleOpt = false;
        this.isImmediate = true;
      }
      if (val === this.triggerNames[0]) {
        this.isScheduleOpt = false;
        this.isImmediate = false;
      }
    }
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

  updateFilter(filters: any) {
    let opt: string[] = this.filterSelect.slice();
    filters.forEach((filter: any) => {
      opt.splice(opt.indexOf(filter.kind), 1);
    });
    filters.forEach((filter: any) => {
      let option: string[] = opt.slice();
      option.unshift(filter.kind);
      this.filterListData.push(this.baseFilterData(filter.kind, option, true));
    });
    this.filterCount = filters.length;
    if (filters.length === this.filterSelect.length) {
      this.isFilterHide = true;
    }
  }

  selectedLabelList(selectedLabels: LabelState[], indexId: number) {
    // set input value of filter label
    const controlArray = <FormArray>this.ruleForm.get('filters');

    this.filterListData.forEach((data, index) => {
      if (data.name === this.filterSelect[2]) {
        let labelsLength = selectedLabels.filter(lab => lab.iconsShow === true).length;
        if (labelsLength > 0) {
          controlArray.controls[index].get('value').setValue(labelsLength + ' labels');
          this.labelInputVal = labelsLength.toString();
        } else {
          controlArray.controls[index].get('value').setValue('');
        }
      }
    });

    // store filter label info
    this.filterLabelInfo = [];
    selectedLabels.forEach(data => {
      if (data.iconsShow === true) {
        this.filterLabelInfo.push(data.label);
      }
    });
  }

  setFilterLabelVal(filters: any[]) {
    let labels: any = filters.find(data => data.kind === this.filterSelect[2]);

    if (labels) {
      filters.splice(filters.indexOf(labels), 1);
      let info: any[] = [];
      this.filterLabelInfo.forEach(data => {
        info.push({ kind: 'label', value: data.id });
      });
      filters.push.apply(filters, info);
    }
  }

  public hasFormChange(): boolean {
    return !isEmptyObject(this.hasChanges());
  }

  onSubmit() {
    // add new Replication rule
    this.inProgress = true;
    let copyRuleForm: ReplicationRule = this.ruleForm.value;
    copyRuleForm.trigger = null;
    if (this.isPushMode) {
      copyRuleForm.src_registry_id = null;
    } else {
      copyRuleForm.dest_registry_id = null;
    }
    // rewrite key name of label when filer contain labels.
    if (copyRuleForm.filters) { this.setFilterLabelVal(copyRuleForm.filters); }

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
    this.initForm();
    this.inlineAlert.close();
    this.selectedProjectList = [];
    this.filterCount = 0;
    this.isFilterHide = false;
    this.filterListData = [];
    this.firstClick = 0;
    this.noSelectedProject = true;
    this.noSelectedEndpoint = true;
    this.isRuleNameValid = true;
    this.deletedLabelCount = 0;

    this.weeklySchedule = false;
    this.isScheduleOpt = false;
    this.isImmediate = false;
    this.policyId = -1;
    this.createEditRuleOpened = true;
    this.filterLabelInfo = [];
    this.labelInputVal = '';

    this.noProjectInfo = "";
    this.noEndpointInfo = "";
    if (this.targetList.length === 0) {
      this.noEndpointInfo = "REPLICATION.NO_ENDPOINT_INFO";
    }
    if (this.projectList.length === 0 && !this.projectName) {
      this.noProjectInfo = "REPLICATION.NO_PROJECT_INFO";
    }

    if (ruleId) {
      this.policyId = +ruleId;
      this.headerTitle = "REPLICATION.EDIT_POLICY_TITLE";
      this.repService.getReplicationRule(ruleId)
        .subscribe(response => {
          this.copyUpdateForm = clone(response);
          // set filter value is [] if callback filter value is null.
          this.updateForm(response);
          // keep trigger same value
          this.copyUpdateForm.trigger = clone(response.trigger);
          this.copyUpdateForm.filters = this.copyUpdateForm.filters === null ? [] : this.copyUpdateForm.filters;
        }, (error: any) => {
          this.inlineAlert.showInlineError(error);
        });
    } else {
      this.headerTitle = "REPLICATION.ADD_POLICY";
      this.copyUpdateForm = clone(this.ruleForm.value);
    }
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

    if (formValue.filters && formValue.filters.length > 0) {
      formValue.filters.forEach((data, index) => {
        if (data.kind === this.filterSelect[2]) {
          formValue.filters.splice(index, 1);
        }
      });
      // rewrite filter label
      this.filterLabelInfo.forEach(data => {
        formValue.filters.push({ kind: "label", pattern: "", value: data });
      });
    }

    if (!compareValue(formValue, initValueCopy)) {
      return true;
    }
    return false;
  }
}
