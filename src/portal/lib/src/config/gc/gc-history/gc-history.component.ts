import { Component, OnInit } from '@angular/core';
import { GcRepoService } from "../gc.service";
import { GcJobViewModel } from "../gcLog";
import { GcViewModelFactory } from "../gc.viewmodel.factory";
import { ErrorHandler } from "../../../error-handler/index";

@Component({
  selector: 'gc-history',
  templateUrl: './gc-history.component.html',
  styleUrls: ['./gc-history.component.scss']
})
export class GcHistoryComponent implements OnInit {
  jobs: Array<GcJobViewModel> = [];
  loading: boolean;
  constructor(
    private gcRepoService: GcRepoService,
    private gcViewModelFactory: GcViewModelFactory,
    private errorHandler: ErrorHandler
    ) {}

  ngOnInit() {
    this.getJobs();
  }

  getJobs() {
    this.loading = true;
    this.gcRepoService.getJobs().subscribe(jobs => {
      this.jobs = this.gcViewModelFactory.createJobViewModel(jobs);
      this.loading = false;
    }, error => {
        this.errorHandler.error(error);
        this.loading = false;
    });
  }

  getLogLink(id): string {
    return this.gcRepoService.getLogLink(id);
  }

}
