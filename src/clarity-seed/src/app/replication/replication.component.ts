import { Component, OnInit } from '@angular/core';
import { Policy } from './policy';
import { Job } from './job';

@Component({
  selector: 'replicaton',
  templateUrl: 'replication.component.html'
})
export class ReplicationComponent implements OnInit {
   policies: Policy[];
   jobs: Job[];

   ngOnInit(): void {
     this.policies = [
       { name: 'sync_01', status: 'Disabled', destination: '10.117.5.135', lastStartTime: '2016-12-21 17:52:35', description: 'test'},
       { name: 'sync_02', status: 'Enabled', destination: '10.117.5.117', lastStartTime: '2016-12-21 12:22:47', description: 'test'},
     ];
     this.jobs = [
       { name: 'project01/ubuntu:14.04', status: 'Finished', operation: 'Transfer', creationTime: '2016-12-21 17:53:50', endTime: '2016-12-21 17:55:01'},
       { name: 'project01/mysql:5.6', status: 'Finished', operation: 'Transfer', creationTime: '2016-12-21 17:54:20', endTime: '2016-12-21 17:55:05'},
       { name: 'project01/photon:latest', status: 'Finished', operation: 'Transfer', creationTime: '2016-12-21 17:54:50', endTime: '2016-12-21 17:55:15'}
     ];
   }
}