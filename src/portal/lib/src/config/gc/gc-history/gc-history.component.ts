import { Component, OnInit } from '@angular/core';
import { GcRepoService } from "../gc.service";
import { GcJobViewModel } from "../gcLog";
import { GcViewModelFactory } from "../gc.viewmodel.factory";

@Component({
  selector: 'gc-history',
  templateUrl: './gc-history.component.html',
  styleUrls: ['./gc-history.component.scss']
})
export class GcHistoryComponent implements OnInit {
  jobs: Array<GcJobViewModel> = [];
  constructor(
    private gcRepoService: GcRepoService,
    private gcViewModelFactory: GcViewModelFactory,
    ) {}

  ngOnInit() {
    this.getJobs();
  }

  getJobs() {
    this.gcRepoService.getJobs().subscribe(jobs => {
      this.jobs = this.gcViewModelFactory.createJobViewModel(jobs);
    });
  }

  getLogLink(id): string {
    return this.gcRepoService.getLogLink(id);
  }

}
