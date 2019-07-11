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
  selector: "edit-rule",
  templateUrl: "./edit-rule.component.html",
  styleUrls: ["./edit-rule.component.scss"]
})
export class EditRuleComponent implements OnInit, OnDestroy {

  editRuleOpened: boolean = false;
  template: string;
  number: number;
  repositories: string;
  tagsSelect: string;
  tagsInput: string;
  labelsSelect: string;
  labelsInput: string;
  rule: Rule = new Rule();
  @Output() clickSave = new EventEmitter<Rule>();
  @Input()
  metadata: any;
  constructor(
  ) {}
  ngOnInit(): void {
  }
  ngOnDestroy(): void {
  }
  init() {
    this.template = this.rule.template;
    this.number = this.rule.params.num;
    if (this.rule.scope_selectors.repository.length > 0 ) {
      let strArr = [];
      this.rule.scope_selectors.repository.forEach(rep => {
        strArr.push(rep.pattern);
      });
      this.repositories = strArr.join(",");
    }
    if (this.rule.tag_selectors.length > 0) {
      let tags = [];
      let labels = [];
      this.rule.tag_selectors.forEach(tag => {
        if (tag.kind === "doublestar") {
          tags.push(tag.pattern);
          this.tagsSelect = tag.decoration;
        }
        if (tag.kind === "label") {
          labels.push(tag.pattern);
          this.labelsSelect = tag.decoration;
        }
      });
      this.tagsInput = tags.join(",");
      this.labelsInput = labels.join(",");
    }
  }
  open() {
    this.editRuleOpened = true;
  }
  close() {
    this.editRuleOpened = false;
  }
  cancel() {
    this.close();
  }
  save() {
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
    if (!this.labelsInput) {
      this.labelsInput = "**";
    }
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
    this.clickSave.emit(rule);
  }
  canAdd(): boolean {
    return !(this.template === 'always' || (this.template && this.number));
  }
}
