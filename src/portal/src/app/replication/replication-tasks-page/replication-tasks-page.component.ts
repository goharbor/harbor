import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
@Component({
  selector: 'app-replication-tasks-page',
  templateUrl: './replication-tasks-page.component.html',
  styleUrls: ['./replication-tasks-page.component.scss']
})
export class ReplicationTasksPageComponent implements OnInit {
  executionId: string;
  constructor(
    private route: ActivatedRoute,
  ) { }

  ngOnInit(): void {
    this.executionId = this.route.snapshot.params["id"];
  }

}
