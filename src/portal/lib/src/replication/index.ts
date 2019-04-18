import { Type } from '@angular/core';
import { ReplicationComponent } from './replication.component';
import { ReplicationTasksComponent } from './replication-tasks/replication-tasks.component';

export * from './replication.component';
export * from './replication-tasks/replication-tasks.component';

export const REPLICATION_DIRECTIVES: Type<any>[] = [
  ReplicationComponent,
  ReplicationTasksComponent
];
