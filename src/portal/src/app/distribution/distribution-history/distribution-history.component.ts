import { Component, OnInit, OnDestroy } from '@angular/core';
import { DistributionHistory, QueryParam } from '../distribution-interface';
import { DistributionService } from '../distribution.service';
import { DEFAULT_PAGE_SIZE } from '../../../lib/utils/utils';

@Component({
  selector: 'app-distribution-history',
  templateUrl: './distribution-history.component.html',
  styleUrls: ['./distribution-history.component.scss']
})
export class DistributionHistoryComponent implements OnInit, OnDestroy {
  loading: boolean = false;
  records: DistributionHistory[] = [];
  pageSize: number = DEFAULT_PAGE_SIZE;
  currentPage: number = 1;
  totalCount: number = 0;
  queryParam: QueryParam = new QueryParam();

  constructor(private disService: DistributionService) {}

  ngOnInit() {
    this.queryParam.pageSize = this.pageSize;
    this.loadData();
  }

  ngOnDestroy(): void {}

  loadData() {
    this.loading = true;
    this.queryParam.page = this.currentPage;
    this.disService.getDistributionHistories(this.queryParam).subscribe(
      response => {
        this.totalCount = Number.parseInt(
          response.headers.get('x-total-count')
        );
        this.records = response.body;
      },
      err => console.error(err)
    );
    this.loading = false;
  }

  refresh() {
    this.currentPage = 1;
    this.loadData();
  }

  doFilter($evt: any) {
    this.currentPage = 1;
    this.queryParam.query = $evt;
    this.loadData();
  }
}
