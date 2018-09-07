import { Type } from "@angular/core";
import { PushImageButtonComponent } from './push-image.component';
import { CopyInputComponent } from './copy-input.component';

export * from "./push-image.component";
export * from './copy-input.component';

export const PUSH_IMAGE_BUTTON_DIRECTIVES: Type<any>[] = [
    CopyInputComponent,
    PushImageButtonComponent
];
