import { Component } from '@angular/core';
import { DEFAULT_PAGE_SIZE } from '../../../lib/utils/utils';
import { finalize } from "rxjs/operators";
import { ErrorHandler } from "../../../lib/utils/error-handler";
import { PreheatService } from "../../../../ng-swagger-gen/services/preheat.service";
import { PreheatHistory } from "../../../../ng-swagger-gen/models/preheat-history";

@Component({
  selector: 'app-distribution-history',
  templateUrl: './distribution-history.component.html',
  styleUrls: ['./distribution-history.component.scss']
})
export class DistributionHistoryComponent  {
  loading: boolean = true;
  records: PreheatHistory[] = [];
  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage: number = 1;
  totalCount: number = 0;
  queryString: string;
  isOpenFilterTag: boolean = false;
  defaultFilter: string = "image";

  constructor(private disService: PreheatService,
              private errorHandler: ErrorHandler) {}
  loadData() {
    const queryParam: PreheatService.ListPreheatHistoriesParams = {
      page: this.currentPage,
      pageSize: this.pageSize
    };
    if (this.queryString) {
      queryParam.q = encodeURIComponent(`${this.defaultFilter}=~${this.queryString}`);
    }
    this.loading = true;
    this.disService.ListPreheatHistoriesResponse(queryParam)
      .pipe(finalize(() => this.loading = false))
      .subscribe(
      response => {
        this.totalCount = Number.parseInt(
          response.headers.get('x-total-count')
        );
        this.records = response.body;
      },
      err => {
        this.errorHandler.error(err);
      }
    );
  }

  refresh() {
    this.currentPage = 1;
    this.loadData();
  }

  doFilter($evt: any) {
    this.currentPage = 1;
    this.queryString = $evt;
    this.loadData();
  }

  openFilter(isOpen: boolean): void {
    this.isOpenFilterTag = isOpen;
  }

  selectFilterKey($event: any): void {
    this.defaultFilter = $event['target'].value;
    this.loadData();
  }
}
