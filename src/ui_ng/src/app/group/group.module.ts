import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { SharedModule } from '../shared/shared.module';
import { GroupComponent } from './group.component';
import { AddGroupModalComponent } from './add-group-modal/add-group-modal.component';
import { GroupService } from './group.service';


@NgModule({
  imports: [
    CommonModule,
    SharedModule,
    FormsModule,
    ReactiveFormsModule
  ],
  exports: [
    GroupComponent,
    AddGroupModalComponent,
    FormsModule,
    ReactiveFormsModule
  ],
  providers: [ GroupService ],
  declarations: [GroupComponent, AddGroupModalComponent]
})
export class GroupModule { }
