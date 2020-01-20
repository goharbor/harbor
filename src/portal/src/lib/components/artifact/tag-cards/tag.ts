import { Tag } from "../../../services";
import { LabelState } from "../artifact-list-tab.component";

export interface TagUi extends Tag {
    showLabels: LabelState[];
    labelFilterName: "";
}
