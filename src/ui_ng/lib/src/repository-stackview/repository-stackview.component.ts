import {
  Component,
  Input,
  Output,
  OnInit,
  ViewChild,
  ChangeDetectionStrategy,
  ChangeDetectorRef,
  EventEmitter
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

@Component({
  selector: 'hbr-repository-stackview',
  template: REPOSITORY_STACKVIEW_TEMPLATE,
  styles: [REPOSITORY_STACKVIEW_STYLES],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class RepositoryStackviewComponent implements OnInit {

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
        }).catch(error => this.errorHandler.error(error));
    }
  }

  ngOnInit(): void {
    if (!this.projectId) {
      this.errorHandler.error('Project ID cannot be unset.');
      return;
    }
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

  deleteRepo(repoName: string) {
    let message = new ConfirmationMessage(
      'REPOSITORY.DELETION_TITLE_REPO',
      'REPOSITORY.DELETION_SUMMARY_REPO',
      repoName,
      repoName,
      ConfirmationTargets.REPOSITORY,
      ConfirmationButtons.DELETE_CANCEL);
    this.confirmationDialog.open(message);
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