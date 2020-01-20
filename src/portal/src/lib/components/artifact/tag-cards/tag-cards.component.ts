import { Component, OnInit, Input, ChangeDetectorRef } from '@angular/core';
import { forkJoin, Subject } from 'rxjs';
import { LabelService } from "../../../services/label.service";
import { LabelState } from '../artifact-list-tab.component';
import { clone } from '../../../utils/utils';
import { ErrorHandler } from '../../../utils/error-handler';
import { Label, TagService, Tag } from '../../../services';
import { TagUi } from './tag';
import { debounceTime, distinctUntilChanged } from 'rxjs/operators';

@Component({
  selector: 'tag-cards',
  templateUrl: './tag-cards.component.html',
  styleUrls: ['./tag-cards.component.scss']
})
export class TagCardsComponent implements OnInit {
  @Input() tags: TagUi[] = [];
  @Input() repositoryName: string;
  @Input() projectId: number;
  imageLabels: LabelState[] = [];
  stickLabelNameFilter: Subject<TagUi> = new Subject<TagUi>();

  constructor(
    private errorHandler: ErrorHandler,
    private tagService: TagService,
    private ref: ChangeDetectorRef,
    public labelService: LabelService
  ) { }

  ngOnInit() {
    this.getAllLabels();
    this.stickLabelNameFilter
      .pipe(debounceTime(500))
      // .pipe(distinctUntilChanged())
      .subscribe((currentTag) => {
        if (currentTag.labelFilterName.length) {
          // this.filterOnGoing = true;

          currentTag.showLabels.forEach(data => {
            if (data.label.name.indexOf(currentTag.labelFilterName) !== -1) {
              data.show = true;
            } else {
              data.show = false;
            }
          });
        }
      });
  }
  getAllLabels(): void {
    forkJoin(this.labelService.getGLabels(), this.labelService.getPLabels(this.projectId)).subscribe(results => {
      let allLabels = results[0].concat(results[1]);
      // results.forEach(labels => {
      this.tags.forEach(item => {
        allLabels.forEach(data => {
          if (item.labels.some(label => label.id === data.id)) {
            item.showLabels.push({ 'iconsShow': true, 'label': data, 'show': true });
          } else {
            item.showLabels.push({ 'iconsShow': false, 'label': data, 'show': true });
          }

        });
      });
    }, error => this.errorHandler.error(error));
  }
  stickLabel(labelInfo: LabelState, currentTag: TagUi): void {
    if (labelInfo && !labelInfo.iconsShow) {
      this.selectLabel(labelInfo, currentTag);
    }
    if (labelInfo && labelInfo.iconsShow) {
      this.unSelectLabel(labelInfo, currentTag);
    }
  }
  // inprogress = false;
  selectLabel(labelInfo: LabelState, currentTag: TagUi): void {
    // if (!this.inprogress) {
    //   this.inprogress = true;
    let labelId = labelInfo.label.id;
    // this.selectedRow = this.selectedTag;
    this.tagService.addLabelToImages(this.repositoryName, currentTag.name, labelId).subscribe(res => {
      labelInfo.iconsShow = true;
      this.sortOperation(currentTag, labelInfo);

      // // set the selected label in front
      // currentTag.splice(currentTag.indexOf(labelInfo), 1);
      // currentTag.some((data, i) => {
      //   if (!data.iconsShow) {
      //     currentTag.splice(i, 0, labelInfo);
      //     return true;
      //   }
      // });

      // // when is the last one
      // if (this.imageStickLabels.every(data => data.iconsShow === true)) {
      //   this.imageStickLabels.push(labelInfo);
      // }

      // labelInfo.iconsShow = true;
      // this.inprogress = false;
    }, err => {
      // this.inprogress = false;
      this.errorHandler.error(err);
    });
    // }
  }

  unSelectLabel(labelInfo: LabelState, currentTag: TagUi): void {
    // if (!this.inprogress) {
    //   this.inprogress = true;
    let labelId = labelInfo.label.id;
    // this.selectedRow = this.selectedTag;
    this.tagService.deleteLabelToImages(this.repositoryName, currentTag.name, labelId).subscribe(res => {

      // // insert the unselected label to groups with the same icons
      labelInfo.iconsShow = false;
      this.sortOperation(currentTag, labelInfo);
      // this.inprogress = false;
    }, err => {
      // this.inprogress = false;
      this.errorHandler.error(err);
    });
    // }
  }
  // insert the unselected label to groups with the same icons
  sortOperation(tag: TagUi, labelInfo: LabelState): void {
    tag.labels = [];
    tag.showLabels.forEach((data, i) => {
      if (data.iconsShow) {
        tag.labels.push(data.label);
      }
    });
  }
  handleStickInputFilter(currenTag: TagUi) {
    if (currenTag.labelFilterName.length) {
      this.stickLabelNameFilter.next(currenTag);
    } else {
      currenTag.showLabels.every(data => data.show = true);
    }
  }
}
