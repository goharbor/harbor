// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { Component, Input, Output, OnInit, EventEmitter } from "@angular/core";
import { Subject } from "rxjs";
import { debounceTime } from 'rxjs/operators';

@Component({
  selector: "hbr-filter",
  templateUrl: "./filter.component.html",
  styleUrls: ["./filter.component.scss"]
})
export class FilterComponent implements OnInit {
  placeHolder: string = "";
  filterTerms = new Subject<string>();
  isExpanded: boolean = false;

  @Output() private filterEvt = new EventEmitter<string>();
  @Output() private openFlag = new EventEmitter<boolean>();
  @Input() readonly: string = null;
  @Input() currentValue: string;
  @Input("filterPlaceholder")
  public set flPlaceholder(placeHolder: string) {
    this.placeHolder = placeHolder;
  }
  @Input() expandMode: boolean = false;
  @Input() withDivider: boolean = false;

  ngOnInit(): void {
    this.filterTerms
      .pipe(debounceTime(500))
      .subscribe(terms => {
        this.filterEvt.emit(terms);
      });
  }

  valueChange(): void {
    // Send out filter terms
    this.filterTerms.next(this.currentValue && this.currentValue.trim());
  }

  inputFocus(): void {
    this.openFlag.emit(this.isExpanded);
  }

  onClick(): void {
    // Only enabled when expandMode is set to false
    if (this.expandMode) {
      return;
    }
    this.isExpanded = !this.isExpanded;
    this.openFlag.emit(this.isExpanded);
  }

  public get isShowSearchBox(): boolean {
    return this.expandMode || (!this.expandMode && this.isExpanded);
  }
}
