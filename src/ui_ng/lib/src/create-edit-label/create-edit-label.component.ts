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
  Output,
  EventEmitter,
  OnDestroy,
  Input,
  OnInit,
  ViewChild,
  ChangeDetectionStrategy,
  ChangeDetectorRef
} from "@angular/core";

import { Label } from "../service/interface";

import { toPromise, clone, compareValue } from "../utils";

import { LabelService } from "../service/label.service";
import { ErrorHandler } from "../error-handler/error-handler";
import { NgForm } from "@angular/forms";
import { Subject } from "rxjs/Subject";
import { LabelColor } from "../shared/shared.const";

@Component({
  selector: "hbr-create-edit-label",
  templateUrl: "./create-edit-label.component.html",
  styleUrls: ["./create-edit-label.component.scss"],
  changeDetection: ChangeDetectionStrategy.Default
})
export class CreateEditLabelComponent implements OnInit, OnDestroy {
  formShow: boolean;
  inProgress: boolean;
  copeLabelModel: Label;
  labelModel: Label = this.initLabel();
  labelId = 0;

  checkOnGoing: boolean;
  isLabelNameExist = false;

  nameChecker = new Subject<string>();

  labelForm: NgForm;
  @ViewChild("labelForm") currentForm: NgForm;

  @Input() projectId: number;
  @Input() scope: string;
  @Output() reload = new EventEmitter();

  constructor(
    private labelService: LabelService,
    private errorHandler: ErrorHandler,
    private ref: ChangeDetectorRef
  ) {}

  ngOnInit(): void {
    this.nameChecker.debounceTime(500).subscribe((name: string) => {
      this.checkOnGoing = true;
      let labelName = this.currentForm.controls["name"].value;
      toPromise<Label[]>(
        this.labelService.getLabels(this.scope, this.projectId, labelName)
      )
        .then(targets => {
          if (targets && targets.length) {
            this.isLabelNameExist = true;
          } else {
            this.isLabelNameExist = false;
          }
          this.checkOnGoing = false;
        })
        .catch(error => {
          this.checkOnGoing = false;
          this.errorHandler.error(error);
        });
      setTimeout(() => {
        setInterval(() => this.ref.markForCheck(), 100);
      }, 3000);
    });
  }

  ngOnDestroy(): void {
    this.nameChecker.unsubscribe();
  }

  get labelColor() {
    return LabelColor;
  }

  initLabel(): Label {
    return {
      name: "",
      description: "",
      color: "",
      scope: "",
      project_id: 0
    };
  }
  openModal(): void {
    this.labelModel = this.initLabel();
    this.formShow = true;
    this.isLabelNameExist = false;
    this.labelId = 0;
    this.copeLabelModel = null;
  }

  editModel(labelId: number, label: Label[]): void {
    this.labelModel = clone(label[0]);
    this.formShow = true;
    this.labelId = labelId;
    this.copeLabelModel = clone(label[0]);
  }

  public get hasChanged(): boolean {
    return !compareValue(this.copeLabelModel, this.labelModel);
  }

  public get isValid(): boolean {
    return !(
      this.checkOnGoing ||
      this.isLabelNameExist ||
      !(this.currentForm && this.currentForm.valid) ||
      !this.hasChanged ||
      this.inProgress
    );
  }

  existValid(text: string): void {
    if (text) {
      this.nameChecker.next(text);
    }
  }

  onSubmit(): void {
    this.inProgress = true;
    if (this.labelId <= 0) {
      this.labelModel.scope = this.scope;
      this.labelModel.project_id = this.projectId;
      toPromise<Label>(this.labelService.createLabel(this.labelModel))
        .then(res => {
          this.inProgress = false;
          this.reload.emit();
          this.labelModel = this.initLabel();
          this.formShow = false;
        })
        .catch(err => {
          this.inProgress = false;
          this.errorHandler.error(err);
        });
    } else {
      toPromise<Label>(
        this.labelService.updateLabel(this.labelId, this.labelModel)
      )
        .then(res => {
          this.inProgress = false;
          this.reload.emit();
          this.labelModel = this.initLabel();
          this.formShow = false;
        })
        .catch(err => {
          this.inProgress = false;
          this.errorHandler.error(err);
        });
    }
  }

  onCancel(): void {
    this.inProgress = false;
    this.labelModel = this.initLabel();
    this.formShow = false;
  }
}
