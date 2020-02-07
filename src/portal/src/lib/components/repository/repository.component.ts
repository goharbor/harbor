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
import { State } from '../../services/interface';

import { RepositoryService } from '../../services/repository.service';
import { Repository, RepositoryItem, Tag, ArtifactClickEvent,
  SystemInfo, SystemInfoService, TagService, ArtifactService } from '../../services';
import { ErrorHandler } from '../../utils/error-handler';
import { ConfirmationState, ConfirmationTargets } from '../../entities/shared.const';
import { ConfirmationDialogComponent, ConfirmationMessage, ConfirmationAcknowledgement } from '../confirmation-dialog';
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
const TabLinkContentMap: {[index: string]: string} = {
  'repo-info': 'info',
  'repo-image': 'image'
};

@Component({
  selector: 'hbr-repository',
  templateUrl: './repository.component.html',
  styleUrls: ['./repository.component.scss']
})
export class RepositoryComponent implements OnInit, OnDestroy {
  signedCon: {[key: string]: any | string[]} = {};
  @Input() projectId: number;
  @Input() memberRoleID: number;
  @Input() repoName: string;
  // @Input() referArtifactName: string;
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

  @ViewChild('confirmationDialog', {static: false})
  confirmationDlg: ConfirmationDialogComponent;
  showCurrentTitle: string ;
  constructor(
    private errorHandler: ErrorHandler,
    private repositoryService: RepositoryService,
    private systemInfoService: SystemInfoService,
    private tagService: TagService,
    private artifactService: ArtifactService,
    private translate: TranslateService,
  ) {  }

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
    // let repoName = this.repoName.split("sha256:")[0];
    // this.referArtifactName = `${this.repoName.split(repoName)[1]}` ? `${this.repoName.split(repoName)[1]}` : "";
    // this.repoName = this.repoName.split('/')[1];
    // this.repoName = this.repoName.split(":sha256:")[0];
    if (!this.projectId) {
      this.errorHandler.error('Project ID cannot be unset.');
      return;
    }
    this.showCurrentTitle = this.repoName || 'null';
    this.retrieve();
    this.inProgress = false;
    this.artifactService.TriggerArtifactChan$.subscribe(res => {
      if (res === 'repoName') {
        this.showCurrentTitle = this.repoName;
      }
      else {
        this.showCurrentTitle = res[res.length-1]
      }
    })
  }

  retrieve(state?: State) {
    this.repositoryService.getRepositories(this.projectId, this.repoName)
      .subscribe( response => {
        if (response.metadata.xTotalCount > 0) {
          this.orgImageInfo = response.data[0].description;
          this.imageInfo = response.data[0].description;
        }
      }, error => this.errorHandler.error(error));
    this.systemInfoService.getSystemInfo()
      .subscribe(systemInfo => this.systemInfo = systemInfo, error => this.errorHandler.error(error));
  }

  saveSignatures(event: {[key: string]: string[]}): void {
    Object.assign(this.signedCon, event);
  }

  refresh() {
    this.retrieve();
  }

  watchTagClickEvt(tagClickEvt: ArtifactClickEvent): void {
    this.tagClickEvent.emit(tagClickEvt);
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

  getTagInfo(repoName: string): Observable<void> {
    // this.signedNameArr = [];
   this.signedCon[repoName] = [];
    return this.tagService
           .getTags(repoName)
           .pipe(map(items => {
             items.forEach((t: any) => {
               if (t.signature !== null) {
                 this.signedCon[repoName].push(t.name);
               }
             });
           })
           , catchError(error => observableThrowError(error)));
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
    this.repositoryService.updateRepositoryDescription(this.repoName, this.imageInfo)
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

  ngOnDestroy(): void {
    this.artifactService.reference = [];
  }
  putReferArtifactArray(referArtifactArray) {
    if (referArtifactArray.length) {
      this.showCurrentTitle = referArtifactArray[referArtifactArray.length - 1];
      // referArtifactArray.pop();
      this.putArtifactReferenceArr.emit(referArtifactArray);
    }

}

}
