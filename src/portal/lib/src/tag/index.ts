import { Type } from "@angular/core";
import { TagComponent } from "./tag.component";
import { TagDetailComponent } from "./tag-detail.component";
import { TagHistoryComponent } from "./tag-history.component";

export * from "./tag.component";
export * from "./tag-detail.component";
export * from "./tag-history.component";

export const TAG_DIRECTIVES: Type<any>[] = [
  TagComponent,
  TagDetailComponent,
  TagHistoryComponent
];
