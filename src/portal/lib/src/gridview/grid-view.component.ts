/*
 * Copyright (c) 2017 VMware, Inc. All Rights Reserved.
 *
 * This product is licensed to you under the Apache License, Version 2.0 (the "License").
 * You may not use this product except in compliance with the License.
 *
 * This product may include a number of subcomponents with separate copyright notices
 * and license terms. Your use of these subcomponents is subject to the terms and
 * conditions of the subcomponent's license, as noted in the LICENSE file.
 */

import {
  Component,
  Input,
  Output,
  ContentChild,
  ViewChild,
  ViewChildren,
  TemplateRef,
  HostListener,
  ViewEncapsulation,
  EventEmitter,
  AfterViewInit
} from "@angular/core";
import { Subscription } from "rxjs";
import { TranslateService } from "@ngx-translate/core";

import { ScrollPosition } from "../service/interface";

@Component({
  selector: "hbr-gridview",
  templateUrl: "./grid-view.component.html",
  styleUrls: ["./grid-view.component.scss"],
  encapsulation: ViewEncapsulation.None
})
/**
 * Grid view general component.
 */
export class GridViewComponent implements AfterViewInit {
  @Input() loading: boolean;
  @Input() totalCount: number;
  @Input() currentPage: number;
  @Input() pageSize: number;
  @Input() expectScrollPercent = 70;
  @Input() withAdmiral: boolean;
  @Input()
  set items(value: any[]) {
    let newCardStyles = value.map((d, index) => {
      if (index < this.cardStyles.length) {
        return this.cardStyles[index];
      }
      return {
        opacity: "0",
        overflow: "hidden"
      };
    });
    this.cardStyles = newCardStyles;
    this._items = value;
  }

  @Output() loadNextPageEvent = new EventEmitter<any>();

  @ViewChildren("cardItem") cards: any;
  @ViewChild("itemsHolder", {static: false}) itemsHolder: any;
  @ContentChild(TemplateRef, {static: false}) gridItemTmpl: any;

  _items: any[] = [];

  cardStyles: any = [];
  itemsHolderStyle: any = {};
  layoutTimeout: any;

  querySub: Subscription;
  routerSub: Subscription;

  totalItemsCount: number;
  loadedPages = 0;
  nextPageLink: string;
  hidePartialRows = false;
  loadPagesTimeout: any;

  CurrentScrollPosition: ScrollPosition = {
    sH: 0,
    sT: 0,
    cH: 0
  };

  preScrollPosition: ScrollPosition = null;

  constructor(private translate: TranslateService) {}

  ngAfterViewInit() {
    this.cards.changes.subscribe(() => {
      this.throttleLayout();
    });
    this.throttleLayout();
  }

  get items() {
    return this._items;
  }

  @HostListener("scroll", ["$event"])
  onScroll(event: any) {
    this.preScrollPosition = this.CurrentScrollPosition;
    this.CurrentScrollPosition = {
      sH: event.target.scrollHeight,
      sT: event.target.scrollTop,
      cH: event.target.clientHeight
    };
    if (
      !this.loading &&
      this.isScrollDown() &&
      this.isScrollExpectPercent() &&
      this.currentPage * this.pageSize < this.totalCount
    ) {
      this.loadNextPageEvent.emit();
    }
  }

  isScrollDown(): boolean {
    return this.preScrollPosition.sT < this.CurrentScrollPosition.sT;
  }

  isScrollExpectPercent(): boolean {
    return (
      (this.CurrentScrollPosition.sT + this.CurrentScrollPosition.cH) /
        this.CurrentScrollPosition.sH >
      this.expectScrollPercent / 100
    );
  }

  @HostListener("window:resize", ["$event"])
  onResize(event: any) {
    this.throttleLayout();
  }

  throttleLayout() {
    clearTimeout(this.layoutTimeout);
    this.layoutTimeout = setTimeout(() => {
      this.layout.call(this);
    }, 40);
  }

  get isFirstPage() {
    return this.currentPage <= 1;
  }

  layout() {
    let el = this.itemsHolder.nativeElement;

    let width = el.offsetWidth;
    let items = el.querySelectorAll(".card-item");
    let items_count = items.length;
    if (items_count === 0) {
      el.height = 0;
      return;
    }

    let itemsHeight = [];
    for (let i = 0; i < items_count; i++) {
      itemsHeight[i] = items[i].offsetHeight;
    }

    let height = Math.max.apply(null, itemsHeight);
    let itemsStyle: CSSStyleDeclaration = window.getComputedStyle(items[0]);

    let minWidthStyle: string = itemsStyle.minWidth;
    let maxWidthStyle: string = itemsStyle.maxWidth;

    let minWidth = parseInt(minWidthStyle, 10);
    let maxWidth = parseInt(maxWidthStyle, 10);

    let marginHeight: number =
      parseInt(itemsStyle.marginTop, 10) +
      parseInt(itemsStyle.marginBottom, 10);
    let marginWidth: number =
      parseInt(itemsStyle.marginLeft, 10) +
      parseInt(itemsStyle.marginRight, 10);

    let columns = Math.floor(width / (minWidth + marginWidth));

    let columnsToUse = Math.max(Math.min(columns, items_count), 1);
    let rows = Math.floor(items_count / columnsToUse);
    let itemWidth = Math.min(
      Math.floor(width / columnsToUse) - marginWidth,
      maxWidth
    );
    let itemSpacing =
      columnsToUse === 1 || columns > items_count
        ? marginWidth
        : (width - marginWidth - columnsToUse * itemWidth) / (columnsToUse - 1);
    if (!this.withAdmiral) {
      // Fixed spacing and margin on standalone mode
      itemSpacing = marginWidth;
      itemWidth = minWidth;
    }

    let visible = items_count;
    if (
      this.hidePartialRows &&
      this.totalItemsCount &&
      items_count !== this.totalItemsCount
    ) {
      visible = rows * columnsToUse;
    }

    let count = 0;
    for (let i = 0; i < visible; i++) {
      let item = items[i];
      let itemStyle = window.getComputedStyle(item);

      let left = (i % columnsToUse) * (itemWidth + itemSpacing);
      let top = Math.floor(count / columnsToUse) * (height + marginHeight);

      // trick to show nice apear animation, where the item is already positioned,
      // but it will pop out
      let oldTransform = itemStyle.transform;
      if (!oldTransform || oldTransform === "none") {
        this.cardStyles[i] = {
          transform: "translate(" + left + "px," + top + "px) scale(0)",
          width: itemWidth + "px",
          transition: "none",
          overflow: "hidden"
        };
        this.throttleLayout();
      } else {
        this.cardStyles[i] = {
          transform: "translate(" + left + "px," + top + "px) scale(1)",
          width: itemWidth + "px",
          transition: null,
          overflow: "hidden"
        };
        this.throttleLayout();
      }

      if (!item.classList.contains("context-selected")) {
        let itemHeight = itemsHeight[i];
        if (itemStyle.display === "none" && itemHeight !== 0) {
          this.cardStyles[i].display = null;
        }
        if (itemHeight !== 0) {
          count++;
        }
      }
    }

    for (let i = visible; i < items_count; i++) {
      this.cardStyles[i] = {
        display: "none"
      };
    }
    this.itemsHolderStyle = {
      height: Math.ceil(count / columnsToUse) * (height + marginHeight) + "px"
    };
  }

  onCardEnter(i: number) {
    this.cardStyles[i].overflow = "visible";
  }

  onCardLeave(i: number) {
    this.cardStyles[i].overflow = "hidden";
  }

  trackByFn(index: number, item: any) {
    return index;
  }
}
