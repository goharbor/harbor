import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ImmutableRetentionRule } from "../tag-retention/retention";
import { ErrorHandler } from "../../../../shared/units/error-handler";
import { ImmutableTagService } from '../immutable-tag/immutable-tag.service';
import { AddAccelerationRuleComponent } from './add-rule/add-acceleration-rule.component';

@Component({
  selector: 'app-tag-acceleration',
  templateUrl: './tag-acceleration.component.html',
  styleUrls: ['./tag-acceleration.component.scss']
})
export class TagAccelerationComponent implements OnInit {
  projectId: number;
  selectedItem: any = null;
  ruleIndex: number = -1;
  index: number = -1;
  get rules(): ImmutableRetentionRule[] {
    return (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || []);
  }
  editIndex: number;
  loadingRule: boolean = false;

  @ViewChild('addRule') addRuleComponent: AddAccelerationRuleComponent;
  constructor(
    private route: ActivatedRoute,
    public errorHandler: ErrorHandler,
    private immutableTagService: ImmutableTagService,
  ) {
  }

  ngOnInit() {
    this.projectId = +this.route.snapshot.parent.parent.parent.params['id'];
    this.getRules();
    this.getMetadata();
  }

  getMetadata() {
  }

  getRules() {
  }
  editRuleByIndex(index) {
    this.editIndex = index;
    this.addRuleComponent.repoSelect =
        (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || [])[index].scope_selectors.repository[0].decoration;
    this.addRuleComponent.repositories =
        (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || [])[index]
            .scope_selectors.repository[0].pattern;
    this.addRuleComponent.tagsSelect = (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`))
        || [])[index].tag_selectors[0].decoration;
    this.addRuleComponent.tagsInput = (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || [])[index]
        .tag_selectors[0].pattern;
    this.addRuleComponent.repoSelectEditOrigin =
        (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || [])[index]
            .scope_selectors.repository[0].decoration;
    this.addRuleComponent.repositoriesEditOrigin =
        (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || [])[index]
            .scope_selectors.repository[0].pattern;
    this.addRuleComponent.tagsSelectEditOrigin =
        (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || [])[index]
            .tag_selectors[0].decoration;
    this.addRuleComponent.tagsInputEditOrigin =
        (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || [])[index]
            .tag_selectors[0].pattern;
    this.addRuleComponent.open();
    this.addRuleComponent.isAdd = false;
    this.ruleIndex = -1;
  }
  toggleDisable(index) {
    this.ruleIndex = -1;
    this.loadingRule = true;
    setTimeout(() => {
      this.loadingRule = false;
      const rules = (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || []);
      rules[index].disabled = !rules[index].disabled;
      localStorage.setItem(`MockedRules${this.projectId}`, JSON.stringify(rules));
    }, 500);
  }
  deleteRule(index) {
    // // if rules is empty, clear schedule.
    this.ruleIndex = -1;
    this.loadingRule = true;

    this.loadingRule = true;
    this.addRuleComponent.close();
    setTimeout(() => {
      const rules = (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || []);
      rules.splice(index, 1);
      localStorage.setItem(`MockedRules${this.projectId}`, JSON.stringify(rules));
      this.loadingRule = false;
    }, 500);
  }

  openAddRule() {
    this.addRuleComponent.open();
    this.addRuleComponent.repoSelect = 'repoMatches';
    this.addRuleComponent.repositories = '**';
    this.addRuleComponent.tagsSelect = 'matches';
    this.addRuleComponent.tagsInput = "**";
    this.addRuleComponent.isAdd = true;
  }

  openEditor(index) {
    if (this.ruleIndex !== index) {
      this.ruleIndex = index;
    } else {
      this.ruleIndex = -1;
    }
  }

  refreshAfterCreatRetention() {
  }

  clickAdd(e) {
    if (e.isAdd) {
      this.loadingRule = true;
      this.addRuleComponent.close();
      setTimeout(() => {
        const rules = (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || []);
        rules.push(e.rule);
        localStorage.setItem(`MockedRules${this.projectId}`, JSON.stringify(rules));
        this.loadingRule = false;
      }, 500);
    } else {
      this.loadingRule = true;
      this.addRuleComponent.close();
      setTimeout(() => {
        const rules = (JSON.parse(localStorage.getItem(`MockedRules${this.projectId}`)) || []);
        rules[this.editIndex] = e.rule;
        localStorage.setItem(`MockedRules${this.projectId}`, JSON.stringify(rules));
        this.loadingRule = false;
      }, 500);
    }
  }

  formatPattern(pattern: string): string {
    let str: string = pattern;
    if (/^{\S+}$/.test(str)) {
      return str.slice(1, str.length - 1);
    }
    return str;
  }

  getI18nKey(str: string) {
    return this.immutableTagService.getI18nKey(str);
  }
}

