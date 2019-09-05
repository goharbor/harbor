
import {debounceTime, distinctUntilChanged} from 'rxjs/operators';
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
  EventEmitter,
  Output,
  ViewChild,
  OnInit,
  OnDestroy,
  Input,
  OnChanges,
  SimpleChanges
} from "@angular/core";
import { NgForm, Validators, AbstractControl } from "@angular/forms";

import { Subject } from "rxjs";
import { TranslateService } from "@ngx-translate/core";

import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { InlineAlertComponent } from "../../shared/inline-alert/inline-alert.component";

import { Project } from "../project";
import { ProjectService, QuotaUnits, QuotaHardInterface, QuotaUnlimited, getByte
  , GetIntegerAndUnit, clone, validateLimit, validateCountLimit} from "@harbor/ui";

@Component({
  selector: "create-project",
  templateUrl: "create-project.component.html",
  styleUrls: ["create-project.scss"]
})
export class CreateProjectComponent implements OnInit, OnChanges, OnDestroy {

  projectForm: NgForm;

  @ViewChild("projectForm", {static: true})
  currentForm: NgForm;
  quotaUnits = QuotaUnits;
  project: Project = new Project();
  countLimit: number;
  storageLimit: number;
  storageLimitUnit: string = QuotaUnits[3].UNIT;
  storageDefaultLimit: number;
  storageDefaultLimitUnit: string;
  countDefaultLimit: number;
  initVal: Project = new Project();

  createProjectOpened: boolean;

  hasChanged: boolean;
  isSubmitOnGoing = false;

  staticBackdrop = true;
  closable = false;

  isNameValid = true;
  nameTooltipText = "PROJECT.NAME_TOOLTIP";
  checkOnGoing = false;
  proNameChecker: Subject<string> = new Subject<string>();

  @Output() create = new EventEmitter<boolean>();
  @Input() quotaObj: QuotaHardInterface;
  @Input() isSystemAdmin: boolean;
  @ViewChild(InlineAlertComponent, {static: true})
  inlineAlert: InlineAlertComponent;

  constructor(private projectService: ProjectService,
    private translateService: TranslateService,
    private messageHandlerService: MessageHandlerService) { }

  ngOnInit(): void {
    this.proNameChecker.pipe(
      debounceTime(300))
      .subscribe((name: string) => {
        let cont = this.currentForm.controls["create_project_name"];
        if (cont) {
          this.isNameValid = cont.valid;
          if (this.isNameValid) {
            // Check exiting from backend
            this.projectService
              .checkProjectExists(cont.value)
              .subscribe(() => {
                // Project existing
                this.isNameValid = false;
                this.nameTooltipText = "PROJECT.NAME_ALREADY_EXISTS";
                this.checkOnGoing = false;
              }, error => {
                this.checkOnGoing = false;
              });
          } else {
            this.nameTooltipText = "PROJECT.NAME_TOOLTIP";
          }
        }
      });
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes && changes["quotaObj"] && changes["quotaObj"].currentValue) {
      this.countLimit = this.quotaObj.count_per_project;
      this.storageLimit = GetIntegerAndUnit(this.quotaObj.storage_per_project, clone(QuotaUnits), 0, clone(QuotaUnits)).partNumberHard;
      this.storageLimitUnit = this.storageLimit === QuotaUnlimited ? QuotaUnits[3].UNIT
      : GetIntegerAndUnit(this.quotaObj.storage_per_project, clone(QuotaUnits), 0, clone(QuotaUnits)).partCharacterHard;

      this.countDefaultLimit = this.countLimit;
      this.storageDefaultLimit = this.storageLimit;
      this.storageDefaultLimitUnit = this.storageLimitUnit;
      if (this.isSystemAdmin) {
        this.currentForm.form.controls['create_project_storage_limit'].setValidators(
          [
          Validators.required,
          Validators.pattern('(^-1$)|(^([1-9]+)([0-9]+)*$)'),
          validateLimit(this.currentForm.form.controls['create_project_storage_limit_unit'])
      ]);
      this.currentForm.form.controls['create_project_count_limit'].setValidators(
        [
          Validators.required,
          Validators.pattern('(^-1$)|(^([1-9]+)([0-9]+)*$)'),
          validateCountLimit()
        ]);
      }
      this.currentForm.form.valueChanges
            .pipe(distinctUntilChanged((a, b) => JSON.stringify(a) === JSON.stringify(b)))
            .subscribe((data) => {
              ['create_project_storage_limit', 'create_project_storage_limit_unit', 'create_project_count_limit'].forEach(fieldName => {
                if (this.currentForm.form.get(fieldName) && this.currentForm.form.get(fieldName).value !== null) {
                  this.currentForm.form.get(fieldName).updateValueAndValidity();
                }
              });
            });
    }
}
  ngOnDestroy(): void {
    this.proNameChecker.unsubscribe();
  }

  onSubmit() {
    if (this.isSubmitOnGoing) {
      return ;
    }
    this.isSubmitOnGoing = true;
    const storageByte = +this.storageLimit === QuotaUnlimited ? this.storageLimit : getByte(+this.storageLimit, this.storageLimitUnit);
    this.projectService
      .createProject(this.project.name, this.project.metadata, +this.countLimit, +storageByte)
      .subscribe(
      status => {
        this.isSubmitOnGoing = false;

        this.create.emit(true);
        this.messageHandlerService.showSuccess("PROJECT.CREATED_SUCCESS");
        this.createProjectOpened = false;
      },
      error => {
        this.isSubmitOnGoing = false;
        this.inlineAlert.showInlineError(error);
      });
  }

  onCancel() {
      this.createProjectOpened = false;
  }

  newProject() {
    this.project = new Project();
    this.hasChanged = false;
    this.isNameValid = true;

    this.createProjectOpened = true;
    this.inlineAlert.close();

    this.countLimit = this.countDefaultLimit ;
    this.storageLimit = this.storageDefaultLimit;
    this.storageLimitUnit = this.storageDefaultLimitUnit;
  }

  public get isValid(): boolean {
    return this.currentForm &&
    this.currentForm.valid &&
    !this.isSubmitOnGoing &&
    this.isNameValid &&
    !this.checkOnGoing;
  }

  // Handle the form validation
  handleValidation(): void {
    let cont = this.currentForm.controls["create_project_name"];
    if (cont) {
      this.proNameChecker.next(cont.value);
    }

  }
}

