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
    Input,
    OnDestroy,
    Output,
    EventEmitter,
} from "@angular/core";
import {Rule} from "../retention";

@Component({
    selector: "add-rule",
    templateUrl: "./add-rule.component.html",
    styleUrls: ["./add-rule.component.scss"]
})
export class AddRuleComponent implements OnInit, OnDestroy {

    addRuleOpened: boolean = false;
    @Output() clickAdd = new EventEmitter<Rule>();
    template: string;
    number: number;
    repositories: string;
    tagsSelect: string;
    tagsInput: string;
    labelsSelect: string;
    labelsInput: string;
    @Input()
    metadata: any;
    constructor() {

    }

    ngOnInit(): void {
    }

    ngOnDestroy(): void {
    }

    init() {
        this.template = null;
        this.number = null;
        this.repositories = null;
        this.tagsSelect = null;
        this.tagsInput = null;
        this.labelsSelect = null;
        this.labelsInput = null;
    }

    canAdd(): boolean {
        return !(this.template === 'always' || (this.template && this.number));
    }

    open() {
        this.init();
        this.addRuleOpened = true;
    }

    close() {
        this.addRuleOpened = false;
    }

    cancel() {
        this.close();
    }

    add() {
        this.close();
        let rule = new Rule();
        rule.template = this.template;
        rule.params.num = this.number;
        // repositories
        if (!this.repositories) {
            this.repositories = "**";
        }
        let reps = this.repositories.split(/[,，]+/);
        if (reps && reps.length > 0) {
            reps.forEach(rep => {
                let repository = {
                    kind: "doublestar",
                    decoration: "matches",
                    pattern: null
                };
                repository.pattern = rep.trim();
                rule.scope_selectors.repository.push(repository);
            });
        }
        // tags
        if (!this.tagsInput) {
            this.tagsInput = "**";
        }
        if (this.tagsInput) {
            let decoration;
            if (this.tagsSelect) {
                decoration = this.tagsSelect;
            } else {
                decoration = "matches";
            }
            let tags = this.tagsInput.split(/[,，]+/);
            if (tags && tags.length > 0) {
                tags.forEach(tag => {
                    let selector = {
                        kind: "doublestar",
                        decoration: decoration,
                        pattern: null
                    };
                    selector.pattern = tag.trim();
                    rule.tag_selectors.push(selector);
                });
            }
        }
        // labels
        if (this.labelsInput) {
            let decoration;
            if (this.labelsSelect) {
                decoration = this.labelsSelect;
            } else {
                decoration = "with";
            }
            let labels = this.labelsInput.split(/[,，]+/);
            if (labels && labels.length > 0) {
                labels.forEach(label => {
                    let selector = {
                        kind: "label",
                        decoration: decoration,
                        pattern: null
                    };
                    selector.pattern = label.trim();
                    rule.tag_selectors.push(selector);
                });
            }
        }
        this.clickAdd.emit(rule);
    }
}
