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
import {
      Component,
      Input,
      AfterViewInit,
      ViewChild,
      ElementRef
} from '@angular/core';

import * as $ from 'jquery';

const RESOURCE_COLOR_GREEN_NORMAL: string = '#5DB700';
const RESOURCE_COLOR_ORANGE_NORMAL: string = '#FBBF00';
const RESOURCE_COLOR_RED_NORMAL: string = '#EA400D';
const RESOURCE_COLOR_GREY500: string = '#D7DEE2';
const RESOURCE_COLOR_GREY600: string = '#C7D1D6';

/**
 * Guage to visualize percent usage.
 */
@Component({
      selector: 'esxc-gauge',
      templateUrl: 'gauge.component.html',
      styleUrls: ['gauge.component.css']
})

export class GaugeComponent implements AfterViewInit {
      private _backgroundColor: string;
      private _colorOne: string;
      private _colorTwo: string;
      private _size: string = "small"; //Support small, medium, large
      private _title: string = "UNKNOWN"; //Lang key
      private _free: number = 0;
      private _threasHold: number = 0;

      /**
       * Background color of the component. Default is white.
       */
      @Input()
      get backgroundColor() {
            if (this._backgroundColor) {
                  return this._backgroundColor;
            }
            return '#FAFAFA';
      }

      set backgroundColor(value: string) {
            this._backgroundColor = value;
      }

      private _positionOne: number;
      /**
       * Keep these two properties
       * Percentage of the total width for the first portion of the bar.
       * Bar one is rendered above bar two, so bar two's position should always
       * be greater than bar one if you want bar two to be visible.
       */
      @Input()
      get positionOne(): number {
            return this._positionOne;
      }

      set positionOne(value: number) {
            this._positionOne = value;
            this.setBars();
      }

      private _positionTwo: number;
      /**
       * Percentage of the total width for the second portion of the bar
       */
      @Input()
      get positionTwo(): number {
            return this._positionTwo;
      }

      set positionTwo(value: number) {
            this._positionTwo = this._positionOne + value;
            this.setBars();
      }

      private _animate: boolean;
      /**
       * Whether to animate transitions in the bars
       */
      @Input()
      get animate(): boolean {
            return this._animate;
      }

      set animate(value: boolean) {
            if (typeof value !== 'undefined') {
                  this._animate = value;
            }
            this.setAnimate();
      }

      //Define the gauge size
      @Input()
      get size(): string {
            return this._size;
      }

      set size(sz: string) {
            if (typeof sz !== 'undefined') {
                  if (sz === 'small' || sz === 'medium' || sz === 'large') {
                        this._size = sz;
                        return;
                  }
            }

            this._size = "small";
      }

      get sizeClass(): string {
            return "esxc-gauge-" + this._size;
      }

      @Input()
      get title(): string {
            return this._title;
      }

      set title(t: string) {
            if (typeof t !== 'undefined') {
                  this._title = t;
            }
      }

      @Input()
      get free(): number {
            return this._free;
      }

      set free(u: number) {
            this._free = u;
            this.determineColors();
      }

      get used(): number {
            return this._threasHold - this._free;
      }

      @Input()
      get threasHold(): number {
            return this._threasHold;
      }

      set threasHold(th: number) {
            this._threasHold = th;
            this.determineColors();
      }

      ngAfterViewInit() {
            this.determineColors();
      }

      @ViewChild('barOne') private barOne: ElementRef;
      @ViewChild('barTwo') private barTwo: ElementRef;

      private determineColors() {
            let percent: number = 0;
            if (this._threasHold !== 0) {
                  let used: number = this._threasHold - this._free;
                  if (used < 0) {
                        used = 0;
                  }
                  percent = (used / this._threasHold) * 100;
            }

            while (percent > 100) {
                  percent = percent - 100;
            }

            if (percent <= 70) {
                  this._colorOne = RESOURCE_COLOR_GREEN_NORMAL;
            } else if (percent > 70 && percent <= 90) {
                  this._colorOne = RESOURCE_COLOR_ORANGE_NORMAL;
            } else if (percent > 90 && percent <= 100) {
                  this._colorOne = RESOURCE_COLOR_RED_NORMAL;
            } else {
                  this._colorOne = RESOURCE_COLOR_GREY600;
            }

            this._positionOne = percent;
            this.setBars();
            this.setColors();
            this.setAnimate();
      }

      private setBars() {
            if (!this.barOne || !this.barTwo) {
                  return;
            }

            let barOne = $(this.barOne.nativeElement);
            let barTwo = $(this.barTwo.nativeElement);

            if (!barOne || !barTwo) {
                  return;
            }

            let posOne, posTwo;

            if (isNaN(this.positionOne)) {
                  posOne = posTwo = 0;
            } else {
                  posOne = (this.positionOne / 100) * 180;
                  posTwo = (this.positionTwo / 100) * 180;
            }

            barOne.css({
                  '-webkit-transform': 'rotate(' + posOne + 'deg)',
                  '-moz-transform': 'rotate(' + posOne + 'deg)',
                  '-ms-transform': 'rotate(' + posOne + 'deg)',
                  '-o-transform': 'rotate(' + posOne + 'deg)',
                  'transform': 'rotate(' + posOne + 'deg)'
            });

            barTwo.css({
                  '-webkit-transform': 'rotate(' + posTwo + 'deg)',
                  '-moz-transform': 'rotate(' + posTwo + 'deg)',
                  '-ms-transform': 'rotate(' + posTwo + 'deg)',
                  '-o-transform': 'rotate(' + posTwo + 'deg)',
                  'transform': 'rotate(' + posTwo + 'deg)'
            });
      }

      private setColors() {
            if (!this.barOne || !this.barTwo) {
                  return;
            }

            let barOne = $(this.barOne.nativeElement);
            let barTwo = $(this.barTwo.nativeElement);

            if (!barOne || !barTwo) {
                  return;
            }

            barOne.css({
                  'background-color': this._colorOne
            });

            barTwo.css({
                  'background-color': this._colorTwo
            });
      }

      private setAnimate() {
            if (!this.barOne || !this.barTwo) {
                  return;
            }

            let barOne = $(this.barOne.nativeElement);
            let barTwo = $(this.barTwo.nativeElement);

            if (!barOne || !barTwo) {
                  return;
            }

            let transition = 'transform 1s ease';
            let prefixes = ['webkit', 'moz', 'ms', 'o'];
            let css = {
                  'transition': transition
            };

            if (!this._animate) {
                  transition = 'none';
            };

            for (let prefix of prefixes) {
                  css['-' + prefix + '-transition'] = transition;
            }

            barOne.css(css);
            barTwo.css(css);
      }

}