import { Type } from '@angular/core';
import { LabelComponent} from "./label.component";
import { LabelMarkerComponent } from './label-marker/label-marker.component';
import { LabelSignPostComponent } from './label-signpost/label-signpost.component';
import { LabelFilterComponent } from './label-filter/label-filter.component';

export const LABEL_DIRECTIVES: Type<any>[] = [
  LabelComponent,
  LabelMarkerComponent,
  LabelSignPostComponent,
  LabelFilterComponent
];
