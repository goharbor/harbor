import {
  Component,
  Input,
  Output,
  OnInit,
  ViewChild,
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  EventEmitter, OnChanges, SimpleChanges
} from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { Comparator } from 'clarity-angular';

import { REPOSITORY_STACKVIEW_TEMPLATE } from './repository-stackview.component.html';
import { REPOSITORY_STACKVIEW_STYLES } from './repository-stackview.component.css';

import {
  Repository,
  SystemInfo,
  SystemInfoService,
  RepositoryService,
  RequestQueryParams,
  RepositoryItem
} from '../service/index';
import { ErrorHandler } from '../error-handler/error-handler';

import { toPromise, CustomComparator } from '../utils';

import { ConfirmationState, ConfirmationTargets, ConfirmationButtons } from '../shared/shared.const';

import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ConfirmationMessage } from '../confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../confirmation-dialog/confirmation-state-message';
import { Subscription } from 'rxjs/Subscription';
import { Tag, TagClickEvent } from '../service/interface';

import { State } from "clarity-angular";
import {
  DEFAULT_PAGE_SIZE,
  calculatePage,
  doFiltering,
  doSorting
} from '../utils';
import {TagService} from '../service/index';

@Component({
  selector: 'hbr-repository-stackview',
  template: REPOSITORY_STACKVIEW_TEMPLATE,
  styles: [REPOSITORY_STACKVIEW_STYLES],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class RepositoryStackviewComponent implements OnChanges, OnInit {
  signedCon: {[key: string]: any | string[]} = {};

  @Input() projectId: number;
  @Input() projectName: string = "unknown";

  @Input() hasSignedIn: boolean;
  @Input() hasProjectAdminRole: boolean;
  @Output() tagClickEvent = new EventEmitter<TagClickEvent>();

  lastFilteredRepoName: string;
  repositories: RepositoryItem[];
  systemInfo: SystemInfo;

  loading: boolean = true;

  @ViewChild('confirmationDialog')
  confirmationDialog: ConfirmationDialogComponent;

  pullCountComparator: Comparator<Repository> = new CustomComparator<Repository>('pull_count', 'number');

  tagsCountComparator: Comparator<Repository> = new CustomComparator<Repository>('tags_count', 'number');

  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage: number = 1;
  totalCount: number = 0;
  currentState: State;

  constructor(
    private errorHandler: ErrorHandler,
    private translateService: TranslateService,
    private repositoryService: RepositoryService,
    private systemInfoService: SystemInfoService,
    private translate: TranslateService,
    private tagService: TagService,
    private ref: ChangeDetectorRef) { }

  public get registryUrl(): string {
    return this.systemInfo ? this.systemInfo.registry_url : "";
  }

  public get withNotary(): boolean {
    return this.systemInfo ? this.systemInfo.with_notary : false;
  }

  public get withClair(): boolean {
    return this.systemInfo ? this.systemInfo.with_clair : false;
  }

  public get isClairDBReady(): boolean {
    return this.systemInfo &&
      this.systemInfo.clair_vulnerability_status &&
      this.systemInfo.clair_vulnerability_status.overall_last_update > 0;
  }

  public get showDBStatusWarning(): boolean {
    return this.withClair && !this.isClairDBReady;
  }

  confirmDeletion(message: ConfirmationAcknowledgement) {
    if (message &&
      message.source === ConfirmationTargets.REPOSITORY &&
      message.state === ConfirmationState.CONFIRMED) {
      let repoName = message.data;
      toPromise<number>(this.repositoryService
        .deleteRepository(repoName))
        .then(
        response => {
          this.refresh();
          let st: State = this.getStateAfterDeletion();
          if (!st) {
            this.refresh();
          } else {
            this.clrLoad(st);
          }
          this.translateService.get('REPOSITORY.DELETED_REPO_SUCCESS')
            .subscribe(res => this.errorHandler.info(res));
        }).catch(error => {
          if (error.status === "412"){
            this.translateService.get('REPOSITORY.TAGS_SIGNED')
                .subscribe(res => this.errorHandler.info(res));
            return;
          }
          this.errorHandler.error(error);
        });
    }
  }

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['projectId']) {
      this.refresh();
    }
  }

  ngOnInit(): void {
    //Get system info for tag views
    toPromise<SystemInfo>(this.systemInfoService.getSystemInfo())
      .then(systemInfo => this.systemInfo = systemInfo)
      .catch(error => this.errorHandler.error(error));

    this.lastFilteredRepoName = '';
  }

  doSearchRepoNames(repoName: string) {
    this.lastFilteredRepoName = repoName;
    this.currentPage = 1;

    let st: State = this.currentState;
    if (!st) {
      st = { page: {} };
    }
    st.page.size = this.pageSize;
    st.page.from = 0;
    st.page.to = this.pageSize - 1;
    this.clrLoad(st);
  }

  saveSignatures(event: {[key: string]: string[]}): void {
    Object.assign(this.signedCon, event);
  }

  deleteRepo(repoName: string) {
    // get children tags data

    let signature: string = '';
    if (this.signedCon[repoName]) {
      if (this.signedCon[repoName].length === 0) {
        this.confirmationDialogSet('DELETION_TITLE_REPO', signature, repoName, 'REPOSITORY.DELETION_SUMMARY_REPO', ConfirmationButtons.DELETE_CANCEL);
        return;
      }
      signature = this.signedCon[repoName].join(',');
      this.confirmationDialogSet('DELETION_TITLE_REPO_SIGNED', signature, repoName, 'REPOSITORY.DELETION_SUMMARY_REPO_SIGNED', ConfirmationButtons.CLOSE);
    } else {
      this.getTagInfo(repoName).then(() => {
        if (this.signedCon[repoName].length) {
          signature = this.signedCon[repoName].join(',');
          this.confirmationDialogSet('DELETION_TITLE_REPO_SIGNED', signature, repoName, 'REPOSITORY.DELETION_SUMMARY_REPO_SIGNED', ConfirmationButtons.CLOSE);
        } else {
          this.confirmationDialogSet('DELETION_TITLE_REPO', signature, repoName, 'REPOSITORY.DELETION_SUMMARY_REPO', ConfirmationButtons.DELETE_CANCEL);
        }
      });
    }
  }
  getTagInfo(repoName: string): Promise<void> {
     // this.signedNameArr = [];
    this.signedCon[repoName] = [];
     return toPromise<Tag[]>(this.tagService
            .getTags(repoName))
            .then(items => {
              items.forEach((t: Tag) => {
                if (t.signature !== null) {
                  this.signedCon[repoName].push(t.name);
                }
              });
            })
            .catch(error => this.errorHandler.error(error));
  }

  confirmationDialogSet(summaryTitle: string, signature: string, repoName: string, summaryKey: string,  button: ConfirmationButtons): void {
    this.translate.get(summaryKey,
        {
          repoName: repoName,
          signedImages: signature,
        })
        .subscribe((res: string) => {
          summaryKey = res;
          let message = new ConfirmationMessage(
              summaryTitle,
              summaryKey,
              repoName,
              repoName,
              ConfirmationTargets.REPOSITORY,
              button);
          this.confirmationDialog.open(message);

          let hnd = setInterval(() => this.ref.markForCheck(), 100);
          setTimeout(() => clearInterval(hnd), 5000);
    });
  }

  refresh() {
    this.doSearchRepoNames("");
  }

  watchTagClickEvt(tagClickEvt: TagClickEvent): void {
    this.tagClickEvent.emit(tagClickEvt);
  }

  clrLoad(state: State): void {
    //Keep it for future filtering and sorting
    this.currentState = state;

    let pageNumber: number = calculatePage(state);
    if (pageNumber <= 0) { pageNumber = 1; }

    //Pagination
    let params: RequestQueryParams = new RequestQueryParams();
    params.set("page", '' + pageNumber);
    params.set("page_size", '' + this.pageSize);

    this.loading = true;

    toPromise<Repository>(this.repositoryService.getRepositories(
      this.projectId,
      this.lastFilteredRepoName,
      params))
      .then((repo: Repository) => {
        this.totalCount = repo.metadata.xTotalCount;
        this.repositories = repo.data;

        this.signedCon = {};
        //Do filtering and sorting
        this.repositories = doFiltering<RepositoryItem>(this.repositories, state);
        this.repositories = doSorting<RepositoryItem>(this.repositories, state);

        this.loading = false;
      })
      .catch(error => {
        this.loading = false;
        this.errorHandler.error(error);
      });

    //Force refresh view
    let hnd = setInterval(() => this.ref.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 5000);
  }

  getStateAfterDeletion(): State {
    let total: number = this.totalCount - 1;
    if (total <= 0) { return null; }

    let totalPages: number = Math.floor(total / this.pageSize);
    let targetPageNumber: number = this.currentPage;

    if (this.currentPage > totalPages) {
      targetPageNumber = totalPages;//Should == currentPage -1
    }

    let st: State = this.currentState;
    if (!st) {
      st = { page: {} };
    }
    st.page.size = this.pageSize;
    st.page.from = (targetPageNumber - 1) * this.pageSize;
    st.page.to = targetPageNumber * this.pageSize - 1;

    return st;
  }
}

