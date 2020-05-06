import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SelectArtifactIconPipe } from './select-artifact-icon/select-artifact-icon.pipe';



@NgModule({
  declarations: [SelectArtifactIconPipe],
  imports: [
    CommonModule
  ],
  exports: [
    SelectArtifactIconPipe
  ]
})
export class AllPipesModule { }
