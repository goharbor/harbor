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
import { Subject } from "rxjs/Subject";
import { Subscription } from "rxjs/Subscription";
import { FormArray, FormBuilder, FormGroup, Validators } from "@angular/forms";
import { compareValue, isEmptyObject, toPromise } from "../utils";
import { InlineAlertComponent } from "../inline-alert/inline-alert.component";
import { ReplicationService } from "../service/replication.service";
import { ErrorHandler } from "../error-handler/error-handler";
import { TranslateService } from "@ngx-translate/core";
import { EndpointService } from "../service/endpoint.service";
import { ProjectService } from "../service/project.service";
import { Project } from "../project-policy-config/project";

const ONE_HOUR_SECONDS = 3600;
const ONE_DAY_SECONDS: number = 24 * ONE_HOUR_SECONDS;

@Component({
  selector: "hbr-create-edit-rule",
  templateUrl: "./create-edit-rule.component.html",
  styleUrls: ["./create-edit-rule.component.scss"]
})
export class CreateEditRuleComponent implements OnInit, OnDestroy {
  _localTime: Date = new Date();
  targetList: Endpoint[] = [];
  projectList: Project[] = [];
  selectedProjectList: Project[] = [];
  isFilterHide = false;
  weeklySchedule: boolean;
  isScheduleOpt: boolean;
  isImmediate = false;
  noProjectInfo = "";
  noEndpointInfo = "";
  noSelectedProject = true;
  noSelectedEndpoint = true;
  filterCount = 0;
  triggerNames: string[] = ["Manual", "Immediate", "Scheduled"];
  scheduleNames: string[] = ["Daily", "Weekly"];
  weekly: string[] = [
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
    "Sunday"
  ];
  filterSelect: string[] = ["repository", "tag"];
  ruleNameTooltip = "TOOLTIP.EMPTY";
  headerTitle = "REPLICATION.ADD_POLICY";

  createEditRuleOpened: boolean;
  filterListData: { [key: string]: any }[] = [];
  inProgress = false;
  inNameChecking = false;
  isRuleNameExist = false;
  nameChecker: Subject<string> = new Subject<string>();
  proNameChecker: Subject<string> = new Subject<string>();
  firstClick = 0;
  policyId: number;

  confirmSub: Subscription;
  ruleForm: FormGroup;
  copyUpdateForm: ReplicationRule;

  @Input() projectId: number;
  @Input() projectName: string;

  @Output() goToRegistry = new EventEmitter<any>();
  @Output() reload = new EventEmitter<boolean>();

  @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;

  emptyProject = {
    project_id: -1,
    name: ""
  };
  emptyEndpoint = {
    id: -1,
    endpoint: "",
    name: "",
    username: "",
    password: "",
    insecure: true,
    type: 0
  };
  constructor(
    private fb: FormBuilder,
    private repService: ReplicationService,
    private endpointService: EndpointService,
    private errorHandler: ErrorHandler,
    private proService: ProjectService,
    private translateService: TranslateService,
    public ref: ChangeDetectorRef
  ) {
    this.createForm();
  }

  baseFilterData(name: string, option: string[], state: boolean) {
    return {
      name: name,
      options: option,
      state: state,
      isValid: true
    };
  }

  ngOnInit(): void {
    toPromise<Endpoint[]>(this.endpointService.getEndpoints())
      .then(targets => {
        this.targetList = targets || [];
      })
      .catch((error: any) => this.errorHandler.error(error));

    if (!this.projectId) {
      toPromise<Project[]>(this.proService.listProjects("", undefined))
        .then(targets => {
          this.projectList = targets || [];
        })
        .catch(error => this.errorHandler.error(error));
    }

    this.nameChecker
      .debounceTime(500)
      .distinctUntilChanged()
      .subscribe((ruleName: string) => {
        this.isRuleNameExist = false;
        this.inNameChecking = true;
        toPromise<ReplicationRule[]>(
          this.repService.getReplicationRules(0, ruleName)
        )
          .then(response => {
            if (response.some(rule => rule.name === ruleName)) {
              this.ruleNameTooltip = "TOOLTIP.RULE_USER_EXISTING";
              this.isRuleNameExist = true;
            }
            this.inNameChecking = false;
          })
          .catch(() => {
            this.inNameChecking = false;
          });
      });

    this.proNameChecker
      .debounceTime(500)
      .distinctUntilChanged()
      .subscribe((name: string) => {
        this.noProjectInfo = "";
        this.selectedProjectList = [];
        toPromise<Project[]>(this.proService.listProjects(name, undefined))
          .then((res: any) => {
            if (res) {
              this.selectedProjectList = res.slice(0, 10);
              // if input value exit in project list
              let pro = res.find((data: any) => data.name === name);
              if (!pro) {
                this.noProjectInfo = "REPLICATION.NO_PROJECT_INFO";
                this.noSelectedProject = true;
              } else {
                this.noProjectInfo = "";
                this.noSelectedProject = false;
                this.setProject([pro]);
              }
            } else {
              this.noProjectInfo = "REPLICATION.NO_PROJECT_INFO";
              this.noSelectedProject = true;
            }
          })
          .catch((error: any) => {
            this.errorHandler.error(error);
            this.noProjectInfo = "REPLICATION.NO_PROJECT_INFO";
            this.noSelectedProject = true;
          });
      });
  }

  ngOnDestroy(): void {
    if (this.confirmSub) {
      this.confirmSub.unsubscribe();
    }
    if (this.nameChecker) {
      this.nameChecker.unsubscribe();
    }
    if (this.proNameChecker) {
      this.proNameChecker.unsubscribe();
    }
  }

  get isValid() {
    return !(
      this.isRuleNameExist ||
      this.noSelectedProject ||
      this.noSelectedEndpoint ||
      this.inProgress
    );
  }

  createForm() {
    this.ruleForm = this.fb.group({
      name: ["", Validators.required],
      description: "",
      projects: this.fb.array([]),
      targets: this.fb.array([]),
      trigger: this.fb.group({
        kind: this.triggerNames[0],
        schedule_param: this.fb.group({
          type: this.scheduleNames[0],
          weekday: 1,
          offtime: "08:00"
        })
      }),
      filters: this.fb.array([]),
      replicate_existing_image_now: true,
      replicate_deletion: false
    });
  }

  initForm(): void {
    this.ruleForm.reset({
      name: "",
      description: "",
      trigger: {
        kind: this.triggerNames[0],
        schedule_param: {
          type: this.scheduleNames[0],
          weekday: 1,
          offtime: "08:00"
        }
      },
      replicate_existing_image_now: true,
      replicate_deletion: false
    });
    this.setProject([this.emptyProject]);
    this.setTarget([this.emptyEndpoint]);
    this.setFilter([]);

    this.copyUpdateForm = Object.assign({}, this.ruleForm.value);
  }

  updateForm(rule: ReplicationRule): void {
    rule.trigger = this.updateTrigger(rule.trigger);
    this.ruleForm.reset({
      name: rule.name,
      description: rule.description,
      trigger: rule.trigger,
      replicate_existing_image_now: rule.replicate_existing_image_now,
      replicate_deletion: rule.replicate_deletion
    });
    this.setProject(rule.projects);
    this.noSelectedProject = false;
    this.setTarget(rule.targets);
    this.noSelectedEndpoint = false;

    if (rule.filters) {
      this.setFilter(rule.filters);
      this.updateFilter(rule.filters);
    }

    // Force refresh view
    let hnd = setInterval(() => this.ref.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 2000);
  }

  get projects(): FormArray {
    return this.ruleForm.get("projects") as FormArray;
  }
  setProject(projects: Project[]) {
    const projectFGs = projects.map(project => this.fb.group(project));
    const projectFormArray = this.fb.array(projectFGs);
    this.ruleForm.setControl("projects", projectFormArray);
  }

  get filters(): FormArray {
    return this.ruleForm.get("filters") as FormArray;
  }
  setFilter(filters: Filter[]) {
    const filterFGs = filters.map(filter => this.fb.group(filter));
    const filterFormArray = this.fb.array(filterFGs);
    this.ruleForm.setControl("filters", filterFormArray);
  }

  get targets(): FormArray {
    return this.ruleForm.get("targets") as FormArray;
  }
  setTarget(targets: Endpoint[]) {
    const targetFGs = targets.map(target => this.fb.group(target));
    const targetFormArray = this.fb.array(targetFGs);
    this.ruleForm.setControl("targets", targetFormArray);
  }

  initFilter(name: string) {
    return this.fb.group({
      kind: name,
      pattern: ["", Validators.required]
    });
  }

  filterChange($event: any) {
    if ($event && $event.target["value"]) {
      let id: number = $event.target.id;
      let name: string = $event.target.name;
      let value: string = $event.target["value"];

      this.filterListData.forEach((data, index) => {
        if (index === +id) {
          data.name = $event.target.name = value;
        } else {
          data.options.splice(data.options.indexOf(value), 1);
        }
        if (data.options.indexOf(name) === -1) {
          data.options.push(name);
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
      let selecedTarget: Endpoint = this.targetList.find(
        target => target.id === +$event.target["value"]
      );
      this.setTarget([selecedTarget]);
      this.noSelectedEndpoint = false;
    }
  }

  // Handle the form validation
  handleValidation(): void {
    let cont = this.ruleForm.controls["projects"];
    if (cont && cont.valid) {
      this.proNameChecker.next(cont.value[0].name);
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

  selectedProjectName(projectName: string) {
    this.noSelectedProject = false;
    let pro: Project = this.selectedProjectList.find(
      data => data.name === projectName
    );
    this.setProject([pro]);
    this.selectedProjectList = [];
    this.noProjectInfo = "";
  }

  selectedProject(project: Project): void {
    if (!project) {
      this.noSelectedProject = true;
    } else {
      this.noSelectedProject = false;
      this.setProject([project]);
    }
  }

  addNewFilter(): void {
    if (this.filterCount === 0) {
      this.filterListData.push(
        this.baseFilterData(
          this.filterSelect[0],
          this.filterSelect.slice(),
          true
        )
      );
      this.filters.push(this.initFilter(this.filterSelect[0]));
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
      this.filters.push(this.initFilter(nameArr[0]));
    }
    this.filterCount += 1;
    if (this.filterCount >= this.filterSelect.length) {
      this.isFilterHide = true;
    }
  }

  // delete a filter
  deleteFilter(i: number): void {
    if (i || i === 0) {
      let delfilter = this.filterListData.splice(i, 1)[0];
      if (this.filterCount === this.filterSelect.length) {
        this.isFilterHide = false;
      }
      this.filterCount -= 1;
      if (this.filterListData.length) {
        let optionVal = delfilter.name;
        this.filterListData.filter(data => {
          if (data.options.indexOf(optionVal) === -1) {
            data.options.push(optionVal);
          }
        });
      }
      const control = <FormArray>this.ruleForm.controls["filters"];
      control.removeAt(i);
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
    if ($event && $event.target && $event.target["value"]) {
      switch ($event.target["value"]) {
        case this.scheduleNames[1]:
          this.weeklySchedule = true;
          this.ruleForm.patchValue({
            trigger: {
              schedule_param: {
                weekday: 1
              }
            }
          });
          break;
        case this.scheduleNames[0]:
          this.weeklySchedule = false;
          break;
      }
    }
  }

  checkRuleName(): void {
    let ruleName: string = this.ruleForm.controls["name"].value;
    if (ruleName) {
      this.nameChecker.next(ruleName);
    } else {
      this.ruleNameTooltip = "TOOLTIP.EMPTY";
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

  updateTrigger(trigger: any) {
    if (trigger["schedule_param"]) {
      this.isScheduleOpt = true;
      this.isImmediate = false;
      trigger["schedule_param"]["offtime"] = this.getOfftime(
        trigger["schedule_param"]["offtime"]
      );
      if (trigger["schedule_param"]["weekday"]) {
        this.weeklySchedule = true;
      } else {
        // set default
        trigger["schedule_param"]["weekday"] = 1;
      }
    } else {
      if (trigger["kind"] === this.triggerNames[0]) {
        this.isImmediate = false;
      }
      if (trigger["kind"] === this.triggerNames[1]) {
        this.isImmediate = true;
      }
      trigger["schedule_param"] = {
        type: this.scheduleNames[0],
        weekday: this.weekly[0],
        offtime: "08:00"
      };
    }
    return trigger;
  }

  setTriggerVaule(trigger: any) {
    if (!this.isScheduleOpt) {
      delete trigger["schedule_param"];
      return trigger;
    } else {
      if (!this.weeklySchedule) {
        delete trigger["schedule_param"]["weekday"];
      } else {
        trigger["schedule_param"]["weekday"] = +trigger["schedule_param"][
          "weekday"
        ];
      }
      trigger["schedule_param"]["offtime"] = this.setOfftime(
        trigger["schedule_param"]["offtime"]
      );
      return trigger;
    }
  }

  public hasFormChange(): boolean {
    return !isEmptyObject(this.getChanges());
  }

  onSubmit() {
    // add new Replication rule
    this.inProgress = true;
    let copyRuleForm: ReplicationRule = this.ruleForm.value;
    copyRuleForm.trigger = this.setTriggerVaule(copyRuleForm.trigger);
    if (this.policyId < 0) {
      this.repService
        .createReplicationRule(copyRuleForm)
        .then(() => {
          this.translateService
            .get("REPLICATION.CREATED_SUCCESS")
            .subscribe(res => this.errorHandler.info(res));
          this.inProgress = false;
          this.reload.emit(true);
          this.close();
        })
        .catch((error: any) => {
          this.inProgress = false;
          this.inlineAlert.showInlineError(error);
        });
    } else {
      this.repService
        .updateReplicationRule(this.policyId, this.ruleForm.value)
        .then(() => {
          this.translateService
            .get("REPLICATION.UPDATED_SUCCESS")
            .subscribe(res => this.errorHandler.info(res));
          this.inProgress = false;
          this.reload.emit(true);
          this.close();
        })
        .catch((error: any) => {
          this.inProgress = false;
          this.inlineAlert.showInlineError(error);
        });
    }
  }

  openCreateEditRule(ruleId?: number | string): void {
    this.initForm();
    this.selectedProjectList = [];
    this.filterCount = 0;
    this.filterListData = [];
    this.firstClick = 0;
    this.noSelectedProject = true;
    this.noSelectedEndpoint = true;
    this.isRuleNameExist = false;

    this.weeklySchedule = false;
    this.isScheduleOpt = false;
    this.isImmediate = false;
    this.policyId = -1;
    this.createEditRuleOpened = true;

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
      toPromise(this.repService.getReplicationRule(ruleId))
        .then(response => {
          this.copyUpdateForm = Object.assign({}, response);
          // set filter value is [] if callback fiter value is null.
          this.copyUpdateForm.filters = response.filters
            ? response.filters
            : [];
          this.updateForm(response);
        })
        .catch((error: any) => {
          this.inlineAlert.showInlineError(error);
        });
    } else {
      this.headerTitle = "REPLICATION.ADD_POLICY";
      if (this.projectId) {
        this.setProject([
          { project_id: this.projectId, name: this.projectName }
        ]);
        this.noSelectedProject = false;
      }

      this.copyUpdateForm = Object.assign({}, this.ruleForm.value);
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

  // UTC time
  public getOfftime(daily_time: any): string {
    let timeOffset = 0; // seconds
    if (daily_time && typeof daily_time === "number") {
      timeOffset = +daily_time;
    }

    // Convert to current time
    let timezoneOffset: number = this._localTime.getTimezoneOffset();
    // Local time
    timeOffset = timeOffset - timezoneOffset * 60;
    if (timeOffset < 0) {
      timeOffset = timeOffset + ONE_DAY_SECONDS;
    }

    if (timeOffset >= ONE_DAY_SECONDS) {
      timeOffset -= ONE_DAY_SECONDS;
    }

    // To time string
    let hours: number = Math.floor(timeOffset / ONE_HOUR_SECONDS);
    let minutes: number = Math.floor(
      (timeOffset - hours * ONE_HOUR_SECONDS) / 60
    );

    let timeStr: string = "" + hours;
    if (hours < 10) {
      timeStr = "0" + timeStr;
    }
    if (minutes < 10) {
      timeStr += ":0";
    } else {
      timeStr += ":";
    }
    timeStr += minutes;

    return timeStr;
  }
  public setOfftime(v: string) {
    if (!v || v === "") {
      return;
    }

    let values: string[] = v.split(":");
    if (!values || values.length !== 2) {
      return;
    }

    let hours: number = +values[0];
    let minutes: number = +values[1];
    // Convert to UTC time
    let timezoneOffset: number = this._localTime.getTimezoneOffset();
    let utcTimes: number = hours * ONE_HOUR_SECONDS + minutes * 60;
    utcTimes += timezoneOffset * 60;
    if (utcTimes < 0) {
      utcTimes += ONE_DAY_SECONDS;
    }

    if (utcTimes >= ONE_DAY_SECONDS) {
      utcTimes -= ONE_DAY_SECONDS;
    }

    return utcTimes;
  }

  getChanges(): { [key: string]: any | any[] } {
    let changes: { [key: string]: any | any[] } = {};
    let ruleValue: { [key: string]: any | any[] } = this.ruleForm.value;
    if (!ruleValue || !this.copyUpdateForm) {
      return changes;
    }
    for (let prop of Object.keys(ruleValue)) {
      let field: any = this.copyUpdateForm[prop];
      if (!compareValue(field, ruleValue[prop])) {
        if (
          ruleValue[prop][0] &&
          ruleValue[prop][0].project_id &&
          ruleValue[prop][0].project_id === field[0].project_id
        ) {
          break;
        }
        if (
          ruleValue[prop][0] &&
          ruleValue[prop][0].id &&
          ruleValue[prop][0].id === field[0].id
        ) {
          break;
        }
        changes[prop] = ruleValue[prop];
        // Number
        if (typeof field === "number") {
          changes[prop] = +changes[prop];
        }

        // Trim string value
        if (typeof field === "string") {
          changes[prop] = ("" + changes[prop]).trim();
        }
      }
    }

    return changes;
  }
}
