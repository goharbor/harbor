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
import { Component, OnInit, ViewChild, Input, Output, EventEmitter, OnDestroy } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { ArtifactClickEvent, RepositoryItem, State, SystemInfo, SystemInfoService } from "../../../../../lib/services";
import {
  ConfirmationAcknowledgement,
  ConfirmationDialogComponent,
  ConfirmationMessage
} from "../../../../../lib/components/confirmation-dialog";
import { ErrorHandler } from "../../../../../lib/utils/error-handler";
import { RepositoryService } from "../../repository.service";
import { ArtifactService } from "../../artifact/artifact.service";
import { ConfirmationState, ConfirmationTargets } from "../../../../../lib/entities/shared.const";
import { ActivatedRoute } from "@angular/router";
import { Project } from '../../../project';
import {
  RepositoryService as NewRepositoryService
} from "../../../../../../ng-swagger-gen/services/repository.service";
const TabLinkContentMap: { [index: string]: string } = {
  'repo-info': 'info',
  'repo-image': 'image'
};

@Component({
  selector: 'artifact-list',
  templateUrl: './artifact-list.component.html',
  styleUrls: ['./artifact-list.component.scss']
})
export class ArtifactListComponent implements OnInit {
  signedCon: { [key: string]: any | string[] } = {};
  @Input() projectId: number;
  @Input() memberRoleID: number;
  @Input() repoName: string;
  @Input() hasSignedIn: boolean;
  @Input() hasProjectAdminRole: boolean;
  @Input() isGuest: boolean;
  @Output() tagClickEvent = new EventEmitter<ArtifactClickEvent>();
  @Output() backEvt: EventEmitter<any> = new EventEmitter<any>();
  @Output() putArtifactReferenceArr: EventEmitter<string[]> = new EventEmitter<[]>();

  onGoing = false;
  editing = false;
  inProgress = true;
  currentTabID = 'repo-image';
  changedRepositories: RepositoryItem[];
  systemInfo: SystemInfo;

  imageInfo: string;
  orgImageInfo: string;

  timerHandler: any;
  projectName: string = '';
  @ViewChild('confirmationDialog', { static: false })
  confirmationDlg: ConfirmationDialogComponent;
  showCurrentTitle: string;
  artifactDigest: string;
  constructor(
    private errorHandler: ErrorHandler,
    private repositoryService: RepositoryService,
    private systemInfoService: SystemInfoService,
    private artifactService: ArtifactService,
    private newRepositoryService: NewRepositoryService,
    private translate: TranslateService,
    private activatedRoute: ActivatedRoute,
  ) {
    this.activatedRoute.params.subscribe(params => {
      let depth = this.activatedRoute.snapshot.params['depth'];
      if (depth) {
        const arr: string[] = depth.split('-');
        this.artifactDigest = depth.split('-')[arr.length - 1];
      }
    });
  }

  public get registryUrl(): string {
    return this.systemInfo ? this.systemInfo.registry_url : '';
  }

  public get withNotary(): boolean {
    return this.systemInfo ? this.systemInfo.with_notary : false;
  }
  public get withAdmiral(): boolean {
    return this.systemInfo ? this.systemInfo.with_admiral : false;
  }

  ngOnInit(): void {
    if (!this.projectId) {
      this.errorHandler.error('Project ID cannot be unset.');
      return;
    }
    const resolverData = this.activatedRoute.snapshot.data;
    if (resolverData) {
      const pro: Project = <Project>resolverData['projectResolver'];
      this.projectName = pro.name;
    }
    this.showCurrentTitle = this.repoName || 'null';
    this.retrieve();
    this.inProgress = false;
    this.artifactService.TriggerArtifactChan$.subscribe(res => {
      if (res === 'repoName') {
        this.showCurrentTitle = this.repoName;
      } else {
        this.showCurrentTitle = res[res.length - 1];
      }
    });
  }

  retrieve(state?: State) {
    this.repositoryService.getRepositories(this.projectId, this.repoName)
      .subscribe(response => {
        if (response.metadata.xTotalCount > 0) {
          this.orgImageInfo = response.data[0].description;
          this.imageInfo = response.data[0].description;
        }
      }, error => this.errorHandler.error(error));
    this.systemInfoService.getSystemInfo()
      .subscribe(systemInfo => this.systemInfo = systemInfo, error => this.errorHandler.error(error));
  }
  refresh() {
    this.retrieve();
  }
  isCurrentTabLink(tabID: string): boolean {
    return this.currentTabID === tabID;
  }

  isCurrentTabContent(ContentID: string): boolean {
    return TabLinkContentMap[this.currentTabID] === ContentID;
  }

  tabLinkClick(tabID: string) {
    this.currentTabID = tabID;
  }

  goBack(): void {
    this.backEvt.emit(this.projectId);
  }

  hasChanges() {
    return this.imageInfo !== this.orgImageInfo;
  }

  reset(): void {
    this.imageInfo = this.orgImageInfo;
  }

  hasInfo() {
    return this.imageInfo && this.imageInfo.length > 0;
  }

  editInfo() {
    this.editing = true;
  }

  saveInfo() {
    if (!this.hasChanges()) {
      return;
    }
    this.onGoing = true;
    let params: NewRepositoryService.UpdateRepositoryParams = {
      repositoryName: this.repoName,
      repository: {description: this.imageInfo},
      projectName: this.projectName,
    };
    this.newRepositoryService.updateRepository(params)
      .subscribe(() => {
        this.onGoing = false;
        this.translate.get('CONFIG.SAVE_SUCCESS').subscribe((res: string) => {
          this.errorHandler.info(res);
        });
        this.editing = false;
        this.refresh();
      }, error => {
        this.onGoing = false;
        this.errorHandler.error(error);
      });
  }

  cancelInfo() {
    let msg = new ConfirmationMessage(
      'CONFIG.CONFIRM_TITLE',
      'CONFIG.CONFIRM_SUMMARY',
      '',
      {},
      ConfirmationTargets.CONFIG
    );
    this.confirmationDlg.open(msg);
  }

  confirmCancel(ack: ConfirmationAcknowledgement): void {
    this.editing = false;
    if (ack && ack.source === ConfirmationTargets.CONFIG &&
      ack.state === ConfirmationState.CONFIRMED) {
      this.reset();
    }
  }
}
