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
import { Router } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';
import { Comparator } from 'clarity-angular';

import { REPOSITORY_LISTVIEW_TEMPLATE } from './repository-listview.component.html';
import { REPOSITORY_LISTVIEW_STYLE } from './repository-listview.component.css';

import {
  Repository,
  SystemInfo,
  SystemInfoService,
  RepositoryService,
  RequestQueryParams,
  RepositoryItem,
  TagService
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
import {BatchInfo, BathInfoChanges} from "../confirmation-dialog/confirmation-batch-message";
import {Observable} from "rxjs/Observable";

@Component({
  selector: 'hbr-repository-listview',
  template: REPOSITORY_LISTVIEW_TEMPLATE,
  styles: [REPOSITORY_LISTVIEW_STYLE],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class RepositoryListviewComponent implements OnChanges, OnInit {
  signedCon: {[key: string]: any | string[]} = {};
  @Input() projectId: number;
  @Input() projectName = 'unknown';
  @Input() urlPrefix: string;

  @Input() hasSignedIn: boolean;
  @Input() hasProjectAdminRole: boolean;
  @Output() tagClickEvent = new EventEmitter<TagClickEvent>();

  lastFilteredRepoName: string;
  repositories: RepositoryItem[];
  systemInfo: SystemInfo;
  selectedRow: RepositoryItem[] = [];

  loading: boolean = true;

  @ViewChild('confirmationDialog')
  confirmationDialog: ConfirmationDialogComponent;

  batchDelectionInfos: BatchInfo[] = [];
  pullCountComparator: Comparator<RepositoryItem> = new CustomComparator<RepositoryItem>('pull_count', 'number');

  tagsCountComparator: Comparator<RepositoryItem> = new CustomComparator<RepositoryItem>('tags_count', 'number');

  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage = 1;
  totalCount = 0;
  currentState: State;

  constructor(
    private errorHandler: ErrorHandler,
    private translateService: TranslateService,
    private repositoryService: RepositoryService,
    private systemInfoService: SystemInfoService,
    private tagService: TagService,
    private ref: ChangeDetectorRef,
    private router: Router) { }

  public get registryUrl(): string {
    return this.systemInfo ? this.systemInfo.registry_url : '';
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

  ngOnChanges(changes: SimpleChanges): void {
    if (changes['projectId'] && changes['projectId'].currentValue) {
      this.refresh();
    }
  }

  ngOnInit(): void {
    // Get system info for tag views
    toPromise<SystemInfo>(this.systemInfoService.getSystemInfo())
      .then(systemInfo => this.systemInfo = systemInfo)
      .catch(error => this.errorHandler.error(error));

    this.lastFilteredRepoName = '';
  }

  confirmDeletion(message: ConfirmationAcknowledgement) {
    if (message &&
        message.source === ConfirmationTargets.REPOSITORY &&
        message.state === ConfirmationState.CONFIRMED) {

      let promiseLists: any[] = [];
      let repoNames: string[] = message.data.split(',');

      repoNames.forEach(repoName => {
        promiseLists.push(this.delOperate(repoName));
      });

      Promise.all(promiseLists).then((item) => {
        this.selectedRow = [];
        this.refresh();
        let st: State = this.getStateAfterDeletion();
        if (!st) {
          this.refresh();
        } else {
          this.clrLoad(st);
        }
      });
    }
  }

  delOperate(repoName:  string) {
    let findedList = this.batchDelectionInfos.find(data => data.name === repoName);
    if (this.signedCon[repoName].length !== 0) {
      this.translateService.get('REPOSITORY.DELETION_TITLE_REPO_SIGNED').subscribe(res => {
        findedList.status = res;
      });
    } else {
      return toPromise<number>(this.repositoryService
          .deleteRepository(repoName))
          .then(
              response => {
                this.translateService.get('BATCH.DELETED_SUCCESS').subscribe(res => {
                  findedList = BathInfoChanges(findedList, res);
                });
              }).catch(error => {
            if (error.status === "412") {
              Observable.forkJoin(this.translateService.get('BATCH.DELETED_FAILURE'),
                  this.translateService.get('REPOSITORY.TAGS_SIGNED')).subscribe(res => {
                findedList = BathInfoChanges(findedList, res[0], false, true, res[1]);
              });
              return;
            }
            this.translateService.get('BATCH.DELETED_FAILURE').subscribe(res => {
              findedList = BathInfoChanges(findedList, res, false, true);
            });
          });
    }
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

  deleteRepos(repoLists: RepositoryItem[]) {
    if (repoLists && repoLists.length) {
      let repoNames: string[] = [];
      this.batchDelectionInfos = [];
      let repArr: any[] = [];

      repoLists.forEach(repo => {
        repoNames.push(repo.name);
        let initBatchMessage = new BatchInfo();
        initBatchMessage.name = repo.name;
        this.batchDelectionInfos.push(initBatchMessage);

        if (!this.signedCon[repo.name]) {
          repArr.push(this.getTagInfo(repo.name));
        }
      });

      Promise.all(repArr).then(() => {
          this.confirmationDialogSet('REPOSITORY.DELETION_TITLE_REPO',  '', repoNames.join(','), 'REPOSITORY.DELETION_SUMMARY_REPO', ConfirmationButtons.DELETE_CANCEL);
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

  signedDataSet(repoName: string): void {
    let signature = '';
    if (this.signedCon[repoName].length === 0) {
      this.confirmationDialogSet('REPOSITORY.DELETION_TITLE_REPO', signature, repoName, 'REPOSITORY.DELETION_SUMMARY_REPO', ConfirmationButtons.DELETE_CANCEL);
      return;
    }
    signature = this.signedCon[repoName].join(',');
    this.confirmationDialogSet('REPOSITORY.DELETION_TITLE_REPO_SIGNED', signature, repoName, 'REPOSITORY.DELETION_SUMMARY_REPO_SIGNED', ConfirmationButtons.CLOSE);
  }

  confirmationDialogSet(summaryTitle: string, signature: string, repoName: string, summaryKey: string,  button: ConfirmationButtons): void {
    this.translateService.get(summaryKey,
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

  selectedChange(): void {
    let hnd = setInterval(() => this.ref.markForCheck(), 100);
    setTimeout(() => clearInterval(hnd), 2000);
  }
  refresh() {
    this.doSearchRepoNames('');
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

    let totalPages: number = Math.ceil(total / this.pageSize);
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
  public gotoLink(projectId: number, repoName: string): void {
    let linkUrl = [this.router.url, repoName];
    this.router.navigate(linkUrl);
  }
}