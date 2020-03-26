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
import {Component, Input, OnInit, OnChanges} from '@angular/core';




import {Label} from "../../services/interface";
import {LabelColor} from "../../entities/shared.const";


@Component({
    selector: 'hbr-label-piece',
    templateUrl: './label-piece.component.html',
    styleUrls: ['./label-piece.component.scss']
})

export class LabelPieceComponent implements OnInit, OnChanges {
    @Input() label: Label;
    @Input() labelWidth: number;
    labelColor: {[key: string]: string};

    ngOnChanges(): void {
        if (this.label) {
            let color = this.label.color;
            if (!color) {color = '#FFFFFF'; }
            this.labelColor = LabelColor.find(data => data.color === color);
        }
    }

    ngOnInit(): void { }
}
