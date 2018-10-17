import { Type } from '@angular/core';
import { LabelComponent} from "./label.component";
import { LabelMarkerComponent } from './label-marker/label-marker.component';

export const LABEL_DIRECTIVES: Type<any>[] = [
  LabelComponent,
  LabelMarkerComponent
];
