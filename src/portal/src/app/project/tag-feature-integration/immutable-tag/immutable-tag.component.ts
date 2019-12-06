import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { AddRuleComponent } from "./add-rule/add-rule.component";
import { ImmutableTagService } from "./immutable-tag.service";
import { ImmutableRetentionRule } from "../tag-retention/retention";
import { finalize } from "rxjs/operators";
import { ErrorHandler } from "../../../../lib/utils/error-handler";
import { clone } from "../../../../lib/utils/utils";
import { forkJoin } from 'rxjs';

@Component({
  selector: 'app-immutable-tag',
  templateUrl: './immutable-tag.component.html',
  styleUrls: ['./immutable-tag.component.scss']
})
export class ImmutableTagComponent implements OnInit {
  projectId: number;
  selectedItem: any = null;
  ruleIndex: number = -1;
  index: number = -1;
  rules: ImmutableRetentionRule[] = [];
  editIndex: number;
  loadingRule: boolean = true;

  @ViewChild('addRule', { static: false }) addRuleComponent: AddRuleComponent;
  constructor(
    private route: ActivatedRoute,
    private immutableTagService: ImmutableTagService,
    public errorHandler: ErrorHandler,
  ) {
  }

  ngOnInit() {
    this.projectId = +this.route.snapshot.parent.parent.params['id'];
    forkJoin(this.immutableTagService.getRules(this.projectId), this.getMetadata())
    .pipe(finalize(() => {
      this.loadingRule = false;
    }))
    .subscribe(
      response => {
        this.rules = response[0] as ImmutableRetentionRule[];
        this.addRuleComponent.metadata = response[1];
        this.loadingRule = false;
      }, error => {
        this.errorHandler.error(error);
        this.loadingRule = false;
      });
  }

  getMetadata() {
    return this.immutableTagService.getRetentionMetadata();
  }

  getRules() {
    this.immutableTagService.getRules(this.projectId).subscribe(
      response => {
        this.rules = response as ImmutableRetentionRule[];
        this.loadingRule = false;
      }, error => {
        this.errorHandler.error(error);
        this.loadingRule = false;
      });
  }

  editRuleByIndex(index) {
    this.editIndex = index;
    this.addRuleComponent.rule = clone(this.rules[index]);
    this.addRuleComponent.editRuleOrigin = clone(this.rules[index]);
    this.addRuleComponent.open();
    this.addRuleComponent.isAdd = false;
    this.ruleIndex = -1;
  }
  toggleDisable(rule, isActionDisable) {
    let cloneRule = clone(rule);
    cloneRule.disabled = isActionDisable;
    this.ruleIndex = -1;
    this.loadingRule = true;
    this.immutableTagService.updateRule(this.projectId, cloneRule).subscribe(
      response => {
        this.getRules();
      }, error => {
        this.loadingRule = false;
        this.errorHandler.error(error);
      });
  }
  deleteRule(ruleId) {
    // // if rules is empty, clear schedule.
    this.ruleIndex = -1;
    this.loadingRule = true;
    this.immutableTagService.deleteRule(this.projectId, ruleId).subscribe(
      response => {
        this.getRules();
      }, error => {
        this.loadingRule = false;
        this.errorHandler.error(error);
      });
  }

  openAddRule() {
    this.addRuleComponent.open();
    this.addRuleComponent.isAdd = true;
    this.addRuleComponent.rule = new ImmutableRetentionRule(this.projectId);
  }

  openEditor(index) {
    if (this.ruleIndex !== index) {
      this.ruleIndex = index;
    } else {
      this.ruleIndex = -1;
    }
  }

  refreshAfterCreatRetention() {
    this.immutableTagService.getProjectInfo(this.projectId).subscribe(
      response => {
        this.getRules();
      }, error => {
        this.loadingRule = false;
        this.errorHandler.error(error);
      });
  }

  clickAdd(rule) {
    this.loadingRule = true;
    this.addRuleComponent.onGoing = true;
    if (this.addRuleComponent.isAdd) {
      if (!rule.id) {
        this.immutableTagService.createRule(this.projectId, rule)
          .pipe(finalize(() => this.addRuleComponent.onGoing = false)).subscribe(
            response => {
              this.refreshAfterCreatRetention();
              this.addRuleComponent.close();
            }, error => {
              if (error && error.error && error.error.message) {
                error = this.immutableTagService.getI18nKey(error.error.message);
              }
              this.addRuleComponent.inlineAlert.showInlineError(error);
              this.loadingRule = false;
            });
      } else {
        this.immutableTagService.updateRule(this.projectId, rule)
          .pipe(finalize(() => this.addRuleComponent.onGoing = false)).subscribe(
            response => {
              this.getRules();
              this.addRuleComponent.close();
            }, error => {
              this.loadingRule = false;
              if (error && error.error && error.error.message) {
                error = this.immutableTagService.getI18nKey(error.error.message);
              }
              this.addRuleComponent.inlineAlert.showInlineError(error);
            });
      }
    } else {
      this.immutableTagService.updateRule(this.projectId, rule)
        .pipe(finalize(() => this.addRuleComponent.onGoing = false)).subscribe(
          response => {
            this.getRules();
            this.addRuleComponent.close();
          }, error => {
            if (error && error.error && error.error.message) {
              error = this.immutableTagService.getI18nKey(error.error.message);
            }
            this.addRuleComponent.inlineAlert.showInlineError(error);
            this.loadingRule = false;
          });
    }
  }

  formatPattern(pattern: string): string {
    return pattern.replace(/[{}]/g, "");
  }

  getI18nKey(str: string) {
    return this.immutableTagService.getI18nKey(str);
  }
}

