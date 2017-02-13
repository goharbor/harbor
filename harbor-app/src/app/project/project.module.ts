import { NgModule } from '@angular/core';
import { ProjectComponent } from './project.component';
import { ProjectListComponent } from './project-list.component';
import { SharedModule } from '../shared.module';
import { RouterModule } from '@angular/router';

@NgModule({
  imports: [ 
    SharedModule,
    RouterModule
  ],
  declarations: [ 
    ProjectComponent,
    ProjectListComponent
  ],
  exports: [ ProjectComponent ]
})
export class ProjectModule {}