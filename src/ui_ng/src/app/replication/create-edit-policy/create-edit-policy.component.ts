import { Component, Input } from '@angular/core';

import { ReplicationService } from '../replication.service';

@Component({
  selector: 'create-edit-policy',
  templateUrl: 'create-edit-policy.component.html'
})
export class CreateEditPolicyComponent {

  createEditPolicyOpened: boolean;

  constructor(private replicationService: ReplicationService) {}
  
  openCreateEditPolicy(): void {
    console.log('createEditPolicyOpened:' + this.createEditPolicyOpened);
    this.createEditPolicyOpened = true;
  } 
}