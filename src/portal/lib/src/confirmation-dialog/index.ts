import { Type } from "@angular/core";

import { ConfirmationDialogComponent } from "./confirmation-dialog.component";

export * from "./confirmation-dialog.component";
export * from "./confirmation-batch-message";
export * from "./confirmation-message";
export * from "./confirmation-state-message";

export const CONFIRMATION_DIALOG_DIRECTIVES: Type<any>[] = [
  ConfirmationDialogComponent
];
