import { Type } from '@angular/core';
import { LabelComponent} from "./label.component";
import { LabelSignPostComponent } from './label-signpost/label-signpost.component';

export const LABEL_DIRECTIVES: Type<any>[] = [
  LabelComponent,
  LabelSignPostComponent,
];
