import { Component, Input, Output, EventEmitter, ViewChild, SimpleChanges } from '@angular/core';
import { Configuration } from '../config';
import { Quota, State, Comparator, ClrDatagridComparatorInterface, QuotaHardLimitInterface } from '../../service/interface';
import { clone, isEmpty, getChanges, getSuitableUnit, calculatePage, CustomComparator, SeparationNumberCharacter } from '../../utils';
import { ErrorHandler } from '../../error-handler/index';
import { QuotaUnits } from '../../shared/shared.const';
import { EditProjectQuotasComponent } from './edit-project-quotas/edit-project-quotas.component';
import {
  ConfigurationService
} from '../../service/index';
import { TranslateService } from '@ngx-translate/core';
import { forkJoin } from 'rxjs';
import { QuotaService } from "../../service/quota.service";
import { Router } from '@angular/router';
import { finalize } from 'rxjs/operators';
const quotaSort = {
  count: 'used.count',
  storage: "used.storage",
  sortType: 'string'
}
@Component({
  selector: 'project-quotas',
  templateUrl: './project-quotas.component.html',
  styleUrls: ['./project-quotas.component.scss']
})
export class ProjectQuotasComponent {

  config: Configuration = new Configuration();
  @ViewChild('editProjectQuotas')
  editQuotaDialog: EditProjectQuotasComponent;
  loading = true;
  quotaHardLimitValue: QuotaHardLimitInterface;
  currentState: State;

  @Output() configChange: EventEmitter<Configuration> = new EventEmitter<Configuration>();
  @Output() refreshAllconfig: EventEmitter<Configuration> = new EventEmitter<Configuration>();
  quotaList: Quota[] = [];
  originalConfig: Configuration;
  currentPage = 1;
  totalCount = 0;
  pageSize = 15;
  @Input()
  get allConfig(): Configuration {
    return this.config;
  }
  set allConfig(cfg: Configuration) {
    this.config = cfg;
    this.configChange.emit(this.config);
  }
  countComparator: Comparator<Quota> = new CustomComparator<Quota>(quotaSort.count, quotaSort.sortType);
  storageComparator: Comparator<Quota> = new CustomComparator<Quota>(quotaSort.storage, quotaSort.sortType);

  constructor(
    private configService: ConfigurationService,
    private quotaService: QuotaService,
    private translate: TranslateService,
    private router: Router,
    private errorHandler: ErrorHandler) { }

  editQuota(quotaHardLimitValue: QuotaHardLimitInterface) {
    const defaultTexts = [this.translate.get('QUOTA.EDIT_PROJECT_QUOTAS'), this.translate.get('QUOTA.SET_QUOTAS')
      , this.translate.get('QUOTA.COUNT_QUOTA'), this.translate.get('QUOTA.STORAGE_QUOTA')]
    forkJoin(...defaultTexts).subscribe(res => {
      const defaultTextsObj = {
        editQuota: res[0],
        setQuota: res[1],
        countQuota: res[2],
        storageQuota: res[3],
        quotaHardLimitValue: quotaHardLimitValue,
        isSystemDefaultQuota: false
      };
      this.editQuotaDialog.openEditQuotaModal(defaultTextsObj);
    })
  }

  editDefaultQuota(quotaHardLimitValue: QuotaHardLimitInterface) {
    const defaultTexts = [this.translate.get('QUOTA.EDIT_DEFAULT_PROJECT_QUOTAS'), this.translate.get('QUOTA.SET_DEFAULT_QUOTAS')
      , this.translate.get('QUOTA.COUNT_DEFAULT_QUOTA'), this.translate.get('QUOTA.STORAGE_DEFAULT_QUOTA')]
    forkJoin(...defaultTexts).subscribe(res => {
      const defaultTextsObj = {
        editQuota: res[0],
        setQuota: res[1],
        countQuota: res[2],
        storageQuota: res[3],
        quotaHardLimitValue: quotaHardLimitValue,
        isSystemDefaultQuota: true
      };
      this.editQuotaDialog.openEditQuotaModal(defaultTextsObj);

    })
  }
  public getChanges() {
    let allChanges = getChanges(this.originalConfig, this.config);
    if (allChanges) {
      return this.getQuotaChanges(allChanges);
    }
    return null;
  }

  getQuotaChanges(allChanges) {
    let changes = {};
    for (let prop in allChanges) {
      if (prop === 'storage_per_project'
        || prop === 'count_per_project'
      ) {
        changes[prop] = allChanges[prop];
      }
    }
    return changes;
  }

  public saveConfig(configQuota): void {
    this.allConfig.count_per_project.value = configQuota.count;
    this.allConfig.storage_per_project.value = configQuota.storage === '-1' ? configQuota.storage : `${configQuota.storage}${configQuota.storageUnit}`;
    let changes = this.getChanges();
    if (!isEmpty(changes)) {
      this.loading = true;
      this.configService.saveConfigurations(changes)
        .pipe(finalize(() => {
          this.loading = false;
          this.editQuotaDialog.openEditQuota = false;
        }))
        .subscribe(response => {
          this.refreshAllconfig.emit();
          this.errorHandler.info('CONFIG.SAVE_SUCCESS');
        }
          , error => {
            this.errorHandler.error(error);
          });
    } else {
      // Inprop situation, should not come here
      this.translate.get('CONFIG.NO_CHANGE').subscribe(res => {
        this.editQuotaDialog.inlineAlert.showInlineError(res);
      })
    }
  }

  confirmEdit(event) {
    if (event.isSystemDefaultQuota) {
      this.saveConfig(event.formValue);
    }
    else {
      this.saveCurrentQuota(event);
    }
  }
  saveCurrentQuota(event) {
    let count = event.formValue.count;
    let storage = event.formValue.storage === '-1' ? event.formValue.storage : `${event.formValue.storage}${event.formValue.storageUnit}`;
    this.loading = true;
    this.quotaService.updateQuota(event.id, { count, storage }).subscribe(res => {
      this.editQuotaDialog.openEditQuota = false;
      this.getQuotaList(this.currentState);
      this.errorHandler.info('QUOTA.SAVE_SUCCESS');
    }, error => {
      this.errorHandler.error(error);
      this.loading = false;
    })
  }

  getquotaHardLimitValue() {
    const storageNumberAndUnit = this.allConfig.storage_per_project ? this.allConfig.storage_per_project.value : '';
    const storageLimit = SeparationNumberCharacter(storageNumberAndUnit, QuotaUnits[3].UNIT).numberStr;
    const storageUnit = SeparationNumberCharacter(storageNumberAndUnit, QuotaUnits[3].UNIT).character;
    const countLimit = this.allConfig.count_per_project ? this.allConfig.count_per_project.value : '';
    this.quotaHardLimitValue = { storageLimit, storageUnit, countLimit };
  }
  getQuotaList(state: State) {
    // Keep state for future filtering and sorting
    this.currentState = state;

    let pageNumber: number = calculatePage(state);
    if (pageNumber <= 0) { pageNumber = 1; }
    let sortBy: any = '';
    if (state.sort) {
      sortBy = state.sort.by as string | ClrDatagridComparatorInterface<any>;
      sortBy = sortBy.fieldName ? sortBy.fieldName : sortBy;
      sortBy = state.sort.reverse ? `-${sortBy}` : sortBy
    }
    this.loading = true;

    this.quotaService.getQuotaList(pageNumber, this.pageSize, sortBy).pipe(finalize=> {
      this.loading = false;
    }).subscribe(res => {
      if (res.headers) {
        let xHeader: string = res.headers.get("X-Total-Count");
        if (xHeader) {
          this.totalCount = parseInt(xHeader, 0);
        }
      }
      // Prevent Page Errors Caused by Server Failure
      this.quotaList = res.body.filter(quota => {
        return quota.ref !== null;
      });
    }, error => {
      this.errorHandler.error(error);
    })
  }
  ngOnChanges(changes: SimpleChanges): void {
    if (changes && changes["allConfig"]) {
      this.originalConfig = clone(this.config);
      this.getquotaHardLimitValue();
    }
  }
  getSuitableUnit(value) {
    const QuotaUnitsCopy = clone(QuotaUnits);
    return getSuitableUnit(value, QuotaUnitsCopy)
  }

  goToLink(proId) {
    let linkUrl = ["harbor", "projects", proId, "summary"];
    this.router.navigate(linkUrl);
  }
  refresh() {
    const state: State = {
      page: {
        from: 0,
        to: 14,
        size: 15
      },
    }
    this.getQuotaList(state);
  }

}
