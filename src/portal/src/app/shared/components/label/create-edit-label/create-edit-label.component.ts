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
} from '@angular/core';
import { clone, compareValue } from '../../../units/utils';
import { ErrorHandler } from '../../../units/error-handler';
import { NgForm } from '@angular/forms';
import { Subject } from 'rxjs';
import { debounceTime } from 'rxjs/operators';
import { LabelColor } from '../../../entities/shared.const';
import { Label } from '../../../../../../ng-swagger-gen/models/label';
import { LabelService } from '../../../../../../ng-swagger-gen/services/label.service';

@Component({
    selector: 'hbr-create-edit-label',
    templateUrl: './create-edit-label.component.html',
    styleUrls: ['./create-edit-label.component.scss'],
})
export class CreateEditLabelComponent implements OnInit, OnDestroy {
    formShow: boolean;
    inProgress: boolean;
    copeLabelModel: Label;
    labelModel: Label = this.initLabel();
    labelId = 0;

    isLabelNameExist = false;

    nameChecker = new Subject<string>();

    labelForm: NgForm;
    @ViewChild('labelForm', { static: true }) currentForm: NgForm;

    @Input() projectId: number;
    @Input() scope: string;
    @Output() reload = new EventEmitter();

    constructor(
        private labelService: LabelService,
        private errorHandler: ErrorHandler
    ) {}

    ngOnInit(): void {
        this.nameChecker.pipe(debounceTime(500)).subscribe((name: string) => {
            this.labelService
                .ListLabels({
                    scope: this.scope,
                    projectId: this.projectId,
                    name: name,
                })
                .subscribe(
                    targets => {
                        this.isLabelNameExist = false;
                        if (targets && targets.length) {
                            if (
                                targets.find(target => {
                                    return (
                                        target.name === name &&
                                        target.id !== this.labelId
                                    );
                                })
                            ) {
                                this.isLabelNameExist = true;
                            }
                        }
                    },
                    error => {
                        this.errorHandler.error(error);
                    }
                );
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
            name: '',
            description: '',
            color: '',
            scope: '',
            project_id: 0,
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
            this.isLabelNameExist ||
            !(this.currentForm && this.currentForm.valid) ||
            !this.hasChanged ||
            this.inProgress
        );
    }

    existValid(text: string): void {
        if (text) {
            this.nameChecker.next(text);
        } else {
            this.isLabelNameExist = false;
        }
    }

    onSubmit(): void {
        this.inProgress = true;
        if (this.labelId <= 0) {
            this.labelModel.scope = this.scope;
            this.labelModel.project_id = this.projectId;
            this.labelService
                .CreateLabel({
                    label: this.labelModel,
                })
                .subscribe(
                    res => {
                        this.inProgress = false;
                        this.reload.emit();
                        this.labelModel = this.initLabel();
                        this.formShow = false;
                    },
                    err => {
                        this.inProgress = false;
                        this.errorHandler.error(err);
                    }
                );
        } else {
            this.labelService
                .UpdateLabel({
                    labelId: this.labelId,
                    label: this.labelModel,
                })
                .subscribe(
                    res => {
                        this.inProgress = false;
                        this.reload.emit();
                        this.labelModel = this.initLabel();
                        this.formShow = false;
                    },
                    err => {
                        this.inProgress = false;
                        this.errorHandler.error(err);
                    }
                );
        }
    }

    onCancel(): void {
        this.inProgress = false;
        this.labelModel = this.initLabel();
        this.formShow = false;
    }
}
