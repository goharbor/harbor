import { NgModule } from '@angular/core';
import { RepositoryComponent } from './repository.component';
import { SharedModule } from '../shared.module';

@NgModule({
  imports: [ SharedModule ],
  declarations: [ RepositoryComponent ],
  exports: [ RepositoryComponent ] 
})
export class RepositoryModule {}