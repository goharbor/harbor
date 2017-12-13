import { Component, Input, OnInit, ViewChild } from '@angular/core';

import { PROJECT_POLICY_CONFIG_TEMPLATE } from './project-policy-config.component.html';
import { PROJECT_POLICY_CONFIG_STYLE } from './project-policy-config.component.css';
import { toPromise, compareValue, clone } from '../utils';
import { ProjectService } from '../service/project.service';
import { ErrorHandler } from '../error-handler/error-handler';
import { State } from 'clarity-angular';

import { ConfirmationState, ConfirmationTargets, ConfirmationButtons } from '../shared/shared.const';
import { ConfirmationMessage } from '../confirmation-dialog/confirmation-message';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ConfirmationAcknowledgement } from '../confirmation-dialog/confirmation-state-message';
import { TranslateService } from '@ngx-translate/core';

import { Project } from './project';
import {SystemInfo, SystemInfoService} from '../service/index';

export class ProjectPolicy {
  Public: boolean;
  ContentTrust: boolean;
  PreventVulImg: boolean;
  PreventVulImgSeverity: string;
  ScanImgOnPush: boolean;

  constructor() {
    this.Public = false;
    this.ContentTrust = false;
    this.PreventVulImg = false;
    this.PreventVulImgSeverity = 'low';
    this.ScanImgOnPush = false;
  }

  initByProject(pro: Project) {
    this.Public = pro.metadata.public === 'true' ? true : false;
    this.ContentTrust = pro.metadata.enable_content_trust === 'true' ? true : false;
    this.PreventVulImg = pro.metadata.prevent_vul === 'true' ? true : false;
    if (pro.metadata.severity) { this.PreventVulImgSeverity = pro.metadata.severity; };
    this.ScanImgOnPush = pro.metadata.auto_scan === 'true' ? true : false;
  };
}

@Component({
  selector: 'hbr-project-policy-config',
  template: PROJECT_POLICY_CONFIG_TEMPLATE,
  styles: [PROJECT_POLICY_CONFIG_STYLE],
})
export class ProjectPolicyConfigComponent implements OnInit {
  onGoing = false;
  @Input() projectId: number;
  @Input() projectName = 'unknown';

  @Input() hasSignedIn: boolean;
  @Input() hasProjectAdminRole: boolean;

  @ViewChild('cfgConfirmationDialog') confirmationDlg: ConfirmationDialogComponent;

  systemInfo: SystemInfo;
  orgProjectPolicy = new ProjectPolicy();
  projectPolicy = new ProjectPolicy();

  severityOptions = [
    {severity: 'high', severityLevel: 'VULNERABILITY.SEVERITY.HIGH'},
    {severity: 'medium', severityLevel: 'VULNERABILITY.SEVERITY.MEDIUM'},
    {severity: 'low', severityLevel: 'VULNERABILITY.SEVERITY.LOW'},
    {severity: 'negligible', severityLevel: 'VULNERABILITY.SEVERITY.NEGLIGIBLE'},
  ];

  constructor(
    private errorHandler: ErrorHandler,
    private translate: TranslateService,
    private projectService: ProjectService,
    private systemInfoService: SystemInfoService,
  ) {}

  ngOnInit(): void {
    // assert if project id exist
    if (!this.projectId) {
      this.errorHandler.error('Project ID cannot be unset.');
      return;
    }

    // get system info
    toPromise<SystemInfo>(this.systemInfoService.getSystemInfo())
    .then(systemInfo => this.systemInfo = systemInfo)
    .catch(error => this.errorHandler.error(error));

    // retrive project level policy data
    this.retrieve();
  }

  public get withNotary(): boolean {
    return this.systemInfo ? this.systemInfo.with_notary : false;
  }

  public get withClair(): boolean {
    return this.systemInfo ? this.systemInfo.with_clair : false;
  }

  retrieve(state?: State): any {
    toPromise<Project>(this.projectService.getProject(this.projectId))
    .then(
      response => {
        this.orgProjectPolicy.initByProject(response);
        this.projectPolicy.initByProject(response);
      })
    .catch(error => this.errorHandler.error(error));
  }

  updateProjectPolicy(projectId: string|number, pp: ProjectPolicy) {
    this.projectService.updateProjectPolicy(projectId, pp);
  }

  refresh() {
    this.retrieve();
  }

  isValid() {
    let flag = false;
    if (!this.projectPolicy.PreventVulImg || this.severityOptions.some(x => x.severity === this.projectPolicy.PreventVulImgSeverity)) {
      flag = true;
    }
    return flag;
  }

  hasChanges() {
    return !compareValue(this.orgProjectPolicy, this.projectPolicy);
  }

  save() {
    if (!this.hasChanges()) {
      return;
    }
    this.onGoing = true;
    toPromise<any>(this.projectService.updateProjectPolicy(this.projectId, this.projectPolicy))
    .then(() => {
      this.onGoing = false;

      this.translate.get('CONFIG.SAVE_SUCCESS').subscribe((res: string) => {
        this.errorHandler.info(res);
      });
      this.refresh();
    })
    .catch(error => {
      this.onGoing = false;
      this.errorHandler.error(error);
    });
  }

  cancel(): void {
    let msg = new ConfirmationMessage(
        'CONFIG.CONFIRM_TITLE',
        'CONFIG.CONFIRM_SUMMARY',
        '',
        {},
        ConfirmationTargets.CONFIG
    );
    this.confirmationDlg.open(msg);
  }

  reset(): void {
    this.projectPolicy = clone(this.orgProjectPolicy);
  }

  confirmCancel(ack: ConfirmationAcknowledgement): void {
    if (ack && ack.source === ConfirmationTargets.CONFIG &&
        ack.state === ConfirmationState.CONFIRMED) {
        this.reset();
    }
  }
}
