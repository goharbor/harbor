import { Component, Input, Output, EventEmitter } from '@angular/core';

import { ReplicationService } from '../replication.service';
import { Policy } from '../policy';

@Component({
  selector: 'list-policy',
  templateUrl: 'list-policy.component.html'
})
export class ListPolicyComponent {
  
  @Input() policies: Policy[];
  @Output() selectOne = new EventEmitter<number>();

  constructor(private replicationService: ReplicationService){}

  selectPolicy(policy: Policy): void {
    console.log('Select policy ID:' + policy.id);
    this.selectOne.emit(policy.id);
  }
}