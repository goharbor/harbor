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
import { Component, OnInit, ViewChild, Input, Output, EventEmitter } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { State } from '../service/interface';

import { RepositoryService } from '../service/repository.service';
import { Repository, RepositoryItem, Tag, TagClickEvent,
  SystemInfo, SystemInfoService, TagService } from '../service/index';
import { ErrorHandler } from '../error-handler/index';
import { ConfirmationState, ConfirmationTargets } from '../shared/shared.const';
import { ConfirmationDialogComponent, ConfirmationMessage, ConfirmationAcknowledgement } from '../confirmation-dialog/index';
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
export class RepositoryComponent implements OnInit {
  signedCon: {[key: string]: any | string[]} = {};
  @Input() projectId: number;
  @Input() memberRoleID: number;
  @Input() repoName: string;
  @Input() hasSignedIn: boolean;
  @Input() hasProjectAdminRole: boolean;
  @Input() isGuest: boolean;
  @Output() tagClickEvent = new EventEmitter<TagClickEvent>();
  @Output() backEvt: EventEmitter<any> = new EventEmitter<any>();

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

  constructor(
    private errorHandler: ErrorHandler,
    private repositoryService: RepositoryService,
    private systemInfoService: SystemInfoService,
    private tagService: TagService,
    private translate: TranslateService,
  ) {  }

  public get registryUrl(): string {
    return this.systemInfo ? this.systemInfo.registry_url : '';
  }

  public get withNotary(): boolean {
    return this.systemInfo ? this.systemInfo.with_notary : false;
  }

  public get withClair(): boolean {
    return this.systemInfo ? this.systemInfo.with_clair : false;
  }

  public get withAdmiral(): boolean {
    return this.systemInfo ? this.systemInfo.with_admiral : false;
  }

  ngOnInit(): void {
    if (!this.projectId) {
      this.errorHandler.error('Project ID cannot be unset.');
      return;
    }
    this.retrieve();
    this.inProgress = false;
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

  watchTagClickEvt(tagClickEvt: TagClickEvent): void {
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
             items.forEach((t: Tag) => {
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
}
